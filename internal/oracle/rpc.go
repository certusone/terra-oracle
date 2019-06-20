package oracle

import (
	"fmt"

	ctypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/terra-project/core/x/oracle"
	"github.com/terra-project/core/x/oracle/client/cli"
)

func (priceOracle *PriceOracle) QueryAccount(addr ctypes.AccAddress) (auth.Account, error) {
	bz, err := priceOracle.cdc.MarshalJSON(auth.NewQueryAccountParams(addr))
	if err != nil {
		return nil, err
	}

	res, err := priceOracle.client.ABCIQuery(fmt.Sprintf("custom/%s/%s", auth.StoreKey, auth.QueryAccount), bz)
	if err != nil {
		return nil, err
	}

	var account auth.Account
	err = priceOracle.cdc.UnmarshalJSON(res.Response.Value, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (priceOracle *PriceOracle) QueryActives() (cli.DenomList, error) {
	res, err := priceOracle.client.ABCIQuery(fmt.Sprintf("custom/%s/%s", "oracle", oracle.QueryActive), nil)
	if err != nil {
		return nil, err
	}

	var actives cli.DenomList
	err = priceOracle.cdc.UnmarshalJSON(res.Response.Value, &actives)
	if err != nil {
		return nil, err
	}

	return actives, nil
}

func (priceOracle *PriceOracle) QueryOracleParams() (*oracle.Params, error) {
	res, err := priceOracle.client.ABCIQuery(fmt.Sprintf("custom/%s/%s", "oracle", oracle.QueryParams), nil)
	if err != nil {
		return nil, err
	}

	var params oracle.Params
	err = priceOracle.cdc.UnmarshalJSON(res.Response.Value, &params)
	if err != nil {
		return nil, err
	}

	return &params, nil
}
