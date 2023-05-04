package wallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
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

	_, _, _, err = SendToTaprootOutput(taprootKey, 500, lndConn)
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

	_, err = txscript.PayToAddrScript(tapScriptAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wallet := walletrpc.NewWalletKitClient(lndConn)

	fundPsbtReq := walletrpc.FundPsbtRequest{
		Template: &walletrpc.FundPsbtRequest_Raw{
			Raw: &walletrpc.TxTemplate{
				Outputs: map[string]uint64{
					tapScriptAddr.String(): uint64(amount),
				},
			},
		},
		Fees: &walletrpc.FundPsbtRequest_SatPerVbyte{
			SatPerVbyte: 1,
		},
	}
	fundRes, err := wallet.FundPsbt(ctx, &fundPsbtReq)
	if err != nil {
		return nil, nil, nil, err
	}

	packet, err := psbt.NewFromRawBytes(bytes.NewReader(fundRes.FundedPsbt), false)
	if err != nil {
		return nil, nil, nil, err
	}

	var byteArray []byte
	byteArray = append(byteArray, txscript.OP_FALSE)
	byteArray = append(byteArray, "Hello World"...)
	packet.UnsignedTx.TxIn[0].Witness = append(packet.UnsignedTx.TxIn[0].Witness, byteArray)

	var buf bytes.Buffer
	err = SerializePacket(packet, &buf)
	if err != nil {
		return nil, nil, nil, err
	}

	fmt.Println(hex.Dump(buf.Bytes()))

	//BIG PROBLEM IN finalizePsbt: Cannot serialize with witness
	_, err = wallet.FinalizePsbt(ctx, &walletrpc.FinalizePsbtRequest{
		FundedPsbt: buf.Bytes(),
	})

	if err != nil {
		return nil, nil, nil, err
	}

	// _, err = wallet.PublishTransaction(ctx, &walletrpc.Transaction{
	// 	TxHex: response.RawFinalTx,
	// })

	// if err != nil {
	// 	return nil, nil, nil, err
	// }

	// fmt.Println(hex.Dump(buf.Bytes()))
	return nil, nil, nil, nil
}
