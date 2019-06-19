package oracle

import (
	"github.com/tendermint/tendermint/crypto"
	cosmos_types "github.com/cosmos/cosmos-sdk/types"
)

type PriceProvider interface {
	GetPrice(denom string) (cosmos_types.Dec, error)
}

type Signer interface {
	Address() cosmos_types.AccAddress
	PubKey() crypto.PubKey
	Sign(bytes []byte) ([]byte, error)
}
