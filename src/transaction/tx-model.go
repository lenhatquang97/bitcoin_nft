package transaction

import "github.com/btcsuite/btcd/wire"

type TxIn struct {
	PreviousOutPoint wire.OutPoint
	ScriptLength     uint32
	SignatureScript  []byte
	Witness          wire.TxWitness
	Sequence         uint32
}

type TxOut struct {
	Value        int64
	ScriptLength uint32
	PkScript     []byte
}

type Tx struct {
	Version  int32
	TxIn     []*wire.TxIn
	TxOut    []*wire.TxOut
	LockTime uint32
}
