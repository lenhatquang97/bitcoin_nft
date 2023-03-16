package initial

import (
	"fmt"

	"github.com/btcsuite/btcd/rpcclient"
)

func CreateConnection() *rpcclient.Client {
	// Set up connection parameters
	rpcConfig := &rpcclient.ConnConfig{
		Host:         "localhost:8332",
		User:         "rpcuser",
		Pass:         "rpcpassword",
		HTTPPostMode: true,
	}

	// Connect to the Bitcoin RPC server
	client, err := rpcclient.New(rpcConfig, nil)
	if err != nil {
		fmt.Println("Error connecting to Bitcoin RPC server:", err)
		return nil
	}
	return client
}
