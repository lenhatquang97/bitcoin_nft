package wallet

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/m25lab/bitcoin_nft/src/layer1"
	"github.com/m25lab/bitcoin_nft/src/layer2"
	"google.golang.org/grpc"
)

// See the difference: SendCoin, SendInvoice, SendMessage, SendTransaction
func HashLockScript(preimage []byte) txscript.TapLeaf {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddFullData(btcutil.Hash160(preimage))
	builder.AddOp(txscript.OP_EQUALVERIFY)
	script1, _ := builder.Script()
	return txscript.NewBaseTapLeaf(script1)
}
func ScriptSchnorrSig(pubkey *btcec.PublicKey) txscript.TapLeaf {
	builder := txscript.NewScriptBuilder()
	builder.AddFullData(schnorr.SerializePubKey(pubkey))
	builder.AddOp(txscript.OP_CHECKSIG)
	script2, _ := builder.Script()
	return txscript.NewBaseTapLeaf(script2)
}

func SendCoinWithTaproot(taprootKeyFamily int32, internalKey *secp256k1.PublicKey) {
	lndConn, err := layer2.GetLndGrpcSetup()
	if err != nil {
		fmt.Println(err)
		return
	}
	client := walletrpc.NewWalletKitClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := walletrpc.KeyReq{KeyFamily: taprootKeyFamily}
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

	p2trOutpoint, _, finalTx, err := SendToTaprootOutput(taprootKey, 1000, lndConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s:%d\n", p2trOutpoint.Hash.String(), p2trOutpoint.Index)

	var buf bytes.Buffer
	modifiedTx := finalTx.MsgTx()
	modifiedTx.TxIn[0].Witness = append(modifiedTx.TxIn[0].Witness, []byte("Hello World"))
	modifiedTx.Serialize(&buf)
	txReq := &walletrpc.Transaction{
		TxHex: buf.Bytes(),
	}

	_, err = client.PublishTransaction(ctx, txReq)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func SendToTaprootOutput(taprootKey *btcec.PublicKey, amount int64, lndConn *grpc.ClientConn) (*wire.OutPoint, []byte, *btcutil.Tx, error) {
	tapScriptAddr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(taprootKey), &chaincfg.TestNet3Params,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("Destination address is %s\n", tapScriptAddr)

	p2trPkScript, err := txscript.PayToAddrScript(tapScriptAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	lncli := lnrpc.NewLightningClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &lnrpc.SendCoinsRequest{
		Addr:   tapScriptAddr.String(),
		Amount: amount,
	}
	res, err := lncli.SendCoins(ctx, req)
	if err != nil {
		return nil, nil, nil, err
	}
	fmt.Printf("Transaction id is %s\n", res.Txid)

	client, err := layer1.GetBitcoinRPCClient()
	if err != nil {
		return nil, nil, nil, err
	}
	hash, err := chainhash.NewHashFromStr(res.Txid)
	if err != nil {
		return nil, nil, nil, err
	}
	p2trOutputIndex := layer1.GetOutputIndex(hash, tapScriptAddr.String(), client)
	p2trOutpoint := wire.OutPoint{
		Hash:  *hash,
		Index: uint32(p2trOutputIndex),
	}
	finalTx, err := client.GetRawTransaction(hash)
	if err != nil {
		return nil, nil, nil, err
	}

	return &p2trOutpoint, p2trPkScript, finalTx, nil

}
