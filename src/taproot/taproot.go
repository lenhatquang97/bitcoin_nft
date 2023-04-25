package taproot

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
	"github.com/m25lab/bitcoin_nft/src/script"
)

func GenerateTaprootKey(
	client walletrpc.WalletKitClient,
	taprootKeyFamily int32,
	internalKey *secp256k1.PublicKey,
	preimage []byte,
) (*secp256k1.PublicKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	keyReq := walletrpc.KeyReq{KeyFamily: taprootKeyFamily}
	keyDesc, err := client.DeriveNextKey(ctx, &keyReq)
	if err != nil {
		return nil, err
	}

	leafSigningKey, _ := btcec.ParsePubKey(keyDesc.RawKeyBytes)
	leaf1 := script.HashLockScriptLeaf(preimage)
	leaf2 := script.SchnorrSigScriptLeaf(leafSigningKey)

	inclusionProof := leaf1.TapHash()
	tapscript := input.TapscriptPartialReveal(internalKey, leaf2, inclusionProof[:])
	taprootKey, _ := tapscript.TaprootKey()
	return taprootKey, nil
}
