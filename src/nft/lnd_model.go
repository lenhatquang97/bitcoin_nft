package nft

import (
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type LndClient struct {
	client   lnrpc.LightningClient
	wg       sync.WaitGroup
	params   *chaincfg.Params
	timeout  time.Duration
	adminMac string
}
