package src

import (
	"github.com/btcsuite/btcd/wire"
)

type SatPoint struct {
	OutPoint wire.OutPoint
	OffSet   int64
}
