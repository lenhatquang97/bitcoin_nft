package layer1

import (
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
)

func LoadCerts() ([]byte, error) {
	certHomeDir := btcutil.AppDataDir("btcd", false)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func GetBitcoinRPCClient() (*rpcclient.Client, error) {
	certs, err := LoadCerts()
	if err != nil {
		return nil, err
	}

	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         "localhost:8334",
		Endpoint:     "ws",
		User:         "4bmeiF7E3ny8cGf8Ok6QJZy/0pk=",
		Pass:         "2oljjSoRFzC5Go7hCGDID6xWi+c=",
		Certificates: certs,
	}, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
