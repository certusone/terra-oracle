package oracle

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

type Vote struct {
	Price     types.Dec
	Salt      string
	Denom     string
	Feeder    types.AccAddress
	Validator types.ValAddress
}

func (v Vote) Hash() string {
	preimage := fmt.Sprintf("%s:%s:%s:%s", v.Salt, v.Price, v.Denom, v.Validator)
	hash := tmhash.SumTruncated([]byte(preimage))

	return hex.EncodeToString(hash[:])
}
