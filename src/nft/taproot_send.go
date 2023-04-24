package nft

import (
	"context"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/keychain"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/lightningnetwork/lnd/lntest/wait"
	"github.com/m25lab/bitcoin_nft/src/layer1"
	"google.golang.org/grpc"
)

func HashLockScript(preimage []byte) txscript.TapLeaf {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(btcutil.Hash160(preimage))
	builder.AddOp(txscript.OP_EQUALVERIFY)
	script1, _ := builder.Script()
	return txscript.NewBaseTapLeaf(script1)
}
func ScriptSchnorrSig(pubkey *btcec.PublicKey) txscript.TapLeaf {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubkey))
	builder.AddOp(txscript.OP_CHECKSIG)
	script2, _ := builder.Script()
	return txscript.NewBaseTapLeaf(script2)
}

func TaprootSignOutputRawScriptSpend(trKeyFamily int32, keyRing keychain.KeyRing, internalKey *secp256k1.PublicKey) {
	lndConn, err := GetLndGrpcSetup()
	if err != nil {
		fmt.Println(err)
		return
	}
	client := walletrpc.NewWalletKitClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := walletrpc.KeyReq{KeyFamily: trKeyFamily}
	keyDesc, err := client.DeriveNextKey(ctx, &req)
	if err != nil {
		fmt.Println(err)
		return
	}
	leafSigningKey, _ := btcec.ParsePubKey(keyDesc.RawKeyBytes)
	leaf1 := HashLockScript([]byte("foobar"))
	leaf2 := ScriptSchnorrSig(leafSigningKey)

	inclusionProof := leaf1.TapHash()
	tapscript := input.TapscriptPartialReveal(internalKey, leaf2, inclusionProof[:])
	taprootKey, _ := tapscript.TaprootKey()

	outpoint, _, err := SendToTaprootOutput(taprootKey, 1000, lndConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(outpoint.Hash.String() + ":" + string(outpoint.Index))
}

func SendToTaprootOutput(taprootKey *btcec.PublicKey, amount int64, lndConn *grpc.ClientConn) (*wire.OutPoint, []byte, error) {
	tapScriptAddr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(taprootKey), &chaincfg.TestNet3Params,
	)
	if err != nil {
		return nil, nil, err
	}
	p2trPkScript, err := txscript.PayToAddrScript(tapScriptAddr)
	if err != nil {
		return nil, nil, err
	}

	// Send some coins to the generated tapscript address.
	req := &lnrpc.SendCoinsRequest{
		Addr:   tapScriptAddr.String(),
		Amount: amount,
	}

	lncli := lnrpc.NewLightningClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lncli.SendCoins(ctx, req)

	// Wait until the TX is found in the mempool.
	client, err := layer1.GetBitcoinRPCClient()
	if err != nil {
		return nil, nil, err
	}

	txid, err := AssertNumTxsInMempool(1, client)
	if err != nil {
		return nil, nil, err
	}

	p2trOutputIndex := GetOutputIndex(txid[0], tapScriptAddr.String(), client)
	p2trOutpoint := wire.OutPoint{
		Hash:  *txid[0],
		Index: uint32(p2trOutputIndex),
	}

	return &p2trOutpoint, p2trPkScript, nil

}

func GetOutputIndex(txid *chainhash.Hash, addr string, client *rpcclient.Client) int {
	// We'll then extract the raw transaction from the mempool in order to
	// determine the index of the p2tr output.
	tx, _ := client.GetRawTransaction(txid)

	p2trOutputIndex := -1
	for i, txOut := range tx.MsgTx().TxOut {
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(
			txOut.PkScript, &chaincfg.TestNet3Params,
		)

		if addrs[0].String() == addr {
			p2trOutputIndex = i
		}
	}
	return p2trOutputIndex
}

func AssertNumTxsInMempool(n int, client *rpcclient.Client) ([]*chainhash.Hash, error) {
	var (
		mem []*chainhash.Hash
	)

	err := wait.NoError(func() error {
		mem, _ = client.GetRawMempool()
		if len(mem) == n {
			return nil
		}

		return fmt.Errorf("want %v, got %v in mempool: %v",
			n, len(mem), mem)
	}, wait.MinerMempoolTimeout)

	if err != nil {
		return nil, err
	}
	return mem, nil
}
