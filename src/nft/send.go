package nft

import (
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil"
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
	// client, err := GetBitcoinRPCClientForWalletCommand(opt, false)
	// if err != nil {
	// 	return err
	// }

	// unspentOutput, err := GetUnspentOutput(index)
	// if err != nil {
	// 	return err
	// }

	inscriptions, err := GetInscription(index)
	if err != nil {
		return err
	}

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
		break
	case enum.OutGoingType.InscriptionId:
		inscriptionId := data.OutGoingTypeData.(*src.InscriptionId)
		if inscriptionId == nil {
			return errors.New("inscription id data is invalid")
		}

		_, err = GetInscriptionSatPointById(index, inscriptionId)
		if err != nil {
			log.Printf("inscription %v not found", inscriptionId)
			return err
		}
		break
	case enum.OutGoingType.Amount:
		break
	}

	return nil
}
