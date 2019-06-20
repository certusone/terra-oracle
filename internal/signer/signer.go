package signer

import (
	ctypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

type Signer interface {
	Address() ctypes.AccAddress
	PubKey() crypto.PubKey
	Sign(bytes []byte) ([]byte, error)
}
