package main

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/m25lab/bitcoin_nft/src/wallet"
)

const (
	testTaprootKeyFamily = 77
)

var (
	hexDecode = func(keyStr string) []byte {
		keyBytes, _ := hex.DecodeString(keyStr)
		return keyBytes
	}

	dummyInternalKey, _ = btcec.ParsePubKey(hexDecode(
		"03464805f5468e294d88cf15a3f06aef6c89d63ef1bd7b42db2e0c74c1ac" +
			"eb90fe",
	))
)

func main() {
	wallet.SendCoinWithTaproot(testTaprootKeyFamily, dummyInternalKey)
}
