package nft

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"math"
	"os"
)

type Options struct {
	BitcoinDataDir         string
	ChainArgument          enum.ChainValue
	Config                 string
	ConfigDir              string
	CookieFile             string
	DataDir                string
	FirstInscriptionHeight int64
	Height                 int64
	Index                  string
	IndexSats              bool
	RegTest                bool
	RpcUrl                 string
	SigNet                 bool
	TestNet                bool
	Wallet                 string
}

func GetChainInfo(opt *Options) *chaincfg.Params {
	if opt.SigNet {
		return &chaincfg.SigNetParams
	} else if opt.RegTest {
		return &chaincfg.RegressionNetParams
	} else if opt.TestNet {
		return &chaincfg.TestNet3Params
	} else {
		return &chaincfg.MainNetParams
	}
}

func Chain(opt *Options) enum.ChainValue {
	if opt.SigNet {
		return enum.Chain.Signet
	} else if opt.RegTest {
		return enum.Chain.RegTest
	} else if opt.TestNet {
		return enum.Chain.Testnet
	} else {
		return enum.Chain.Bitcoin
	}
}

func GetFirstInscriptionHeight(opt *Options) int64 {
	if opt.RegTest {
		return int64(math.Max(float64(opt.FirstInscriptionHeight), 0))
	} else {
		if opt.FirstInscriptionHeight > 0 {
			return opt.FirstInscriptionHeight
		}

		return src.GetFirstInscriptionHeight(opt.ChainArgument)
	}
}

func GetRPCUrl(opt *Options) string {
	// check format by regress
	s := fmt.Sprintf("127.0.0.1:%d/wallet/%s", src.GetDefaultRPCPort(opt.ChainArgument), opt.Wallet)
	if opt.RpcUrl != "" {
		return opt.RpcUrl
	}
	return s
}

func GetCookieFile(opt *Options) string {
	if opt.CookieFile != "" {
		return opt.CookieFile
	}

	path := ""
	if opt.BitcoinDataDir != "" {
		path = opt.BitcoinDataDir
	} else {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		path = dirname + ".bitcoin"
	}

	return src.JoinWithDataDir(path, opt.ChainArgument) + ".cookie"
}

func GetDataDir(opt *Options) string {
	path := ""
	if opt.DataDir != "" {
		path = opt.DataDir
	} else {
		dirname, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		path = dirname
	}

	return src.JoinWithDataDir(path, opt.ChainArgument)
}

func LoadConfig(opt *Options) (*os.File, error) {
	if opt.Config != "" {
		data, err := os.Open(opt.Config)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else {
		if opt.ConfigDir != "" {
			data, err := os.Open(opt.Config + "ord.yaml")
			if err != nil {
				return nil, err
			}
			return data, nil
		}

		return nil, errors.New("File doesn't exists")
	}
}

func FormatBitcoinCoreVersion(version int64) string {
	return fmt.Sprintf("%d.%d.%d", version/10000, version%10000/100, version%1000)
}

func GetBitcoinRPCClient(opt *Options) (*rpcclient.Client, error) {
	cookieFile := GetCookieFile(opt)
	if cookieFile == "" {
		return nil, errors.New("Cookie file was not found")
	}

	rpcUrl := GetRPCUrl(opt)
	if rpcUrl == "" {
		return nil, errors.New("Rpc url is empty")
	}

	// log info

	// note: web socket connection for btcd
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:       rpcUrl, // /ws or /wallet ?
		CookiePath: cookieFile,
	}, nil)

	data, err := client.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}

	chain := Chain(&Options{ChainArgument: enum.ChainValue(data.Chain)})
	if chain != opt.ChainArgument {
		// panic err
	}

	return client, nil
}

func GetBitcoinRPCClientForWalletCommand(opt *Options, create bool) (*rpcclient.Client, error) {
	client, err := GetBitcoinRPCClient(opt)
	if err != nil {
		return nil, err
	}

	var minVersion int32 = 240000
	bitcoinVersion, err := client.GetNetworkInfo()
	if err != nil {
		return nil, err
	}

	if bitcoinVersion.Version < minVersion {
		s := fmt.Sprintf("Bitcoin Core %d or newer required, current version is %d", minVersion, bitcoinVersion.Version)
		return nil, errors.New(s)
	}

	if !create {
		_, _ = client.LoadWallet(opt.Wallet)

	}

	return client, nil
}

func Open(opt *Options) *Options {
	rpcUrl := GetRPCUrl(opt)
	if rpcUrl == "" {
		return nil
	}

	file := GetCookieFile(opt)
	if file == "" {
		return nil
	}

	// log info

	// note: web socket connection for btcd
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:       rpcUrl, // /ws or /wallet ?
		CookiePath: file,
	}, nil)

	if err != nil {
		return nil
	}

	dataDir := GetDataDir(opt)
	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		return nil
	}

	path := ""
	if opt.Index != "" {
		path = opt.Index
	} else {
		path = dataDir + "index.redb"
	}

}
