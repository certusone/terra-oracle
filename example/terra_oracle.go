package main

import (
	"log"
	"os"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/certusone/terra-oracle/pkg/oracle"
	"github.com/certusone/terra-oracle/pkg/signer"
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

	mnemonic := os.Getenv("MNEMONIC")
	hdSigner, err := signer.NewHdSignerFromMnemonic(mnemonic)
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
		Signer:        hdSigner,
		TxFee:         txFee,
	})

	log.Printf("starting voter for:\n\tValidator: %s\n\tFeeder: %s\n\tChain: %s\n", valAddress.String(), hdSigner.Address().String(), chainID)

	// TODO(hendrik): Allow graceful stop with os.Signal
	log.Fatal(oracle.ProcessingLoop())
}
