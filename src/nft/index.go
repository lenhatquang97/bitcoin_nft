package nft

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"github.com/m25lab/bitcoin_nft/src/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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

type UnspentOutputRange struct {
	Outpoint *wire.OutPoint
	Ranges   []int64
}

type Index struct {
	Auth                            *Auth
	Client                          *rpcclient.Client
	Database                        *mongo.Database
	Path                            string // no use case
	FirstInscriptionHeight          int64
	GenesisBlockCoinbaseTransaction *wire.MsgTx
	GenesisBlockCoinbaseTxID        string
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

func Open(opt *Options) (*Index, error) {
	rpcUrl := GetRPCUrl(opt)
	if rpcUrl == "" {
		return nil, errors.New("RPC url is empty")
	}

	file := GetCookieFile(opt)
	if file == "" {
		return nil, errors.New("Cookie file is empty")
	}

	// log info
	auth, err := GetAuth(file)
	if err != nil {
		return nil, errors.New("Auth file is empty")
	}

	// note: web socket connection for btcd
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:       rpcUrl, // /ws or /wallet ?
		CookiePath: file,
	}, nil)

	if err != nil {
		return nil, err
	}

	dataDir := GetDataDir(opt)
	err = os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		return nil, err
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

		var res *StatisticToCount
		err = data.Decode(&res)
		if err != nil {
			return nil, err
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
		GenesisBlockCoinbaseTxID:        "0", // check how to get tx id, cal by hash + index
		Auth:                            auth,
		Client:                          client,
		Path:                            path, // no use case use this field
		FirstInscriptionHeight:          GetFirstInscriptionHeight(opt),
		HeightLimit:                     opt.HeightLimit,
		Reorged:                         &reorged,
		RpcUrl:                          rpcUrl,
	}, nil
}

func GetUnspentOutput(index *Index) (map[wire.OutPoint]btcutil.Amount, error) {
	utxos := make(map[wire.OutPoint]btcutil.Amount)
	// client list unspent
	client := index.Client
	unspentRes, err := client.ListUnspent()
	if err != nil {
		return nil, err
	}

	for txid, item := range unspentRes {
		txHash, err := chainhash.NewHashFromStr(item.TxID)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		utxos[wire.OutPoint{
			Hash:  *txHash,
			Index: uint32(txid),
		}] = btcutil.Amount(item.Amount)
	}

	listLockUnspent, err := client.ListLockUnspent()
	if err != nil {
		return nil, err
	}

	for _, item := range listLockUnspent {
		rawTx, err := client.GetRawTransaction(&item.Hash)
		if err != nil {
			return nil, err
		}
		utxos[wire.OutPoint{
			Hash:  item.Hash,
			Index: item.Index,
		}] = btcutil.Amount(rawTx.MsgTx().TxOut[item.Index].Value)
	}

	outpointToValue := index.Database.Collection(OUTPOINT_TO_VALUE)
	for outpoint := range utxos {
		filter := bson.M{}
		var key []byte
		txId := blockchain.HashToBig(&outpoint.Hash)
		key = append(key, txId.Bytes()...)
		key = append(key, utils.IntToBytes(int(outpoint.Index))...)
		filter["key"] = key
		data := outpointToValue.FindOne(context.TODO(), filter)
		if data.Err() != nil {
			return nil, data.Err()
		}

		var res OutPointToValue
		err = data.Decode(&res)
		if err != nil {
			return nil, err
		}
	}

	return utxos, nil
}

// maybe no use
func List(index *Index, outpoint *wire.OutPoint) []int64 {
	return []int64{}
}

// no use case use
func GetUnspentOutputRanges(index *Index) ([]*UnspentOutputRange, error) {
	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return nil, err
	}

	for _ = range unspentOutput {

	}

	return nil, nil
}

func HasSatIndex(index *Index) bool {
	outpointToSatRange := index.Database.Collection(OUTPOINT_TO_SAT_RANGES)
	if outpointToSatRange == nil {
		return false
	}

	return true
}

func RequirementSatIndex(index *Index) bool {
	return HasSatIndex(index)
}

// impl soon
func GetInfo(index *Index) *Info {
	return nil
}

// impl for server
func BlockTime() {

}

// impl soon (1)
func Update(index *Index) *Index {
	return index
}

func IsReorged(index *Index) *bool {
	return index.Reorged
}

// no use
func BeginRead() {

}

// no use
func BeginWrite() {

}

// impl soon (2)
func IncrementStatistic(index *Index) error {
	return nil
}

// test func
func Statistic(index *Index, statistic enum.StatisticValue) int64 {
	return 0
}

func GetHeightTx(index *Index) int64 {
	return 0
}

func BlockCount(index *Index) int64 {
	return 0
}

// impl soon (4)
func GetBlock(index *Index) {

}

// impl soon (4)
func RareSatSatPoints(index *Index) {

}

// impl soon (4)
func RateSatSatPoint(index *Index) {

}

// impl soon (4)
func BlockHeader(index *Index) {

}

// impl soon (2)
func GetBlockByHeight(index *Index) {

}

// impl soon (4)
func GetBlockByHash(index *Index, height int64) (*wire.MsgBlock, error) {
	if index.Client == nil {
		return nil, errors.New("Client is nil")
	}

	res, err := index.Client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}

	data, err := index.Client.GetBlock(res)
	return data, err
}

// impl soon (4)
func GetInscriptionIdBySat(index *Index) {

}

// impl soon (4)
func GetInscriptionIdByInscriptionNumber(index *Index) {

}

// impl soon (3)
func GetInscriptionSatPointById(index *Index, inscriptionId *src.InscriptionId) (*src.SatPoint, error) {
	if index.Database == nil {
		return nil, errors.New("Database client is nil")
	}

	inscriptionIdToSatPoint := index.Database.Collection(INSCRIPTION_ID_TO_SATPOINT)
	inscriptionIdStore, err := src.GetInscriptionIDStore(inscriptionId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{}
	filter["key"] = inscriptionIdStore
	data := inscriptionIdToSatPoint.FindOne(context.TODO(), filter)
	if data.Err() != nil {
		return nil, data.Err()
	}
	var res *InscriptionIDToSatPoint
	err = data.Decode(&res)
	if err != nil {
		return nil, err
	}

	satPointDecode, err := src.LoadIntoSatPoint(res.Value)
	if err != nil {
		return nil, err
	}

	return satPointDecode, nil
}

// impl soon (3)
func GetInscriptionById(index *Index, inscriptionId *src.InscriptionId) (*Inscription, error) {
	if index.Database == nil {
		return nil, errors.New("Client data is nil")
	}

	inscriptionIdToSatPoint := index.Database.Collection(INSCRIPTION_ID_TO_SATPOINT)
	inscriptionIdStore, err := src.GetInscriptionIDStore(inscriptionId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{}
	filter["key"] = inscriptionIdStore
	data := inscriptionIdToSatPoint.FindOne(context.TODO(), filter)
	if data.Err() != nil {
		return nil, data.Err()
	}
	var res *InscriptionIDToSatPoint
	err = data.Decode(&res)
	if err != nil {
		return nil, err
	}

	// oke data is nil
	if res == nil {
		return nil, nil
	}

	tx, err := GetTransaction(index, inscriptionId.TxID)
	if err != nil {
		return nil, err
	}

	return NftFromTransaction(tx)
}

// impl soon (3)
func GetInscriptionOnOutput(index *Index, outpoint *wire.OutPoint) ([]src.InscriptionId, error) {
	if index.Database == nil {
		return nil, errors.New("Database is nil")
	}

	lower, err := src.GetSatPointStore(&src.SatPoint{
		OutPoint: *outpoint,
		OffSet:   0,
	})

	higher, err := src.GetSatPointStore(&src.SatPoint{
		OutPoint: *outpoint,
		OffSet:   math.MaxInt64,
	})

	satpointToInscriptionId := index.Database.Collection(SAT_TO_INSCRIPTION_ID)
	filter := bson.M{}
	filter["key"] = bson.M{
		"$gte": lower,
		"$lte": higher,
	}

	cursor, err := satpointToInscriptionId.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	satPointMap := make(map[src.SatPoint]src.InscriptionId)
	for cursor.Next(context.TODO()) {
		var res *SatPointToInscriptionID
		err = cursor.Decode(&res)
		if err != nil {
			return nil, err
		}
		parseSatPoint, err := src.LoadIntoSatPoint(res.Key)
		if err != nil {
			return nil, err
		}

		parseInscriptionID, err := src.LoadIntoInscriptionID(res.Value)
		if err != nil {
			return nil, err
		}

		satPointMap[*parseSatPoint] = *parseInscriptionID
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	var res []src.InscriptionId
	for _, inscriptionId := range satPointMap {
		res = append(res, inscriptionId)
	}

	cursor.Close(context.TODO())

	return res, nil
}

// impl soon (2)
func GetTransaction(index *Index, txId string) (*wire.MsgTx, error) {
	if txId == index.GenesisBlockCoinbaseTxID {
		return index.GenesisBlockCoinbaseTransaction, nil
	} else {
		txHash, err := chainhash.NewHashFromStr(txId)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		tx, err := index.Client.GetRawTransaction(txHash)
		if err != nil {
			return nil, err
		}
		return tx.MsgTx(), nil
	}
}

// impl soon (4)
func GetTransactionBlockHash(index *Index) {

}

// maybe no use
func IsTransactionInActiveChain(index *Index) {

}

// impl soon (4): maybe no use
func Find(index *Index, sat int64) (*src.SatPoint, error) {
	if !RequirementSatIndex(index) {
		return nil, errors.New("requires index created -- find")
	}

	return nil, nil
}

// maybe no use
func ListInner(index *Index) {

}

// impl soon (4)
func GetBlockTime(index *Index) {

}

// impl soon (1)
func GetInscription(index *Index) (map[src.SatPoint]src.InscriptionId, error) {
	satPointToInscriptionId := index.Database.Collection(SAT_TO_INSCRIPTION_ID)
	if satPointToInscriptionId == nil {
		return nil, errors.New("collection SAT_TO_INSCRIPTION_ID is null")
	}

	cursor, err := satPointToInscriptionId.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	satPointMap := make(map[src.SatPoint]src.InscriptionId)
	for cursor.Next(context.TODO()) {
		var res *SatPointToInscriptionID
		err = cursor.Decode(&res)
		if err != nil {
			return nil, err
		}
		parseSatPoint, err := src.LoadIntoSatPoint(res.Key)
		if err != nil {
			return nil, err
		}

		parseInscriptionID, err := src.LoadIntoInscriptionID(res.Value)
		if err != nil {
			return nil, err
		}

		satPointMap[*parseSatPoint] = *parseInscriptionID
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	cursor.Close(context.TODO())

	return satPointMap, nil
}

// impl soon (4)
func GetHomePageInscription(index *Index) {

}

// impl soon (4)
func GetLatestInscriptionWithPrevAndNext(index *Index) {

}

// impl soon (4)
func GetFeeInscription(index *Index) {

}

// impl soon (4)
func GetInscriptionEntry(index *Index) {

}

// impl soon (4)
func AssertInscriptionLocation(index *Index) {

}

// impl soon (3): maybe no use
// just is a range query in mongo
func InscriptionOnOutput(index *Index) {
	//index.Client.ImportAddress()
}
