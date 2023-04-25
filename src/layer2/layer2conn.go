package layer2

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

var maxMsgRecvSize = grpc.MaxCallRecvMsgSize(1 * 1024 * 1024 * 200)

func ReadMacaroon(macPath string) (grpc.DialOption, error) {
	// Load the specified macaroon file.
	macBytes, err := ioutil.ReadFile(macPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read macaroon path : %v", err)
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macBytes); err != nil {
		return nil, fmt.Errorf("unable to decode macaroon: %v", err)
	}

	macConstraints := []macaroons.Constraint{
		macaroons.TimeoutConstraint(60),
	}

	// Apply constraints to the macaroon.
	constrainedMac, err := macaroons.AddConstraints(mac, macConstraints...)
	if err != nil {
		return nil, err
	}

	// Now we append the macaroon credentials to the dial options.
	cred, err := macaroons.NewMacaroonCredential(constrainedMac)
	if err != nil {
		return nil, fmt.Errorf("error creating macaroon credential: %v",
			err)
	}
	return grpc.WithPerRPCCredentials(cred), nil
}

// Done: Only need rpcUrl
func GetLndGrpcSetup() (*grpc.ClientConn, error) {
	lndDir := btcutil.AppDataDir("lnd", false)
	macaroonFileLocation := filepath.Join(lndDir, "/data/chain/bitcoin/testnet/admin.macaroon")
	tlsCertPath := filepath.Join(lndDir, "tls.cert")
	macOption, err := ReadMacaroon(macaroonFileLocation)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(maxMsgRecvSize),
		macOption,
	}

	// TLS cannot be disabled, we'll always have a cert file to read.
	creds, _ := credentials.NewClientTLSFromFile(tlsCertPath, "")
	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial("localhost:10009", opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
