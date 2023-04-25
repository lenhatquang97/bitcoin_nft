package model

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightningnetwork/lnd/lnrpc"
)

const SEPARATOR = "@m25@"

type Outpoint struct {
	TxidBytes   []byte
	TxidStr     string
	OutputIndex uint32
}

func (o *Outpoint) Serialize() string {
	res := string(o.TxidBytes) + SEPARATOR + o.TxidStr + SEPARATOR + strconv.Itoa(int(o.OutputIndex))
	return res
}

func ConvertToOutpoint(o *wire.OutPoint) *Outpoint {
	txId := o.Hash.String()
	return &Outpoint{
		TxidStr:     txId,
		TxidBytes:   []byte(txId),
		OutputIndex: o.Index,
	}
}

func (o *Outpoint) IsEqual(out *wire.OutPoint) bool {
	txHash, err := chainhash.NewHashFromStr(o.TxidStr)
	if err != nil {
		log.Print(err)
		return false
	}

	return *txHash == out.Hash && o.OutputIndex == out.Index
}

func DeserializeOutpoint(value string) (*Outpoint, error) {
	outpointValue := strings.Split(value, SEPARATOR)
	if len(outpointValue) != 3 {
		panic(fmt.Sprintf("Deserialize outpoint failed - %s, len = %d", value, len(outpointValue)))
	}

	outputIndex, err := strconv.Atoi(outpointValue[2])
	if err != nil {
		return nil, err
	}

	return &Outpoint{
		TxidBytes:   []byte(outpointValue[0]),
		TxidStr:     outpointValue[1],
		OutputIndex: uint32(outputIndex),
	}, nil
}

func MappingOutpoint(inps []*lnrpc.Utxo) map[string]int64 {
	res := make(map[string]int64)
	for _, item := range inps {
		outpoint := Outpoint{
			TxidBytes:   item.Outpoint.TxidBytes,
			TxidStr:     item.Outpoint.TxidStr,
			OutputIndex: item.Outpoint.OutputIndex,
		}
		res[outpoint.Serialize()] = item.AmountSat
	}

	return res
}
