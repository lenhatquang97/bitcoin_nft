package nft

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/m25lab/bitcoin_nft/src/utils"
)

type InscriptionId struct {
	TxID  string
	Index int32
}

func GetInscriptionIDStore(inscriptionId *InscriptionId) ([]byte, error) {
	if inscriptionId == nil {
		return nil, errors.New("Inscription ID is nil")
	}

	var res []byte
	res = append(res, []byte(inscriptionId.TxID)...)
	res = append(res, utils.IntToBytes(int(inscriptionId.Index))...)
	return res, nil
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
