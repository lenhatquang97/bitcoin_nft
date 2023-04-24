package nft

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/m25lab/bitcoin_nft/src/utils"
)

type SatPoint struct {
	OutPoint Outpoint
	OffSet   int64
}

func (s *SatPoint) Serialize() string {
	outpointSerialized := s.OutPoint.Serialize()
	offsetStr := strconv.Itoa(int(s.OffSet))
	return outpointSerialized + "::" + offsetStr
}

func DeserializeSatPoint(value string) (*SatPoint, error) {
	seprator := "::"
	deserializeValue := strings.Split(value, seprator)
	if len(deserializeValue) != 2 {
		panic(fmt.Sprintf("Deserialize satpoint failed %s - %d", value, len(deserializeValue)))
	}

	outpoint, err := DeserializeOutpoint(deserializeValue[0])
	if err != nil {
		return nil, err
	}

	offset, err := strconv.Atoi(deserializeValue[1])
	if err != nil {
		return nil, err
	}

	return &SatPoint{
		OutPoint: *outpoint,
		OffSet:   int64(offset),
	}, nil
}

func GetSatPointStore(sat *SatPoint) ([]byte, error) {
	if sat == nil {
		return nil, fmt.Errorf("sat point is nil")
	}

	if &sat.OutPoint == nil {
		return nil, fmt.Errorf("outpoint of satpoint is nil")
	}

	var key []byte
	txId := sat.OutPoint.TxidBytes
	key = append(key, txId...)
	key = append(key, utils.IntToBytes(int(sat.OutPoint.OutputIndex))...)
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

	idx := utils.BytesToInt(index)
	return &SatPoint{
		OutPoint: Outpoint{
			TxidBytes:   hashByte,
			TxidStr:     string(hashByte),
			OutputIndex: uint32(idx),
		}, OffSet: utils.BytesToInt(offset),
	}, nil
}
