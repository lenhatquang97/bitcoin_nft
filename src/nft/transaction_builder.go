package nft

import (
	"errors"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"github.com/m25lab/bitcoin_nft/src/utils"
)

type TransactionBuilder struct {
	Amount              map[wire.OutPoint]btcutil.Amount
	ChangeAddresses     map[btcutil.Address]string
	FeeRate             float64
	Inputs              []wire.OutPoint
	Inscriptions        map[src.SatPoint]src.InscriptionId
	OutGoing            src.SatPoint
	Outputs             []utils.Account
	Recipient           btcutil.Address
	UnusedChangeAddress []btcutil.Address
	Utxos               map[wire.OutPoint]string
	Target              enum.TargetValue
}

func BuildTransactionWithValue(outGoing src.SatPoint,
	inscriptions map[src.SatPoint]src.InscriptionId,
	amount map[wire.OutPoint]btcutil.Amount,
	recipient btcutil.Address,
	feeRate float64,
	outputValue btcutil.Amount) (*wire.MsgTx, error) {

	// check is dust

	// The output is considered dust if the cost to the network to spend the
	// coins is more than 1/3 of the minimum free transaction relay fee.
	// minFreeTxRelayFee is in Satoshi/KB, so multiply by 1000 to
	// convert to bytes.
	//
	// Using the typical values for a pay-to-pubkey-hash transaction from
	// the breakdown above and the default minimum free transaction relay
	// fee of 1000, this equates to values less than 546 satoshi being
	// considered dust.
	//
	// The following is equivalent to (value/totalSize) * (1/3) * 1000
	// without needing to do floating point math.

	//	mp.cfg.Policy.MinRelayTxFee
	scriptPub := recipient.ScriptAddress()
	if mempool.IsDust(&wire.TxOut{
		Value:    int64(outputValue),
		PkScript: scriptPub,
	}, mempool.DefaultMinRelayTxFee) {
		return nil, errors.New("")
	}

	transactionBuilder := &TransactionBuilder{}

	res := BuildTransaction(transactionBuilder)

	return res, nil
}

func BuildTransaction(transactionBuilder *TransactionBuilder) *wire.MsgTx {
	return nil
}
