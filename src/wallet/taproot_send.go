package wallet

import (
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
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/m25lab/bitcoin_nft/src/layer1"
	"github.com/m25lab/bitcoin_nft/src/layer2"
	"github.com/m25lab/bitcoin_nft/src/taproot"
	"google.golang.org/grpc"
)

// See the difference: SendCoin, SendInvoice, SendMessage, SendTransaction
func SendCoinWithTaproot(taprootKeyFamily int32, internalKey *secp256k1.PublicKey) {
	lndConn, err := layer2.GetLndGrpcSetup()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := walletrpc.NewWalletKitClient(lndConn)

	taprootKey, err := taproot.GenerateTaprootKey(client, taprootKeyFamily, internalKey, []byte("foobar"))
	if err != nil {
		fmt.Println(err)
		return
	}

	outpoint, _, err := SendToTaprootOutput(taprootKey, 1000, lndConn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s:%d\n", outpoint.Hash.String(), outpoint.Index)
}

func SendToTaprootOutput(taprootKey *btcec.PublicKey, amount int64, lndConn *grpc.ClientConn) (*wire.OutPoint, []byte, error) {
	tapScriptAddr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(taprootKey), &chaincfg.TestNet3Params,
	)
	if err != nil {
		return nil, nil, err
	}
	fmt.Printf("Destination address is %s\n", tapScriptAddr)

	p2trPkScript, err := txscript.PayToAddrScript(tapScriptAddr)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}
	fmt.Printf("Transaction id is %s\n", res.Txid)

	client, err := layer1.GetBitcoinRPCClient()
	if err != nil {
		return nil, nil, err
	}
	hash, err := chainhash.NewHashFromStr(res.Txid)
	if err != nil {
		return nil, nil, err
	}
	p2trOutputIndex := layer1.GetOutputIndex(hash, tapScriptAddr.String(), client)
	p2trOutpoint := wire.OutPoint{
		Hash:  *hash,
		Index: uint32(p2trOutputIndex),
	}

	return &p2trOutpoint, p2trPkScript, nil

}
