package oracle

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/certusone/terra-oracle/pkg/signer"
	ctypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	tcontext "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/google/uuid"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
	"github.com/terra-project/core/app"
	"github.com/terra-project/core/x/oracle"
)

type PriceOracleConfig struct {
	Client        *client.HTTP
	ValAddress    ctypes.ValAddress
	ChainID       string
	TxFee         ctypes.Coin
	PriceProvider PriceProvider
	Signer        signer.Signer
}

type PriceOracle struct {
	cdc        *amino.Codec
	client     *client.HTTP
	valAddress ctypes.ValAddress
	chainID    string
	txFee      ctypes.Coin

	prevotes       []Vote
	prevotesPeriod int64

	priceProvider PriceProvider
	signer        signer.Signer
}

func NewPriceOracle(config PriceOracleConfig) *PriceOracle { // wsClient *client.HTTP, valAddress ctypes.ValAddress, chainID string, priceProvider PriceProvider, signer Signer, txFee ctypes.Coin) *PriceOracle {
	oracle := &PriceOracle{
		cdc:            app.MakeCodec(),
		client:         config.Client,
		valAddress:     config.ValAddress,
		chainID:        config.ChainID,
		txFee:          config.TxFee,
		prevotes:       []Vote{},
		prevotesPeriod: 0,
		priceProvider:  config.PriceProvider,
		signer:         config.Signer,
	}

	return oracle
}

func (priceOracle *PriceOracle) ProcessingLoop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*500)
	defer cancel()

	blocks, err := priceOracle.client.Subscribe(ctx, "oracle", types.QueryForEvent(types.EventNewBlock).String())
	if err != nil {
		return err
	}

	params, err := priceOracle.QueryOracleParams()
	if err != nil {
		return err
	}

	for event := range blocks {
		blockData := event.Data.(types.EventDataNewBlock)
		log.Printf("new block %d", blockData.Block.Height)

		blockHeight := blockData.Block.Height
		voteCycle := blockHeight / params.VotePeriod

		// if we send reveal messaage, increment offset to send prevotes as next sequence
		var offset uint64 = 0

		// first submit reveals for previous vote if we have previous vote
		if len(priceOracle.prevotes) > 0 && priceOracle.prevotesPeriod < voteCycle {
			revealMsgs := priceOracle.RevealVotes()

			if len(revealMsgs) > 0 {
				log.Printf("revealing votes for period: %d\n", priceOracle.prevotesPeriod)
				err := priceOracle.SendMessages(0, revealMsgs)
				if err != nil {
					log.Printf("error sending tx: %v", err)
				} else {
					priceOracle.prevotes = []Vote{}
					offset = 1
				}
			}
		}

		// TODO(roman) if the reveal failed, the prevotes will override existing prevotes
		// and we won't reveal - consider changing to keep prevotes until revealed
		// or until the period passes and there is no point in revealing

		// we have already submitted a prevote for this period
		if priceOracle.prevotesPeriod == voteCycle {
			continue
		}

		// if we are within 4 blocks of the end of the cycle, we don't prevote
		// this is a "workaround" for "ensuring" our prevotes are included within the proper
		// vote period and not the next vote period
		cycleLastHeight := (voteCycle * params.VotePeriod) + params.VotePeriod
		if blockHeight > (cycleLastHeight - 4) {
			continue
		}

		// generate prevotes
		votes, prevoteMsgs := priceOracle.SubmitVotes()
		if len(prevoteMsgs) > 0 {
			log.Printf("prevote for period: %d\n", voteCycle)
			err := priceOracle.SendMessages(offset, prevoteMsgs)
			if err != nil {
				log.Printf("error sending tx: %v", err)
			} else {
				priceOracle.prevotes = votes
				priceOracle.prevotesPeriod = voteCycle
			}
		}
	}

	return nil
}

func (priceOracle *PriceOracle) RevealVotes() []ctypes.Msg {
	var msgs []ctypes.Msg

	for _, vote := range priceOracle.prevotes {
		msg := oracle.NewMsgPriceVote(vote.Price, vote.Salt, vote.Denom, vote.Feeder, priceOracle.valAddress)
		msgs = append(msgs, msg)
	}

	return msgs
}

func (priceOracle *PriceOracle) SubmitVotes() ([]Vote, []ctypes.Msg) {
	actives, err := priceOracle.QueryActives()
	if err != nil {
		panic(err)
	}

	var (
		msgs  []ctypes.Msg
		votes []Vote
	)
	for _, denom := range actives {
		// truncate leading `u` from the denom for price lookup
		truncatedDenom := denom[1:]

		price, err := priceOracle.priceProvider.GetPrice(truncatedDenom)
		if err != nil {
			log.Printf("could not get %s price; err=%v", denom, err)
			continue
		}

		vote := Vote{
			Feeder:    priceOracle.signer.Address(),
			Denom:     denom,
			Salt:      uuid.New().String()[:3],
			Price:     price,
			Validator: priceOracle.valAddress,
		}

		log.Printf("voting for %s with price %s", denom, price.String())

		msg := oracle.NewMsgPricePrevote(vote.Hash(), vote.Denom, vote.Feeder, vote.Validator)
		msgs = append(msgs, msg)
		votes = append(votes, vote)
	}

	return votes, msgs
}

func (priceOracle *PriceOracle) SendMessages(sequenceOffset uint64, msgs []ctypes.Msg) error {
	tx := auth.StdTx{}
	tx.Msgs = msgs

	acc, err := priceOracle.QueryAccount(priceOracle.signer.Address())
	if err != nil {
		return err
	}

	fee := auth.NewStdFee(50000, ctypes.Coins{priceOracle.txFee})
	tx.Fee = fee

	signMsg := tcontext.StdSignMsg{
		ChainID:       priceOracle.chainID,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence() + sequenceOffset,
		Memo:          "",
		Msgs:          msgs,
		Fee:           fee,
	}
	signBytes := signMsg.Bytes()

	signature, err := priceOracle.signer.Sign(signBytes)
	if err != nil {
		return err
	}

	tx.Signatures = []auth.StdSignature{
		{
			Signature: signature,
			PubKey:    priceOracle.signer.PubKey(),
		},
	}

	txBytes, err := auth.DefaultTxEncoder(priceOracle.cdc)(tx)
	if err != nil {
		return err
	}

	res, err := priceOracle.client.BroadcastTxSync(txBytes)
	if err != nil {
		return err
	}

	if res.Code != 0 {
		return fmt.Errorf("error sending tx: %v", res)
	}

	log.Printf("submitted tx: %v\n", res)

	return nil
}
