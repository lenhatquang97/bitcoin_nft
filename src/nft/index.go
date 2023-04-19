package nft

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
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
	SCHEMA_VERSION                                      = 3
)

var ctx context.Context

// Done: load certs
func LoadCerts() ([]byte, error) {
	certHomeDir := btcutil.AppDataDir("btcwallet", false)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return nil, err
	}
	return certs, nil
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

		return nil, errors.New("file doesn't exists")
	}
}

// Done: Only need rpcUrl
func GetBitcoinRPCClient(opt *Options) (*rpcclient.Client, error) {
	certs, err := LoadCerts()
	if err != nil {
		return nil, err
	}

	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         opt.RpcUrl,
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

// Done: Only need rpcUrl and walletName
func GetBitcoinRPCClientForWalletCommand(opt *Options, create bool) (*rpcclient.Client, error) {
	client, err := GetBitcoinRPCClient(opt)
	if err != nil {
		return nil, err
	}

	if !create {
		result, err := client.LoadWallet(opt.Wallet)
		if err != nil {
			return nil, err
		}
		fmt.Println(result)
	}

	return client, nil
}

// Need to improve
func Open(opt *Options) (*Index, error) {
	//Step 1: Connect RPC
	client, err := GetBitcoinRPCClient(opt)
	if err != nil {
		return nil, err
	}

	//Step 2: Connect MongoDB
	ctx = context.TODO() // init context global
	uriConn := "mongodb+srv://tuankiet:kietlu1712@bankaccount.lfuju.mongodb.net/?retryWrites=true&w=majority"
	option := options.Client().ApplyURI(uriConn)
	mongoclient, err := mongo.Connect(ctx, option)
	if err != nil {
		return nil, err
	}
	err = mongoclient.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	//Step 3: Get height whether can connect to BTCD or not?
	height, err := client.GetBlockCount()
	if err != nil {
		return nil, err
	}

	//TODO: Will add database collection after
	return &Index{
		GenesisBlockCoinbaseTransaction: chaincfg.TestNet3Params.GenesisBlock.Transactions[0],
		GenesisBlockCoinbaseTxID:        "0",
		Client:                          client,
		FirstInscriptionHeight:          height,
		RpcUrl:                          opt.RpcUrl,
	}, nil
}

func GetUnspentOutput(index *Index) (map[wire.OutPoint]btcutil.Amount, error) {
	utxos := make(map[wire.OutPoint]btcutil.Amount)
	// client list unspent
	unspentRes, err := index.Client.ListUnspent()
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

	listLockUnspent, err := index.Client.ListLockUnspent()
	if err != nil {
		return nil, err
	}

	for _, item := range listLockUnspent {
		rawTx, err := index.Client.GetRawTransaction(&item.Hash)
		if err != nil {
			return nil, err
		}
		utxos[wire.OutPoint{
			Hash:  item.Hash,
			Index: item.Index,
		}] = btcutil.Amount(rawTx.MsgTx().TxOut[item.Index].Value)
	}

	// outpointToValue := index.Database.Collection(OUTPOINT_TO_VALUE)
	// for outpoint := range utxos {
	// 	filter := bson.M{}
	// 	var key []byte
	// 	txId := blockchain.HashToBig(&outpoint.Hash)
	// 	key = append(key, txId.Bytes()...)
	// 	key = append(key, utils.IntToBytes(int(outpoint.Index))...)
	// 	filter["key"] = key
	// 	data := outpointToValue.FindOne(context.TODO(), filter)
	// 	if data.Err() != nil {
	// 		return nil, data.Err()
	// 	}

	// 	var res OutPointToValue
	// 	err = data.Decode(&res)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return utxos, nil
}

// **
func List(index *Index, outpoint *wire.OutPoint) []int64 {
	return []int64{}
}

func HasSatIndex(index *Index) bool {
	outpointToSatRange := index.Database.Collection(OUTPOINT_TO_SAT_RANGES)
	return outpointToSatRange != nil
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
func GetBlock(index *Index) []chainhash.Hash {
	return nil
}

// impl soon (4)
func RareSatSatPoints(index *Index) {

}

// impl soon (4)
func RateSatSatPoint(index *Index) {

}

// impl soon (4)
func BlockHeader(index *Index, hash *chainhash.Hash) (*wire.BlockHeader, error) {
	client := index.Client
	if client == nil {
		panic("Client is nil")
	}

	res, err := client.GetBlockHeader(hash)
	if err != nil {
		return nil, err
	}

	return res, err
}

// impl soon (2)
func GetBlockByHeight(index *Index, height int64) (*wire.MsgBlock, error) {
	client := index.Client
	if client == nil {
		panic("Client is nil")
	}

	blockHash, err := client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}

	block, err := client.GetBlock(blockHash)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// impl soon (4)
func GetBlockByHash(index *Index, hash *chainhash.Hash) (*wire.MsgBlock, error) {
	if index.Client == nil {
		return nil, errors.New("Client is nil")
	}

	data, err := index.Client.GetBlock(hash)
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

// impl soon (2)
func GetTransactionBlockHash(index *Index, txId string) (*chainhash.Hash, error) {
	client := index.Client
	if client == nil {
		panic("Client is nil")
	}

	// convert txid to tx hash
	txHash, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return nil, err
	}

	res, err := client.GetRawTransaction(txHash)
	if err != nil {
		return nil, err
	}

	return res.Hash(), err
}

// maybe no use
func IsTransactionInActiveChain(index *Index) {

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
