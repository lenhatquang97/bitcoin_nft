package nft

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"github.com/m25lab/bitcoin_nft/src/inscript"
	"github.com/m25lab/bitcoin_nft/src/model"
	"github.com/m25lab/bitcoin_nft/src/utils"
)

const (
	ADDITIONAL_INPUT_VBYTES  = 58
	MAX_POSTAGE              = 2 * 10000
	TARGET_POSTAGE           = 10000
	ADDITIONAL_OUTPUT_VBYTES = 43
)

type TransactionBuilder struct {
	Amounts             map[string]int64 // map of outpoint
	ChangeAddresses     []string
	FeeRate             float64
	Inputs              []model.Outpoint
	Inscriptions        map[string]InscriptionId // map of satpoint
	OutGoing            SatPoint
	Outputs             []utils.Account
	Recipient           string
	UnusedChangeAddress []string
	Utxos               map[string]string // map of outpoint serialize
	Target              enum.TargetValue
	OutputValue         btcutil.Amount
}

func BuildTransactionWithValue(outGoing SatPoint,
	inscriptions map[string]InscriptionId, // map of satpoint serialize
	amount map[string]int64, //map of outpoint
	recipient string,
	change []string,
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
	dustLimit := mempool.GetDustThreshold(&wire.TxOut{
		Value:    int64(outputValue),
		PkScript: []byte(recipient),
	})

	if int64(outputValue) < dustLimit {
		return nil, errors.New("")
	}

	transactionBuilder := &TransactionBuilder{
		OutGoing:        outGoing,
		Inscriptions:    inscriptions,
		Amounts:         amount,
		Recipient:       recipient,
		ChangeAddresses: change,
		FeeRate:         feeRate,
		Target:          enum.Target.Value,
		OutputValue:     outputValue,
	}

	return BuildTransaction(transactionBuilder)
}

func BuildTransactionWithPostage(outGoing SatPoint,
	inscriptions map[string]InscriptionId, // map of satpoint
	amount map[string]int64, // map of outpoint
	recipient string,
	change []string,
	feeRate float64) (*wire.MsgTx, error) {

	transactionBuilder := &TransactionBuilder{
		OutGoing:        outGoing,
		Inscriptions:    inscriptions,
		Amounts:         amount,
		Recipient:       recipient,
		ChangeAddresses: change,
		FeeRate:         feeRate,
		Target:          enum.Target.Value,
	}

	return BuildTransaction(transactionBuilder)
}

func SelectOutGoing(transactionBuilder *TransactionBuilder) (*TransactionBuilder, error) {
	for inscribeSatPoint, _ := range transactionBuilder.Inscriptions {
		satpointDeserialize, err := DeserializeSatPoint(inscribeSatPoint)
		if err != nil {
			return nil, err
		}

		if transactionBuilder.OutGoing.OutPoint.Serialize() == satpointDeserialize.OutPoint.Serialize() &&
			transactionBuilder.OutGoing.OffSet != satpointDeserialize.OffSet {
			return nil, errors.New("_UTXO_CONTAIN_ADDITIONAL_TRANSACTION_")
		}
	}

	outGoingOutpoint, ok := transactionBuilder.Amounts[transactionBuilder.OutGoing.OutPoint.Serialize()]
	if !ok {
		return nil, errors.New("_OUTPOINT_NOT_IN_WALLET_")
	}

	delete(transactionBuilder.Utxos, transactionBuilder.OutGoing.OutPoint.Serialize())
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
		if outpoint.Serialize() == transactionBuilder.OutGoing.OutPoint.Serialize() {
			return satOffset + transactionBuilder.OutGoing.OffSet, nil
		} else {
			satOffset += int64(transactionBuilder.Amounts[outpoint.Serialize()])
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
			Amount:  satOffset,
		}

		transactionBuilder.Outputs[len(transactionBuilder.Outputs)-1].Amount -= satOffset
	}

	return transactionBuilder, nil
}

func PadAlignmentOutput(transactionBuilder *TransactionBuilder) (*TransactionBuilder, error) {
	if transactionBuilder.Outputs[0].Address == transactionBuilder.Recipient {
		fmt.Println("no alignment output")
	} else {
		scriptPub := []byte(transactionBuilder.Recipient)
		dustLimit := mempool.GetDustThreshold(&wire.TxOut{
			Value:    int64(transactionBuilder.Outputs[0].Amount),
			PkScript: scriptPub,
		})
		if int64(transactionBuilder.Outputs[0].Amount) >= dustLimit {
			fmt.Println("no padding needed")
		} else {
			utxo, size, err := SelectCardinalUtxo(transactionBuilder, dustLimit-int64(transactionBuilder.Outputs[0].Amount))
			if err != nil {
				fmt.Println(err)
				return transactionBuilder, err
			}
			transactionBuilder.Inputs = append(transactionBuilder.Inputs[1:], transactionBuilder.Inputs...)
			transactionBuilder.Inputs[0] = *utxo
			transactionBuilder.Outputs[0].Amount += int64(size)
			fmt.Printf("padded alignment output to %v with additional %v sat input", transactionBuilder.Outputs[0].Amount, size)
		}
	}

	return transactionBuilder, nil
}

func EstimateVbyteWith(inputs int64, outputs []string) btcutil.Amount {
	var txIn []*wire.TxIn
	var txOut []*wire.TxOut
	for i := 0; i < int(inputs); i++ {
		emptyScript, _ := txscript.NewScriptBuilder().Script()
		emptyWitness := wire.TxWitness{}
		txIn = append(txIn, &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{},
			SignatureScript:  emptyScript,
			Witness:          emptyWitness,
			Sequence:         0,
		})
	}

	for _, address := range outputs {
		txOut = append(txOut, &wire.TxOut{
			Value:    0,
			PkScript: []byte(address),
		})
	}

	tx := wire.MsgTx{
		Version:  1,
		LockTime: 0,
		TxIn:     txIn,
		TxOut:    txOut,
	}

	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(&tx))

	return btcutil.Amount(txSize)
}

func EstimateVbytes(builder *TransactionBuilder) btcutil.Amount {
	var addresses []string
	for _, acc := range builder.Outputs {
		addresses = append(addresses, acc.Address)
	}
	return EstimateVbyteWith(int64(len(builder.Inputs)), addresses)
}

func EstimateFee(builder *TransactionBuilder) btcutil.Amount {
	return Fee(builder.FeeRate, float64(EstimateVbytes(builder)))
}

func AddValue(builder *TransactionBuilder) (*TransactionBuilder, error) {
	estimateFee := EstimateFee(builder)

	var minValue btcutil.Amount
	if builder.Target == enum.Target.Value {
		address := []byte(builder.Outputs[len(builder.Outputs)-1].Address)
		minValue = builder.OutputValue + btcutil.Amount(mempool.GetDustThreshold(&wire.TxOut{
			Value:    0,
			PkScript: address,
		}))
	} else if builder.Target == enum.Target.PostAge {
		minValue = 0
	}

	total := minValue + estimateFee
	deficit := total - btcutil.Amount(builder.Outputs[len(builder.Outputs)-1].Amount)
	if deficit > 0 {
		needed := deficit + Fee(builder.FeeRate, float64(ADDITIONAL_INPUT_VBYTES))
		utxo, value, err := SelectCardinalUtxo(builder, int64(needed))
		if err != nil {
			return builder, err
		}
		builder.Inputs = append(builder.Inputs, *utxo)
		builder.Outputs[len(builder.Outputs)-1].Amount += int64(value)
		fmt.Printf("added %v sat input to cover %v sat deficit", value, deficit)
	}

	return builder, nil
}

func StripValue(builder *TransactionBuilder) (*TransactionBuilder, error) {
	satOffset, err := CalculateSatOffset(builder)
	if err != nil {
		return builder, err
	}

	var totalOutputAmount btcutil.Amount
	isFind := false
	for _, acc := range builder.Outputs {
		if acc.Address == builder.Recipient {
			isFind = true
		}
		totalOutputAmount += btcutil.Amount(acc.Amount)
	}

	if !isFind {
		return builder, errors.New("couldn't find output that contain the index")
	}

	value := totalOutputAmount - btcutil.Amount(satOffset)
	excess := value - Fee(builder.FeeRate, float64(EstimateVbytes(builder)))
	var maxAmount, targetAmount btcutil.Amount
	if builder.Target == enum.Target.PostAge {
		maxAmount, targetAmount = MAX_POSTAGE, TARGET_POSTAGE
	} else if builder.Target == enum.Target.Value {
		maxAmount, targetAmount = builder.OutputValue, builder.OutputValue
	} else {
		return builder, errors.New("Transaction builder - Target field is invalid!")
	}

	res := value - targetAmount
	amount := btcutil.Amount(mempool.GetDustThreshold(&wire.TxOut{PkScript: []byte(builder.UnusedChangeAddress[len(builder.UnusedChangeAddress)-1])})) + Fee(builder.FeeRate, ADDITIONAL_OUTPUT_VBYTES)
	if excess > maxAmount && res > amount {
		fmt.Printf("stripped %v sat", res)
		address := builder.UnusedChangeAddress[0]
		builder.UnusedChangeAddress = append(builder.UnusedChangeAddress[:1], builder.UnusedChangeAddress[2:]...)
		builder.Outputs = append(builder.Outputs, utils.Account{
			Address: address,
			Amount:  int64(res),
		})
	}

	return builder, nil
}

func DeductFee(builder *TransactionBuilder) (*TransactionBuilder, error) {
	satOffset, err := CalculateSatOffset(builder)
	if err != nil {
		fmt.Println(err)
		return builder, err
	}

	estimateFee := EstimateFee(builder)
	var totalOutputAmount btcutil.Amount
	for _, acc := range builder.Outputs {
		totalOutputAmount += btcutil.Amount(acc.Amount)
	}

	acc := builder.Outputs[len(builder.Outputs)-1]
	if totalOutputAmount-estimateFee <= btcutil.Amount(satOffset) {
		return builder, errors.New("invariant: deducting fee does not consume sat")
	}

	if btcutil.Amount(acc.Amount) < estimateFee {
		fmt.Printf("invariant: last output can pay fee: %v %v", acc.Amount, estimateFee)
	}

	return builder, nil
}

func Build(builder *TransactionBuilder) (*wire.MsgTx, error) {
	recipient := []byte(builder.Recipient)
	var txIns []*wire.TxIn
	var txOuts []*wire.TxOut
	for _, outpoint := range builder.Inputs {
		emptyScript, _ := txscript.NewScriptBuilder().Script()
		emptyWitness := wire.TxWitness{}
		txHash, err := chainhash.NewHashFromStr(outpoint.TxidStr)
		if err != nil {
			return nil, err
		}
		txIns = append(txIns, &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  *txHash,
				Index: outpoint.OutputIndex,
			},
			Sequence:        inscript.ENABLE_RBF_NO_LOCKTIME,
			SignatureScript: emptyScript,
			Witness:         emptyWitness,
		})
	}

	for _, output := range builder.Outputs {
		txOuts = append(txOuts, &wire.TxOut{
			Value:    int64(output.Amount),
			PkScript: []byte(output.Address),
		})
	}
	tx := &wire.MsgTx{
		Version:  1,
		LockTime: 0,
		TxIn:     txIns,
		TxOut:    txOuts,
	}

	// check out going sat contain in utxos

	// check input spend out going sat

	var satOffset int64
	isFind := false
	for _, txIn := range tx.TxIn {
		if builder.OutGoing.OutPoint.IsEqual(&txIn.PreviousOutPoint) {
			satOffset += builder.OutGoing.OffSet
			isFind = true
			break
		} else {
			outpoint := model.ConvertToOutpoint(&txIn.PreviousOutPoint)
			satOffset += int64(builder.Amounts[outpoint.Serialize()])
		}
	}

	// check isFind
	if !isFind {

	}

	var outputEnd int64
	isFind = false
	for _, txOut := range tx.TxOut {
		outputEnd += txOut.Value
		if outputEnd > satOffset {
			// check script pubkey and recipient
			if !bytes.Equal(recipient, txOut.PkScript) {
				// return err
			}
			isFind = true
			break
		}
	}

	// check found
	if !isFind {

	}

	//      "invariant: recipient address appears exactly once in outputs",

	//       "invariant: change addresses appear at most once in outputs",

	var offset int64
	for _, output := range tx.TxOut {
		if bytes.Equal([]byte(builder.Recipient), output.PkScript) {
			slop := Fee(builder.FeeRate, ADDITIONAL_OUTPUT_VBYTES)
			fmt.Println(slop)

			// check asert value
		} else {

			// check asert value
		}

		offset += output.Value
	}

	var actualFee btcutil.Amount
	for _, txIn := range tx.TxIn {
		outpoint := model.ConvertToOutpoint(&txIn.PreviousOutPoint)
		actualFee += btcutil.Amount(builder.Amounts[outpoint.Serialize()])
	}

	for _, txOut := range tx.TxOut {
		actualFee -= btcutil.Amount(txOut.Value)
	}

	copyTx := tx.Copy()

	for _, txIn := range copyTx.TxIn {
		txIn.Witness = append(txIn.Witness, make([]byte, schnorr.SignatureSize))
	}

	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(copyTx))
	expectedFee := Fee(builder.FeeRate, float64(txSize))

	if actualFee != expectedFee {
		// alert err
	}

	for _, txOut := range tx.TxOut {
		// assert err
		fmt.Println(txOut)
	}

	return tx, nil
}

func SelectCardinalUtxo(builder *TransactionBuilder, minimumValue int64) (outpoint *model.Outpoint, amount btcutil.Amount, err error) {
	inscriptionUtxos := make(map[string]string) // map outpoint
	for satPoint := range builder.Inscriptions {
		satpointDeserialize, err := DeserializeSatPoint(satPoint)
		if err != nil {
			return nil, 0, err
		}
		inscriptionUtxos[satpointDeserialize.OutPoint.Serialize()] = ""
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
			var err error
			outpoint, err = model.DeserializeOutpoint(utxo)
			if err != nil {
				return nil, 0, err
			}
			amount = btcutil.Amount(value)
			break
		}
	}

	if outpoint == nil {
		err = errors.New("Not enough cardinal utxo")
		fmt.Println(err)
		return nil, 0, err
	}

	delete(builder.Utxos, outpoint.Serialize())
	return outpoint, amount, nil
}

func BuildTransaction(transactionBuilder *TransactionBuilder) (*wire.MsgTx, error) {
	tx, err := SelectOutGoing(transactionBuilder)
	if err != nil {
		return nil, err
	}

	tx, err = AlignOutGoing(tx)
	if err != nil {
		return nil, err
	}

	tx, err = PadAlignmentOutput(transactionBuilder)
	if err != nil {
		return nil, err
	}

	tx, err = AddValue(transactionBuilder)
	if err != nil {
		return nil, err
	}

	tx, err = StripValue(tx)
	if err != nil {
		return nil, err
	}

	tx, err = DeductFee(tx)
	if err != nil {
		return nil, err
	}

	return Build(tx)
}
