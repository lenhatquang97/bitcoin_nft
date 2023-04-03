package nft

import "github.com/btcsuite/btcd/btcutil"

func Fee(feeRate float64, vsize float64) btcutil.Amount {
	return btcutil.Amount(feeRate * vsize)
}
