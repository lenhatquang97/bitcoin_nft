package nft

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"math"
	"os"
	"strings"
)

// collection
const (
	HEIGHT_TO_BLOCK_HASH                                = "height_to_block_hash"
	INSCRIPTION_ID_TO_INSCRIPTION_ENTRY                 = "inscription_id_to_inscription_entry"
	INSCRIPTION_ID_TO_SATPOINT                          = "inscription_id_to_satpoint"
	INSCRIPTION_NUMBER_TO_INSCRIPTION_ID                = "inscription_number_to_inscription_id"
	OUTPOINT_TO_SAT_RANGES                              = "outpoint_to_sat_ranges"
	OUTPOINT_TO_VALUE                                   = "outpoint_to_value"
	SATPOINT_TO_INSCRIPTION_ID                          = "satpoint_to_inscription_id"
	SAT_TO_INSCRIPTION_ID                               = "sat_to_inscription_id"
	SAT_TO_SATPOINT                                     = "sat_to_satpoint"
	STATISTIC_TO_COUNT                                  = "statistic_to_account"
	WRITE_TRANSACTION_STARTING_BLOCK_COUNT_TO_TIMESTAMP = "write_transaction_starting_block_to_timestamp"

	SCHEMA_VERSION = 3
)

type Auth struct {
	UserName string
	Password string
}

type Index struct {
	Auth                            *Auth
	Client                          *rpcclient.Client
	Database                        *mongo.Database
	Path                            string // no use case
	FirstInscriptionHeight          int64
	GenesisBlockCoinbaseTransaction *wire.MsgTx
	GenesisBlockCoinbaseTxID        int32
	HeightLimit                     int64
	Reorged                         *bool
	RpcUrl                          string
}

type Info struct {
	BlockIndexed    int64
	BranchPages     int64
	FragmentBytes   int64
	IndexFileSize   int64  // no use
	IndexPath       string // no use
	LeafPage        int64
	MetaDataBytes   int64
	OutputTraversed int64
	PageSize        int64
	SatRange        int64
	StoredBytes     int64
	Transactions    []TransactionInfo
	TreeWeight      int64
	UtxoIndex       int64
}

type TransactionInfo struct {
	StartingBlockCount int64
	StartingTimeTemp   int64
}

type Options struct {
	BitcoinDataDir         string
	ChainArgument          enum.ChainValue
	Config                 string
	ConfigDir              string
	CookieFile             string
	DataDir                string
	FirstInscriptionHeight int64
	HeightLimit            int64
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

func GetAuth(cookieFile string) (*Auth, error) {
	filerc, err := os.Open(cookieFile)
	if err != nil {
		return nil, err
	}
	defer filerc.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(filerc)
	contents := buf.String()

	userInfo := strings.Split(contents, ":")
	return &Auth{
		UserName: userInfo[0],
		Password: userInfo[1],
	}, nil
}

func Open(opt *Options) *Index {
	rpcUrl := GetRPCUrl(opt)
	if rpcUrl == "" {
		return nil
	}

	file := GetCookieFile(opt)
	if file == "" {
		return nil
	}

	// log info
	auth, err := GetAuth(file)
	if err != nil {
		return nil
	}

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

	fmt.Println(path)

	ctx := context.TODO() // init context global

	uriConn := "mongodb+srv://tuankiet:kietlu1712@bankaccount.lfuju.mongodb.net/?retryWrites=true&w=majority"
	option := options.Client().ApplyURI(uriConn)
	mongoclient, err := mongo.Connect(ctx, option)
	if err != nil {
		log.Fatal("error while connecting with mongo", err)
	}

	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("error while trying to ping mongo", err)
	}

	database := mongoclient.Database("ordinal")
	collection := database.Collection(STATISTIC_TO_COUNT)
	filter := bson.M{}
	filter["key"] = enum.Statistic.Schema
	data := collection.FindOne(ctx, filter)
	if data != nil {

		var res *StatisticToAccount
		err = data.Decode(&res)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if res.Value < SCHEMA_VERSION {
			// print info
		} else if res.Value > SCHEMA_VERSION {
			// print info
		}
	} else {
		// insert version

		// insert empty value
	}

	chaincfgParam := GetChainInfo(opt)

	// get genesis block coin base tx

	reorged := false
	return &Index{
		GenesisBlockCoinbaseTransaction: chaincfgParam.GenesisBlock.Transactions[0],
		GenesisBlockCoinbaseTxID:        0,
		Auth:                            auth,
		Client:                          client,
		Path:                            path, // no use case use this field
		FirstInscriptionHeight:          GetFirstInscriptionHeight(opt),
		HeightLimit:                     opt.HeightLimit,
		Reorged:                         &reorged,
		RpcUrl:                          rpcUrl,
	}
}
