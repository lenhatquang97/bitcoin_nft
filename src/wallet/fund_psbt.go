package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/m25lab/bitcoin_nft/src/layer2"
)

func FundPsbt() {
	lndConn, err := layer2.GetLndGrpcSetup()
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := walletrpc.NewWalletKitClient(lndConn)

	var rawInputs []*lnrpc.OutPoint

	rawInputs = append(rawInputs, &lnrpc.OutPoint{
		TxidStr:     "123455",
		OutputIndex: 0,
	})

	fundPsbtReq := walletrpc.FundPsbtRequest{
		Template: &walletrpc.FundPsbtRequest_Raw{
			Raw: &walletrpc.TxTemplate{
				Inputs: rawInputs,
			},
		},
	}
	fundRes, err := client.FundPsbt(ctx, &fundPsbtReq)
	if err != nil {
		fmt.Println(err)
		return
	}
	finalizePsbtReq := walletrpc.FinalizePsbtRequest{
		FundedPsbt: fundRes.FundedPsbt,
	}
	finalRes, err := client.FinalizePsbt(ctx, &finalizePsbtReq)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(finalRes)
}
