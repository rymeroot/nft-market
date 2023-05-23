package nfttoken

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"nft-market/nftuser"
	"nft-market/storage"
)

type tokenInfo struct {
	ID       string `json:"id"`
	Metadata string `json:"metadata"`
}

type tokenRequest struct {
	UserID   string                `json:"userid"`
	Mint     *tokenMintRequest     `json:"mint,omitempty"`
	Sell     *tokenSellRequest     `json:"sell,omitempty"`
	Buy      *tokenBuyRequest      `json:"buy,omitempty"`
	Transfer *tokenTransferRequest `json:"transfer,omitempty"`
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

	if !storage.UserExists(req.UserID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user " + req.UserID + " doesn't exist"})
	}

	var resMint *tokenMintResponse = nil
	var resSell *tokenSellResponse = nil
	var resBuy *tokenBuyResponse = nil
	var resTransfer *tokenTransferResponse = nil

	if req.Mint != nil {
		resMint = new(tokenMintResponse)
		err := tokenMint(req.UserID, req.Mint, resMint)
		if err != nil {
			log.Printf("error creating in minting: %v", err)
		}
	}

	if req.Sell != nil {
		// create order in IMX
		resSell = new(tokenSellResponse)
		err := tokenSell(req.UserID, req.Sell, resSell)
		if err != nil {
			log.Printf("error creating sell listing: %v", err)
		}
	}

	if req.Buy != nil {
		// create trade in IMX
		resBuy = new(tokenBuyResponse)
		err := tokenBuy(req.UserID, req.Buy, resBuy)
		if err != nil {
			log.Printf("error buying a token: %v", err)
		}
	}

	if req.Transfer != nil {
		resTransfer = new(tokenTransferResponse)
		err := tokenTransfer(req.UserID, req.Transfer, resTransfer)
		if err != nil {
			log.Printf("error transferring a token: %v", err)
		}
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
