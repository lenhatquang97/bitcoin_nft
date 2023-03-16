package initial

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// Check valid address: mempool.space
// P2TR Address for case 1 person
type AddressType int64

const (
	P2PKH  AddressType = 0
	P2WPKH AddressType = 1
	P2TR   AddressType = 2
)

func GenerateBTCAddress(addressType AddressType) ([]byte, string) {
	priv, err := btcec.NewPrivateKey()
	network := &chaincfg.TestNet3Params
	if err != nil {
		fmt.Printf("BitcoinNFT: Failed to create PrivateKey for %s", err)
	}
	switch addressType {
	case P2PKH:
		pubKeyHash := btcutil.Hash160(priv.PubKey().SerializeUncompressed())
		address, err := btcutil.NewAddressPubKeyHash(pubKeyHash, network)
		if err != nil {
			fmt.Printf("BitcoinNFT: Failed to create address with error: %s", err)
		}
		privateKeyBytes := priv.Serialize()
		return privateKeyBytes, address.EncodeAddress()
	case P2WPKH:
		pubKeyHash := btcutil.Hash160(priv.PubKey().SerializeUncompressed())
		address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, network)
		if err != nil {
			fmt.Printf("BitcoinNFT: Failed to create address with error: %s", err)
		}
		privateKeyBytes := priv.Serialize()
		return privateKeyBytes, address.EncodeAddress()
	case P2TR:
		tapKey := txscript.ComputeTaprootKeyNoScript(priv.PubKey())
		address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(tapKey), network)
		if err != nil {
			fmt.Printf("BitcoinNFT: Failed to create address with error: %s", err)
		}
		privateKeyBytes := priv.Serialize()
		return privateKeyBytes, address.EncodeAddress()
	}
	return nil, ""
}
