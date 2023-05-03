package nftimx

import (
	"context"
	"github.com/immutable/imx-core-sdk-golang/imx"
	"github.com/immutable/imx-core-sdk-golang/imx/api"
	"github.com/immutable/imx-core-sdk-golang/imx/signers/ethereum"
	"log"
)

func Connect() (context.Context, imx.Config, *imx.Client) {
	ctx := context.TODO()
	imxConfig := api.NewConfiguration()
	cfg := imx.Config{
		APIConfig:     imxConfig,
		AlchemyAPIKey: "WmzAboIrYGOEDnuUJnxQ8ucuu3jfoR_q",
		Environment:   imx.Sandbox,
	}

	imxClient, err := imx.NewClient(&cfg)
	if err != nil {
		log.Printf("failed to create imx client: %v\n", err)
		return nil, cfg, nil
	}

	return ctx, cfg, imxClient
}

func Mint(userPrivateKey string, userAddress string, contractAddress string, tokenID string, tokenMetadata string) int32 {
	return 0
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)

	var royaltyPercentage float32 = 10
	var newToken = imx.UnsignedMintRequest{
		ContractAddress: contractAddress,
		Royalties: []imx.MintFee{
			{
				Percentage: royaltyPercentage,
				Recipient:  userAddress,
			},
		},
		Users: []imx.User{
			{
				User: userAddress,
				Tokens: []imx.MintableTokenData{
					{
						ID: tokenID,
						Royalties: []imx.MintFee{
							{
								Percentage: royaltyPercentage,
								Recipient:  userAddress,
							},
						},
						Blueprint: &tokenMetadata,
					},
				},
			},
		},
	}

	req := make([]imx.UnsignedMintRequest, 1)
	req[0] = newToken

	imxres, err := imxClient.Mint(ctx, l1signer, req)
	if err != nil {
		log.Printf("error in minting.MintTokensWorkflow: %v\n", err)
	}

	res := imxres.GetResults()
	return res[0].GetTxId()
}
