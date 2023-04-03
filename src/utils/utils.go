package utils

import "github.com/btcsuite/btcd/btcutil"

type Account struct {
	Address btcutil.Address
	Amount  btcutil.Amount
}
