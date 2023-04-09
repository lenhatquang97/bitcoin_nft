package nft

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
)

type Sats struct {
	TSV string
}

func SatsRun(opt *Options, sats *Sats) error {
	index, err := Open(opt)
	if err != nil {
		return err
	}
	Update(index)
	utxos, err := GetUnspentOutputRanges(index)
	if err != nil {
		return err
	}

	if sats.TSV != "" {
		res, err := SatsFromTSV(utxos, sats.TSV)
		if err != nil {
			return err
		}

		for _, item := range res {
			fmt.Print(item)
		}
	} else {
		res, err := RareSats(utxos)
		if err != nil {
			return err
		}

		for _, item := range res {
			fmt.Print(item)
		}
	}

	return nil
}

type SatRunRes struct {
	Outpoint wire.OutPoint
	TSV      string
}

func SatsFromTSV(unspentOutputRanges []*UnspentOutputRange, tsv string) ([]SatRunRes, error) {
	return nil, errors.New("not yet impl")
}

type RateSatRes struct {
	Outpoint    wire.OutPoint
	Sat         int64
	StartOffset int64
	//Rarity string: no use
}

func RareSats(unspentOutputRanges []*UnspentOutputRange) ([]RateSatRes, error) {
	return nil, errors.New("not yet impl")
}
