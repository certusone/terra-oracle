package oracle

import (
	cosmos_types "github.com/cosmos/cosmos-sdk/types"
)

type PriceProvider interface {
	GetPrice(denom string) (cosmos_types.Dec, error)
}
