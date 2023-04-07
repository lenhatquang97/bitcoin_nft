package utils

import (
	"github.com/btcsuite/btcd/btcutil"
	"math/big"
)

type Account struct {
	Address btcutil.Address
	Amount  btcutil.Amount
}

func IntToBytes(i int) []byte {
	if i > 0 {
		return append(big.NewInt(int64(i)).Bytes(), byte(1))
	}
	return append(big.NewInt(int64(i)).Bytes(), byte(0))
}

func BytesToInt(b []byte) int64 {
	if b[len(b)-1] == 0 {
		return -big.NewInt(0).SetBytes(b[:len(b)-1]).Int64()
	}
	return big.NewInt(0).SetBytes(b[:len(b)-1]).Int64()
}
