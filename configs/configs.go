package configs

import (
	"log"
	"os"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

type Config struct {
	RpcUser  string `env:"RPCUSER,required"`
	RpcPass  string `env:"RPCPASS,required"`
	Endpoint string `env:"ENDPOINT,required"`
	RpcPort  int    `env:"RPCPORT,required"`
}

func ParseConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := Config{}
	err = env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Unable to parse environment variables: %v", err)
	}
	return cfg
}
func GetRpcCert() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return os.Getenv("RPCCERT")
}
