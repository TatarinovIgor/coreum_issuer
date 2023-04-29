package side_functions

import (
	"context"
	"fmt"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Config struct {
	GRPCAddress     string
	NetworkConfig   config.NetworkConfig
	FundingMnemonic string
	StakerMnemonics []string
	LogFormat       logger.Format
	LogVerbose      bool
	RunUnsafe       bool
}

var (
	cfg   Config
	chain Chain
)

type stringsFlag []string

func (m *stringsFlag) String() string {
	if len(*m) == 0 {
		return ""
	}
	return fmt.Sprint(*m)
}

func (m *stringsFlag) Set(val string) error {
	*m = append(*m, val)
	return nil
}

func InitVariables() (Config, Chain, error) {
	var (
		fundingMnemonic, coredAddress, logFormat string
		chainID                                  string
		stakerMnemonics                          stringsFlag
		runUnsafe                                bool
	)

	coredAddress = "localhost:9090"
	fundingMnemonic = "pitch basic bundle cause toe sound warm love town crucial divorce shell olympic convince scene middle garment glimpse narrow during fix fruit suffer honey"
	logFormat = string(logger.ToolDefaultConfig.Format)
	chainID = string(constant.ChainIDDev)
	runUnsafe = true //Todo change for production

	// set the default staker mnemonic used in the dev znet by default
	if len(stakerMnemonics) == 0 {
		stakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
			"enemy fix tribe swift alcohol metal salad edge episode dry tired address bless cloth error useful define rough fold swift confirm century wasp acoustic",
			"act electric demand cancel duck invest below once obvious estate interest solution drink mango reason already clean host limit stadium smoke census pattern express",
		}
	}

	networkConfig, err := NewNetworkConfig(constant.ChainID(chainID))
	if err != nil {
		panic(fmt.Sprintf("can't create network config for the integration tests: %s", err))
	}
	cfg = Config{
		GRPCAddress:     coredAddress,
		NetworkConfig:   networkConfig,
		FundingMnemonic: fundingMnemonic,
		StakerMnemonics: stakerMnemonics,
		LogFormat:       logger.Format(logFormat),
		LogVerbose:      true,
		RunUnsafe:       runUnsafe,
	}

	chain = NewChain(ChainConfig{
		GRPCAddress:     cfg.GRPCAddress,
		NetworkConfig:   cfg.NetworkConfig,
		FundingMnemonic: cfg.FundingMnemonic,
		StakerMnemonics: cfg.StakerMnemonics,
	})
	return cfg, chain, nil
}

func InitChain() Chain {
	_, chain, err := InitVariables()
	if err != nil {
		return Chain{}
	}
	return chain
}

func AssetNFTIssue() error {
	chain := InitChain()

	loggerConfig := logger.Config{
		Format:  cfg.LogFormat,
		Verbose: cfg.LogVerbose,
	}

	ctx, cancel := context.WithCancel(logger.WithLogger(context.Background(), logger.New(loggerConfig)))
	defer cancel()

	issuer := chain.GenAccount()
	//Test only
	err := chain.Faucet.FundAccountsWithOptions(ctx, issuer, BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
		},
	})
	if err != nil {
		return err
	}
	assetNftClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	// issue new NFT class with too long data

	data, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: []byte("Hello world!")})

	if err != nil {
		return err
	}

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "BHT",
		Name:        "BH Token",
		Description: "BirdHouse Token",
		URI:         "https://bird-house.link/",
		URIHash:     "content-hash",
		Data:        data,
	}
	fmt.Printf(issuer.String())

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer), //Error here
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	if err != nil {
		return err
	}

	jsonData := []byte(`{"name": "Name", "description": "Description"}`)

	// issue new NFT class
	data, err = codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: jsonData})
	if err != nil {
		return err
	}

	issueMsg = &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		Data:        data,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_disable_sending,
		},
		RoyaltyRate: sdk.MustNewDecFromStr("0.1"),
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	if err != nil {
		return err
	}

	//requireT.Equal(chain.GasLimitByMsgs(issueMsg), uint64(res.GasUsed))
	tokenIssuedEvents, err := event.FindTypedEvents[*assetnfttypes.EventClassIssued](res.Events)

	if err != nil {
		return err
	}

	issuedEvent := tokenIssuedEvents[0]
	fmt.Println(issuedEvent.ID)
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)

	// query nft asset with features
	assetNftClassRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: classID,
	})

	fmt.Printf(assetNftClassRes.String())

	//requireT.Equal(jsonData, data2.Data)
	return nil
}
