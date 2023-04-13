package nfttoken

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"nft-market/nftuser"
)

type TokenInfo struct {
	ID       string `json:"id"`
	Metadata string `json:"metadata"`
}

type tokenMintRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
	Metadata     string `json:"metadata"`
}

type tokenSellRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
	Price        string `json:"price"`
}

type tokenBuyRequest struct {
	SellID string `json:"sell_id"`
}

type tokenTransferRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
	To           string `json:"to"`
}

type tokenRequest struct {
	UserID   string                `json:"userid"`
	Mint     *tokenMintRequest     `json:"mint,omitempty"`
	Sell     *tokenSellRequest     `json:"sell,omitempty"`
	Buy      *tokenBuyRequest      `json:"buy,omitempty"`
	Transfer *tokenTransferRequest `json:"transfer,omitempty"`
}

type tokenMintResponse struct {
	MintID string `json:"mint_id"`
	Error  string `json:"error"`
}

type tokenSellResponse struct {
	SellID string `json:"sell_id"`
	Error  string `json:"error"`
}

type tokenBuyResponse struct {
	BuyID string `json:"buy_id"`
	Error string `json:"error"`
}

type tokenTransferResponse struct {
	TransferID string `json:"transfer_id"`
	Error      string `json:"error"`
}

type tokenResponse struct {
	Mint     *tokenMintResponse     `json:"mint,omitempty"`
	Sell     *tokenSellResponse     `json:"sell,omitempty"`
	Buy      *tokenBuyResponse      `json:"buy,omitempty"`
	Transfer *tokenTransferResponse `json:"transfer,omitempty"`
}

func Token(c echo.Context) error {
	var req tokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := nftuser.VerifyUserID(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var resMint *tokenMintResponse = nil
	var resSell *tokenSellResponse = nil
	var resBuy *tokenBuyResponse = nil
	var resTransfer *tokenTransferResponse = nil

	if req.Mint != nil {
		resMint = new(tokenMintResponse)
		resMint.MintID = "new mint transaction id here"
		resMint.Error = ""
	}

	if req.Sell != nil {
		resSell = new(tokenSellResponse)
		resSell.SellID = "new sell transaction id here"
		resSell.Error = ""
	}

	if req.Buy != nil {
		resBuy = new(tokenBuyResponse)
		resBuy.BuyID = "new buy transaction id here"
		resBuy.Error = ""
	}

	if req.Transfer != nil {
		resTransfer = new(tokenTransferResponse)
		resTransfer.TransferID = "new transfer transaction id here"
		resMint.Error = ""
	}

	res := tokenResponse{
		Mint:     resMint,
		Sell:     resSell,
		Buy:      resBuy,
		Transfer: resTransfer,
	}

	pretty := c.QueryParam("pretty") == "true"
	if pretty {
		return c.JSONPretty(http.StatusOK, res, "    ")
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
