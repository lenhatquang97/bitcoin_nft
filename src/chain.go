package src

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/m25lab/bitcoin_nft/src/enum"
)

func GetNetwork(chain enum.ChainValue) *chaincfg.Params {
	if chain == enum.Chain.Bitcoin {
		return &chaincfg.MainNetParams
	} else if chain == enum.Chain.RegTest {
		return &chaincfg.RegressionNetParams
	} else if chain == enum.Chain.Signet {
		return &chaincfg.SigNetParams
	} else {
		return &chaincfg.TestNet3Params
	}
}

func GetDefaultRPCPort(chain enum.ChainValue) int64 {
	if chain == enum.Chain.Bitcoin {
		return 8332
	} else if chain == enum.Chain.RegTest {
		return 18443
	} else if chain == enum.Chain.Signet {
		return 38332
	} else {
		return 18332
	}
}

func GetFirstInscriptionHeight(chain enum.ChainValue) int64 {
	if chain == enum.Chain.Bitcoin {
		return 767430
	} else if chain == enum.Chain.RegTest {
		return 0
	} else if chain == enum.Chain.Signet {
		return 112402
	} else {
		return 2413343
	}
}

func GetGenesisBlock(chain enum.ChainValue) *btcutil.Block {

	return nil
}

func AddressFromScript(script []byte, chain enum.ChainValue) (btcutil.Address, error) {
	return btcutil.DecodeAddress(string(script), GetNetwork(chain))
}

func JoinWithDataDir(dataDir string, chain enum.ChainValue) string {
	if chain == enum.Chain.Bitcoin {
		return dataDir
	} else {
		return dataDir + string(chain)
	}
}
