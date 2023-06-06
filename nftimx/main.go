package nftimx

import (
	"context"
	"github.com/immutable/imx-core-sdk-golang/imx"
	"github.com/immutable/imx-core-sdk-golang/imx/api"
	"github.com/immutable/imx-core-sdk-golang/imx/signers/ethereum"
	"github.com/immutable/imx-core-sdk-golang/imx/signers/stark"
	"log"
	"math/big"
	"strconv"
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

func Register(userPrivateKey string, l2signer *stark.Signer, email string) string {
	return ""
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return ""
	}

	imxres, err := imxClient.RegisterOffchain(ctx, l1signer, l2signer, email)
	if err != nil {
		log.Printf("failed to register user in ImmutableX: %v\n", err)
		return ""
	}

	return imxres.TxHash
}

func Mint(userPrivateKey string, userAddress string, contractAddress string, tokenID string, tokenMetadata string) string {
	return "0"
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return ""
	}

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
		log.Printf("error in IMX Mint: %v\n", err)
		return ""
	}

	res := imxres.GetResults()
	return res[0].TokenId
}

func Sell(userPrivateKey string, userAddress string, starkPrivateKeyStr string, contractAddress string, tokenID string, amount imx.Wei) (int32, error) {
	return 0, nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return 0, err
	}

	starkPrivateKey := new(big.Int)
	starkPrivateKey.SetString(starkPrivateKeyStr, 16)
	l2signer, err := stark.NewSigner(starkPrivateKey)
	if err != nil {
		log.Printf("error in creating StarkSigner: %v\n", err)
		return 0, err
	}

	sellToken := imx.SignableERC721Token(tokenID, contractAddress)
	buyToken := imx.SignableETHToken()
	createOrderRequest := &api.GetSignableOrderRequest{
		AmountBuy:  strconv.FormatUint(amount, 10),
		AmountSell: "1",
		Fees:       nil,
		TokenBuy:   buyToken,
		TokenSell:  sellToken,
		User:       userAddress,
	}
	createOrderRequest.SetExpirationTimestamp(0)

	createOrderResponse, err := imxClient.CreateOrder(ctx, l1signer, l2signer, createOrderRequest)
	if err != nil {
		log.Printf("error in IMX CreateOrder: %v", err)
		return 0, err
	}

	log.Printf("CreateOrder ID: %v", createOrderResponse.OrderId)
	return createOrderResponse.OrderId, nil
}

func CancelSale(userPrivateKey string, starkPrivateKeyStr string, saleID string) (int32, error) {
	return 0, nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return 0, err
	}

	starkPrivateKey := new(big.Int)
	starkPrivateKey.SetString(starkPrivateKeyStr, 16)
	l2signer, err := stark.NewSigner(starkPrivateKey)
	if err != nil {
		log.Printf("error in creating StarkSigner: %v\n", err)
		return 0, err
	}

	id, _ := strconv.ParseInt(saleID, 10, 32)
	cancelOrderRequest := api.GetSignableCancelOrderRequest{
		OrderId: int32(id),
	}

	cancelOrderResponse, err := imxClient.CancelOrder(ctx, l1signer, l2signer, cancelOrderRequest)
	if err != nil {
		log.Printf("error in IMX CancelOrder: %v", err)
		return 0, err
	}

	log.Printf("cancelled selling for ID: %v", cancelOrderResponse.OrderId)
	return cancelOrderResponse.OrderId, nil
}

func Buy(userPrivateKey string, starkPrivateKeyStr string, saleID string) (int32, error) {
	return 0, nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return 0, err
	}

	starkPrivateKey := new(big.Int)
	starkPrivateKey.SetString(starkPrivateKeyStr, 16)
	l2signer, err := stark.NewSigner(starkPrivateKey)
	if err != nil {
		log.Printf("error in creating StarkSigner: %v\n", err)
		return 0, err
	}

	id, _ := strconv.ParseInt(saleID, 10, 64)
	tradeRequest := api.GetSignableTradeRequest{
		Fees:    nil,
		OrderId: int32(id),
	}
	tradeRequest.SetExpirationTimestamp(0)
	tradeResponse, err := imxClient.CreateTrade(ctx, l1signer, l2signer, tradeRequest)
	if err != nil {
		log.Printf("error in IMX CreateTrade: %v", err)
		return 0, err
	}

	log.Printf("trade ID: %v", tradeResponse.TradeId)
	return tradeResponse.TradeId, nil
}

func Transfer(userPrivateKey string, starkPrivateKeyStr string, receiver string) (int32, error) {
	return 0, nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return 0, err
	}

	starkPrivateKey := new(big.Int)
	starkPrivateKey.SetString(starkPrivateKeyStr, 16)
	l2signer, err := stark.NewSigner(starkPrivateKey)
	if err != nil {
		log.Printf("error in creating StarkSigner: %v\n", err)
		return 0, err
	}

	request := api.GetSignableTransferRequestV1{
		Amount:   "1",
		Sender:   l1signer.GetAddress(),
		Token:    imx.SignableETHToken(),
		Receiver: receiver,
	}

	response, err := imxClient.Transfer(ctx, l1signer, l2signer, request)
	if err != nil {
		log.Printf("error calling transfer workflow: %v", err)
		return 0, err
	}

	log.Printf("trade ID: %v", response.TransferId)
	return response.TransferId, nil
}

func Deposit(userPrivateKey string, amount string) error {
	return nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return err
	}

	ethAmountInWei, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		log.Printf("error in converting ethAmountInWei from string to int: %v\n", err)
		return err
	}

	transaction, err := imx.NewETHDeposit(ethAmountInWei).Deposit(ctx, imxClient, l1signer, nil)
	if err != nil {
		log.Printf("Eth deposit failure: %v", err)
		return err
	}

	log.Println("Eth Deposit transaction hash:", transaction.Hash())
	return nil
}

func WithdrawPrepare(userPrivateKey string, starkPrivateKeyStr string, amount string) (int32, error) {
	return 0, nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return 0, err
	}

	starkPrivateKey := new(big.Int)
	starkPrivateKey.SetString(starkPrivateKeyStr, 16)
	l2signer, err := stark.NewSigner(starkPrivateKey)
	if err != nil {
		log.Printf("error in creating StarkSigner: %v\n", err)
		return 0, err
	}

	withdrawRequest := api.GetSignableWithdrawalRequest{
		Amount: amount,
		Token:  imx.SignableETHToken(),
	}

	response, err := imxClient.PrepareWithdrawal(ctx, l1signer, l2signer, withdrawRequest)
	if err != nil {
		log.Printf("error calling PrepareWithdrawal in IMX: %v", err)
		return 0, err
	}

	return response.WithdrawalId, nil
}

func WithdrawGetState(withdrawID int32) (string, error) {
	return "", nil
	ctx, _, imxClient := Connect()

	withdrawState, err := imxClient.GetWithdrawal(ctx, strconv.FormatInt(int64(withdrawID), 10))
	if err != nil {
		log.Printf("error calling GetWithdrawal in IMX: %v", err)
		return "", err
	}

	return withdrawState.RollupStatus, nil
}

// NOTE: this should be called only after WithdrawGetState function returns "confirmed" (as per IMX documentation)
func WithdrawFinalize(userPrivateKey string, starkPrivateKeyStr string) error {
	return nil
	ctx, cfg, imxClient := Connect()
	l1signer, err := ethereum.NewSigner(userPrivateKey, cfg.ChainID)
	if err != nil {
		log.Printf("failed to create L1Signer: %v\n", err)
		return err
	}

	starkPrivateKey := new(big.Int)
	starkPrivateKey.SetString(starkPrivateKeyStr, 16)
	l2signer, err := stark.NewSigner(starkPrivateKey)
	if err != nil {
		log.Printf("error in creating StarkSigner: %v\n", err)
		return err
	}

	ethWithdrawal := imx.NewEthWithdrawal()
	transaction, err := ethWithdrawal.CompleteWithdrawal(ctx, imxClient, l1signer, l2signer.GetAddress(), nil)
	if err != nil {
		log.Panicf("error calling CompleteEthWithdrawal in IMX: %v", err)
		return err
	}

	log.Println("Eth withdraw transaction hash:", transaction.Hash())
	return nil
}
