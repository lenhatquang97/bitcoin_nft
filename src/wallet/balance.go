package wallet

import (
	"context"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/m25lab/bitcoin_nft/src/layer2"
)

/*
* Have reviewed in 9/4/2023
* Balance:
* Step 1: Connect node
* Step 2: Get all stored NFT (inscriptions go with satpoint)
* Step 3: Balance = Total balance - stored NFT
* Needs integration test
 */
type BalanceOutput struct {
	TotalBalance       int64
	ConfirmedBalance   int64
	UnconfirmedBalance int64
}

func RetrieveBalance() (*BalanceOutput, error) {
	lndConn, err := layer2.GetLndGrpcSetup()
	if err != nil {
		return nil, err
	}
	defer lndConn.Close()

	lncli := lnrpc.NewLightningClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	walletBalanceReq := lnrpc.WalletBalanceRequest{}
	result, err := lncli.WalletBalance(ctx, &walletBalanceReq)
	if err != nil {
		return nil, err
	}
	output := BalanceOutput{
		TotalBalance:       result.TotalBalance,
		ConfirmedBalance:   result.ConfirmedBalance,
		UnconfirmedBalance: result.UnconfirmedBalance,
	}
	return &output, nil
}
