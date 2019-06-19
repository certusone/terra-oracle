module github.com/certusone/terra-oracle

require (
	github.com/cosmos/cosmos-sdk v0.34.7
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/google/uuid v1.1.1
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/tendermint v0.31.5
	github.com/terra-project/core v0.2.1
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/YunSuk-Yeo/cosmos-sdk v0.34.7-terra
	golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
)
