package nft

import "github.com/btcsuite/btcd/btcutil"

/*
* No need to review
 */
func Fee(feeRate float64, vsize float64) btcutil.Amount {
	return btcutil.Amount(feeRate * vsize)
}
