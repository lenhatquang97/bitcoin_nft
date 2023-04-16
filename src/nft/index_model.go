package nft

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"go.mongodb.org/mongo-driver/mongo"
)

type UnspentOutputRange struct {
	Outpoint *wire.OutPoint
	Ranges   []int64
}

type Index struct {
	Client                          *rpcclient.Client
	Database                        *mongo.Database
	FirstInscriptionHeight          int64
	GenesisBlockCoinbaseTransaction *wire.MsgTx
	GenesisBlockCoinbaseTxID        string
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
	Wallet                 string
}
