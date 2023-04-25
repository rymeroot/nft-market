package nftimx

import (
	"context"
	"github.com/immutable/imx-core-sdk-golang/imx"
	"github.com/immutable/imx-core-sdk-golang/imx/api"
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
