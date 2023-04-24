package nft

import (
	"context"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

type TransactionOutput struct {
	TxID          string
	Confirmations int64
}

/*
* No need to review
 */

func RetrieveTransaction() (*lnrpc.TransactionDetails, error) {
	lndConn, err := nft.GetLndGrpcSetup()
	if err != nil {
		return nil, err
	}
	defer lndConn.Close()

	lncli := lnrpc.NewLightningClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	txReq := lnrpc.GetTransactionsRequest{}

	output, err := lncli.GetTransactions(ctx, &txReq)
	if err != nil {
		return nil, err
	}
	return output, nil
}
