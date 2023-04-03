package nft

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"github.com/m25lab/bitcoin_nft/src/utils"
)

type TransactionBuilder struct {
	Amounts             map[wire.OutPoint]btcutil.Amount
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
	dustLimit := mempool.GetDustThreshold(&wire.TxOut{
		Value:    int64(outputValue),
		PkScript: scriptPub,
	})

	if int64(outputValue) < dustLimit {
		return nil, errors.New("")
	}

	transactionBuilder := &TransactionBuilder{}

	res := BuildTransaction(transactionBuilder)

	return res, nil
}

func SelectOutGoing(transactionBuilder *TransactionBuilder) (*TransactionBuilder, error) {
	for inscribeSatPoint, inscriptionId := range transactionBuilder.Inscriptions {
		if transactionBuilder.OutGoing.OutPoint == inscribeSatPoint.OutPoint &&
			transactionBuilder.OutGoing.OffSet != inscribeSatPoint.OffSet {
			return nil, errors.New("_UTXO_CONTAIN_ADDITIONAL_TRANSACTION_")
		}
	}

	outGoingOutpoint, ok := transactionBuilder.Amounts[transactionBuilder.OutGoing.OutPoint]
	if !ok {
		return nil, errors.New("_OUTPOINT_NOT_IN_WALLET_")
	}

	delete(transactionBuilder.Utxos, transactionBuilder.OutGoing.OutPoint)
	transactionBuilder.Inputs = append(transactionBuilder.Inputs, transactionBuilder.OutGoing.OutPoint)
	transactionBuilder.Outputs = append(transactionBuilder.Outputs, utils.Account{
		Address: transactionBuilder.Recipient,
		Amount:  outGoingOutpoint,
	})

	return transactionBuilder, nil
}

func CalculateSatOffset(transactionBuilder *TransactionBuilder) (int64, error) {
	var satOffset int64
	for _, outpoint := range transactionBuilder.Inputs {
		if outpoint == transactionBuilder.OutGoing.OutPoint {
			return satOffset + transactionBuilder.OutGoing.OffSet, nil
		} else {
			satOffset += int64(transactionBuilder.Amounts[outpoint])
		}
	}

	return 0, errors.New("_OUTGOING_NOT_FOUND_")
}

func AlignOutGoing(transactionBuilder *TransactionBuilder) (*TransactionBuilder, error) {
	if len(transactionBuilder.Outputs) != 1 {
		return nil, errors.New("_ONLY_ONE_OUTPUT_")
	}

	if transactionBuilder.Outputs[0].Address != transactionBuilder.Recipient {
		return nil, errors.New("_FIST_OUTPUT_IS_RECIPIENT_")
	}

	satOffset, err := CalculateSatOffset(transactionBuilder)
	if err != nil {
		return nil, err
	}

	if satOffset == 0 {
		fmt.Println("_OUTGOING_IS_ALIGNED_")
	} else {
		fmt.Sprintf("Aligned outgoing with %v sat padding output", satOffset)

		transactionBuilder.Outputs = append(transactionBuilder.Outputs[1:], transactionBuilder.Outputs...)
		address := transactionBuilder.UnusedChangeAddress[0]
		transactionBuilder.UnusedChangeAddress = append(transactionBuilder.UnusedChangeAddress[:1], transactionBuilder.UnusedChangeAddress[2:]...)
		transactionBuilder.Outputs[0] = utils.Account{
			Address: address,
			Amount:  btcutil.Amount(satOffset),
		}

		transactionBuilder.Outputs[len(transactionBuilder.Outputs)-1].Amount -= btcutil.Amount(satOffset)
	}

	return transactionBuilder, nil
}

func PadAlignmentOutput(transactionBuilder *TransactionBuilder) (*TransactionBuilder, error) {
	if transactionBuilder.Outputs[0].Address == transactionBuilder.Recipient {
		fmt.Println("no alignment output")
	} else {
		scriptPub := transactionBuilder.Recipient.ScriptAddress()
		dustLimit := mempool.GetDustThreshold(&wire.TxOut{
			Value:    int64(transactionBuilder.Outputs[0].Amount),
			PkScript: scriptPub,
		})
		if int64(transactionBuilder.Outputs[0].Amount) >= dustLimit {
			fmt.Println("no padding needed")
		} else {
			utxo, size, err := SelectCardinalUtxo(transactionBuilder, dustLimit-int64(transactionBuilder.Outputs[0].Amount))
			if err != nil {

			}
		}
	}

	return transactionBuilder, nil
}

func SelectCardinalUtxo(builder *TransactionBuilder, minimumValue int64) (outpoint *wire.OutPoint, amount btcutil.Amount, err error) {
	inscriptionUtxos := make(map[wire.OutPoint]string)
	for satPoint := range builder.Inscriptions {
		inscriptionUtxos[satPoint.OutPoint] = ""
	}

	for utxo := range builder.Utxos {
		_, ok := inscriptionUtxos[utxo]
		if ok {
			continue
		}

		value, ok := builder.Amounts[utxo]
		if !ok {
			fmt.Println("utxo not found (SelectCardinalUtxo)")
			continue
		}
		if int64(value) >= minimumValue {
			outpoint = &utxo
			amount = value
			break
		}
	}

	if outpoint == nil {
		err = errors.New("Not enough cardinal utxo")
		fmt.Println(err)
		return
	}

	delete(builder.Utxos, *outpoint)
	return
}

func BuildTransaction(transactionBuilder *TransactionBuilder) *wire.MsgTx {
	return nil
}
