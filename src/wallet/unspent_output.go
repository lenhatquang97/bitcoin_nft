package wallet

import (
	"context"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/m25lab/bitcoin_nft/src/layer2"
	"github.com/m25lab/bitcoin_nft/src/model"
)

func GetUnspentOutput() (map[string]int64, error) {
	lndConn, err := layer2.GetLndGrpcSetup()
	if err != nil {
		return nil, err
	}
	defer lndConn.Close()

	walletClient := walletrpc.NewWalletKitClient(lndConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	unspentRes, err := walletClient.ListUnspent(ctx, &walletrpc.ListUnspentRequest{})
	if err != nil {
		return nil, err
	}
	res := model.MappingOutpoint(unspentRes.Utxos)
	return res, nil
}
