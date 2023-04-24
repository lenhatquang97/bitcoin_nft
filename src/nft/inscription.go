package nft

import (
	"fmt"

	"github.com/m25lab/bitcoin_nft/src/enum"
)

/*
* Review that needs to self-create inscriptions for Server running
 */

type InscriptionOutput struct {
	InscriptionID InscriptionId
	Location      SatPoint
	Explorer      string
}

func InscriptionRun(opt *Options) error {
	index, err := Open(opt)
	if err != nil {
		return err
	}

	inscriptions, err := GetInscription(index)
	if err != nil {
		return err
	}

	unspentOutputs, err := GetUnspentOutput()
	if err != nil {
		return err
	}

	explorer := ""
	switch opt.ChainArgument {
	case enum.Chain.Bitcoin:
		explorer = "https://ordinals.com/inscription/"
	case enum.Chain.RegTest:
		explorer = "http://localhost/inscription/"
	case enum.Chain.Signet:
		explorer = "https://signet.ordinals.com/inscription/"
	case enum.Chain.Testnet:
		explorer = "https://testnet.ordinals.com/inscription/"
	}

	var res []InscriptionOutput
	for location, inscriptionId := range inscriptions {
		locationDeserialize, err := DeserializeSatPoint(location)
		if err != nil {
			return err
		}

		_, ok := unspentOutputs[locationDeserialize.OutPoint.Serialize()]
		if ok {
			res = append(res, InscriptionOutput{
				InscriptionID: inscriptionId,
				Location:      *locationDeserialize,
				Explorer:      explorer,
			})
		}
	}

	fmt.Print(res)

	return nil
}
