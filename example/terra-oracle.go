package main

import (
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/rpc/client"

	oracle "github.com/certusone/terra-oracle"
)

// examplePriceProvider implements the PriceProvider interface
// It returns the same price for all denominations.
type examplePriceProvider struct {
}

func NewExamplePriceProvider() (*examplePriceProvider, error) {
	return &examplePriceProvider{}, nil
}

func (provider *examplePriceProvider) GetPrice(denom string) (types.Dec, error) {
	return types.NewDecWithPrec(int64(10), 2), nil
}

func main() {
	oracle.Init()

	wsClient := client.NewHTTP(os.Getenv("RPC_HOST"), "/websocket")
	err := wsClient.Start()
	if err != nil {
		panic(err)
	}

	chainID := os.Getenv("CHAIN_ID")

	valAddressString := os.Getenv("VAL_ADDR")
	valAddress, err := types.ValAddressFromBech32(valAddressString)
	if err != nil {
		panic(err)
	}

	mneumonic := os.Getenv("MNEUMONIC")
	signer, err := oracle.NewHdSignerFromMneumonic(mneumonic)
	if err != nil {
		panic(err)
	}

	priceProvider, err := NewExamplePriceProvider()
	if err != nil {
		panic(err)
	}

	txFee := types.Coin{
		Denom:  "ukrw",
		Amount: types.NewInt(750),
	}

	oracle := oracle.NewPriceOracle(oracle.PriceOracleConfig{
		Client:        wsClient,
		ValAddress:    valAddress,
		ChainID:       chainID,
		PriceProvider: priceProvider,
		Signer:        signer,
		TxFee:         txFee,
	})

	log.Printf("Starting voter for:\n\tValidator: %s\n\tFeeder: %s\n\tChain: %s\n", valAddress.String(), signer.Address().String(), chainID)

	oracle.ProcessingLoop()
}
