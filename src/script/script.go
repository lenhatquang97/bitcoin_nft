package script

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

func HashLockScriptLeaf(preimage []byte) txscript.TapLeaf {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_DUP)
	builder.AddOp(txscript.OP_HASH160)
	builder.AddFullData(btcutil.Hash160(preimage))
	builder.AddOp(txscript.OP_EQUALVERIFY)
	script1, _ := builder.Script()
	return txscript.NewBaseTapLeaf(script1)
}
func SchnorrSigScriptLeaf(pubkey *btcec.PublicKey) txscript.TapLeaf {
	builder := txscript.NewScriptBuilder()
	builder.AddFullData(schnorr.SerializePubKey(pubkey))
	builder.AddOp(txscript.OP_CHECKSIG)
	script2, _ := builder.Script()
	return txscript.NewBaseTapLeaf(script2)
}
