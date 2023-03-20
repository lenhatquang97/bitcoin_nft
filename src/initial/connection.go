package initial

import (
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/m25lab/bitcoin_nft/configs"
)

func CreateConnection() *rpcclient.Client {
	loadedConfig := configs.ParseConfig()
	rpcConfig := &rpcclient.ConnConfig{
		Host:                 fmt.Sprintf("localhost:%d", loadedConfig.RpcPort),
		Endpoint:             loadedConfig.Endpoint,
		User:                 loadedConfig.RpcUser,
		Pass:                 loadedConfig.RpcPass,
		DisableAutoReconnect: false,
		DisableConnectOnNew:  true,
		DisableTLS:           true,
	}

	// Connect to the Bitcoin RPC server
	client, err := rpcclient.New(rpcConfig, nil)
	if err != nil {
		fmt.Println("Error connecting to Bitcoin RPC server:", err)
		return nil
	}
	return client
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
