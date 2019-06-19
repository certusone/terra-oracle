package oracle

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	cosmos_types "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type HdSigner struct {
	privKey secp256k1.PrivKeySecp256k1
	address cosmos_types.AccAddress
}

func NewHdSignerFromMneumonic(mneumonic string) (*HdSigner, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mneumonic, "")
	if err != nil {
		return nil, err
	}
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	params := hd.NewFundraiserParams(0, 0)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, params.String())
	privKey := secp256k1.PrivKeySecp256k1(derivedPriv)

	return &HdSigner{
		privKey: privKey,
		address: cosmos_types.AccAddress(privKey.PubKey().Address()),
	}, nil
}

func (signer *HdSigner) Address() cosmos_types.AccAddress {
	return signer.address
}

func (signer *HdSigner) PubKey() crypto.PubKey {
	return signer.privKey.PubKey()
}

func (signer *HdSigner) Sign(bytes []byte) ([]byte, error) {
	return signer.privKey.Sign(bytes)
}
