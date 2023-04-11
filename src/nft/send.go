package nft

import (
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
)

type SendData struct {
	Address          btcutil.Address
	OutGoingType     enum.OutGoingTypeValue
	OutGoingTypeData interface{}
	FeeRate          float64
}

type SendDataOutput struct {
	TxID string
}

func SendRun(opt *Options, data *SendData) error {
	if !data.Address.IsForNet(GetChainInfo(opt)) {
		return errors.New("Address is not valid")
	}

	index, err := Open(opt)
	if err != nil {
		return err
	}
	Update(index)
	client, err := GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		return err
	}

	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

	inscriptions, err := GetInscription(index)
	if err != nil {
		return err
	}

	outGoing := new(src.SatPoint)

	switch data.OutGoingType {
	case enum.OutGoingType.Satpoint:
		satpoint := data.OutGoingTypeData.(*src.SatPoint)
		if satpoint == nil {
			return errors.New("Satpoint data is invalid")
		}

		for inscriptionSatpoint := range inscriptions {
			if *satpoint == inscriptionSatpoint {
				s := "inscription must be spent by inscription id"
				fmt.Println(s)
				return errors.New(s)
			}
		}
		outGoing = satpoint
	case enum.OutGoingType.InscriptionId:
		inscriptionId := data.OutGoingTypeData.(*src.InscriptionId)
		if inscriptionId == nil {
			return errors.New("inscription id data is invalid")
		}

		outGoing, err = GetInscriptionSatPointById(index, inscriptionId)
		if err != nil {
			log.Printf("inscription %v not found", inscriptionId)
			return err
		}
	case enum.OutGoingType.Amount:
		allInscriptionOutput := make(map[wire.OutPoint]string)
		for inscriptionOuput := range inscriptions {
			allInscriptionOutput[inscriptionOuput.OutPoint] = ""
		}

		var walletInscriptionOutput []*wire.OutPoint
		for utxo := range unspentOutput {
			_, ok := allInscriptionOutput[utxo]
			if ok {
				walletInscriptionOutput = append(walletInscriptionOutput, &utxo)
			}
		}

		amount := data.OutGoingTypeData.(btcutil.Amount)
		err = client.LockUnspent(false, walletInscriptionOutput)
		if err != nil {
			return err
		}

		res, err := client.SendToAddress(data.Address, amount)
		if err != nil {
			return err
		}

		fmt.Println("Chain hash: ", res)
	}

	firstAddress, err := client.GetRawChangeAddressType("", "bech32m")
	if err != nil {
		return err
	}

	secondAddress, err := client.GetRawChangeAddressType("", "bech32m")
	if err != nil {
		return err
	}

	var commitTxChange []btcutil.Address
	commitTxChange = append(commitTxChange, firstAddress)
	commitTxChange = append(commitTxChange, secondAddress)
	unsignedCommitTx, err := BuildTransactionWithPostage(*outGoing, inscriptions, unspentOutput, data.Address, commitTxChange, data.FeeRate)
	if err != nil {
		return err
	}

	signedTx, _, err := client.SignRawTransactionWithWallet(unsignedCommitTx)
	if err != nil {
		return err
	}

	res, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		return err
	}
	fmt.Println("Tx hash: ", res)
	return nil
}
