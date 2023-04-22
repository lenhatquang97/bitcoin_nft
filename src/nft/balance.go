package nft

import (
	"context"
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
)

/*
* Have reviewed in 9/4/2023
* Balance:
* Step 1: Connect node
* Step 2: Get all stored NFT (inscriptions go with satpoint)
* Step 3: Balance = Total balance - stored NFT
* Needs integration test
 */

func BalanceRun() error {
	lndConn, err := GetLndGrpcSetup()
	if err != nil {
		return err
	}
	defer lndConn.Close()

	lncli := lnrpc.NewLightningClient(lndConn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	walletBalanceReq := lnrpc.WalletBalanceRequest{}
	result, err := lncli.WalletBalance(ctx, &walletBalanceReq)
	if err != nil {
		return err
	}
	fmt.Printf("Total balance: %d\n", result.TotalBalance)
	fmt.Printf("Confirmed balance: %d\n", result.ConfirmedBalance)
	fmt.Printf("Unconfirmed balance: %d\n", result.UnconfirmedBalance)

	return nil
}
