package nft

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/m25lab/bitcoin_nft/src"
)

type Output struct {
	Commit        int64
	InscriptionID src.InscriptionId
	Reveal        int64
	Fee           int64
}

type Inscribe struct {
	SatPoint      src.SatPoint
	FeeRate       float64
	CommitFeeRate float64
	File          string
	NoBackup      bool
	NoLimit       bool
	DryRun        bool
	Destination   btcutil.Address
}
