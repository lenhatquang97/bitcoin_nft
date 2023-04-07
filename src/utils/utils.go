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
