package src

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src/utils"
)

type SatPoint struct {
	OutPoint wire.OutPoint
	OffSet   int64
}

func GetSatPointStore(sat *SatPoint) ([]byte, error) {
	if sat == nil {
		return nil, errors.New("Sat point is nil")
	}

	if &sat.OutPoint == nil {
		return nil, errors.New("Outpoint of satpoint is nil")
	}

	var key []byte
	txId := blockchain.HashToBig(&sat.OutPoint.Hash)
	key = append(key, txId.Bytes()...)
	key = append(key, utils.IntToBytes(int(sat.OutPoint.Index))...)
	key = append(key, utils.IntToBytes(int(sat.OffSet))...)
	return key, nil
}

func LoadIntoSatPoint(input []byte) (*SatPoint, error) {
	if len(input) != 44 {
		return nil, errors.New("Satpoint byte data is invalid")
	}

	hashByte := input[:32]
	index := input[32:36]
	offset := input[36:]
	hash, err := chainhash.NewHash(hashByte)
	if err != nil {
		return nil, err
	}

	return &SatPoint{
		OutPoint: wire.OutPoint{
			Hash:  *hash,
			Index: uint32(utils.BytesToInt(index)),
		}, OffSet: utils.BytesToInt(offset),
	}, nil
}

func LoadIntoInscriptionID(input []byte) (*InscriptionId, error) {
	if len(input) != 36 {
		return nil, errors.New(fmt.Sprintf("Inscription byte data len is invalid (36 is right) but it is %v", len(input)))
	}

	var Buf bytes.Buffer
	Buf.Write(input[:32])
	txId := Buf.String()
	index := utils.BytesToInt(input[32:])
	return &InscriptionId{
		TxID:  txId,
		Index: int32(index),
	}, nil
}
