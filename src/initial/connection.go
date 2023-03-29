package initial

import (
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/m25lab/bitcoin_nft/configs"
)

type BtcConnection struct {
	Host, RpcUser, RpcPass   string
	HTTPPostMode, DisableTLS bool
}

func (conn *BtcConnection) CreateConnection() *rpcclient.Client {
	rpcConfig := &rpcclient.ConnConfig{
		Host:         conn.Host,
		User:         conn.RpcUser,
		Pass:         conn.RpcPass,
		HTTPPostMode: conn.HTTPPostMode,
		DisableTLS:   conn.DisableTLS,
	}

	// Connect to the Bitcoin RPC server
	client, err := rpcclient.New(rpcConfig, nil)
	if err != nil {
		fmt.Println("Error connecting to Bitcoin RPC server:", err)
		return nil
	}
	return client
}

func InitConnectionFromScript() *BtcConnection {
	loadedConfig := configs.ParseConfig()
	return &BtcConnection{
		Host:         fmt.Sprintf("localhost:%d", loadedConfig.RpcPort),
		RpcUser:      loadedConfig.RpcUser,
		RpcPass:      loadedConfig.RpcPass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
}

func ReadCAFile() []byte {
	//Need to make sure that you have disabled client TLS
	var certs []byte
	var err error
	certs, err = os.ReadFile(configs.GetRpcCert())
	if err != nil {
		log.Printf("Cannot open CA file: %v", err)
		certs = nil
	}
	return certs
}
