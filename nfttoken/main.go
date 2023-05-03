package nfttoken

import (
	"errors"
	"github.com/holiman/uint256"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"nft-market/nftcollection"
	"nft-market/nftimx"
	"nft-market/nftuser"
	"os"
)

type tokenInfo struct {
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
	MintID  string `json:"mint_id,omitempty"`
	TokenID string `json:"token_id,omitempty"`
	Error   string `json:"error,omitempty"`
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

func verifyTokenMintRequest(req *tokenMintRequest) error {
	// TODO: verify formatting
	if req.CollectionID == "" {
		return errors.New("collection ID missing")
	}
	return nil
}

func tokenReserved(userid string, tokenid string) bool {
	if _, err := os.Stat("tokens/" + tokenid); err != nil {
		return false
	}
	return true
}

func tokenSaveReservation(userid string, collectionID string, tokenID *uint256.Int) error {
	res := tokenID.String()[2:]

	err := os.MkdirAll("tokens/"+res, os.ModePerm)
	if err != nil {
		return errors.New("failed to reserve token")
	}

	err = os.WriteFile("tokens/"+res+"/user_id", []byte(userid), 0644)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token (writing user id)")
	}

	err = os.WriteFile("tokens/"+res+"/collection_id", []byte(collectionID), 0644)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token (writing collection id)")
	}

	err = os.MkdirAll(userid+"/collections/"+collectionID+"/"+res, os.ModePerm)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token in collection")
	}

	err = os.WriteFile("tokens/index", tokenID.Bytes(), 0644)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token (writing index)")
	}

	return nil
}

func tokenReserve(userid string, collectionID string) (string, error) {
	tokenID := uint256.NewInt(0)

	if _, err := os.Stat("tokens/"); err != nil {
		err = os.MkdirAll("tokens/", os.ModePerm)
		if err != nil {
			return "", errors.New("failed to reserve token (cannot create storage)")
		}
	}

	if _, err := os.Stat("tokens/index"); err == nil {
		bytes, err := os.ReadFile("tokens/index")
		if err != nil {
			return "", errors.New("failed to reserve token (unreadable index)")
		}
		tokenID = new(uint256.Int).Add(tokenID.SetBytes(bytes), uint256.NewInt(1))
	}

	if err := tokenSaveReservation(userid, collectionID, tokenID); err != nil {
		return "", err
	}

	res := tokenID.String()[2:]
	log.Printf("new token reservation: '%v'", res)
	return res, nil
}

func tokenMarkMinted(userid string, tokenid string) bool {
	err := os.WriteFile("tokens/"+tokenid+"/minted", []byte("1"), 0644)
	if err != nil {
		return false
	}

	return true
}

func tokenMinted(userid string, tokenid string) bool {
	if _, err := os.Stat("tokens/" + tokenid + "/minted"); err != nil {
		return false
	}

	return true
}

func tokenMint(userid string, req *tokenMintRequest, res *tokenMintResponse) (string, error) {
	if err := verifyTokenMintRequest(req); err != nil {
		res.Error = err.Error()
		return "", err
	}

	if !nftcollection.CollectionExists(userid, req.CollectionID) {
		res.Error = "collection " + req.CollectionID + " doesn't exist"
		return "", errors.New(res.Error)
	}

	collectionPath := userid + "/collections/" + req.CollectionID
	collectionContractAddress, err := os.ReadFile(collectionPath + "/contract_address")
	if err != nil {
		res.Error = "failed to read collection contract address"
		return "", err
	}
	privateKey, err := os.ReadFile(userid + "/private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return "", err
	}
	userAddress, err := os.ReadFile(userid + "/address")
	if err != nil {
		res.Error = "failed to read user address"
		return "", err
	}

	if req.TokenID == "" {
		log.Printf("reserving token\n")
		res.TokenID, _ = tokenReserve(userid, req.CollectionID)
		return "", nil
	}

	if !tokenReserved(userid, req.TokenID) {
		res.Error = "wrong token ID"
		return "", err
	}

	if tokenMinted(userid, req.TokenID) {
		res.Error = "token already minted"
		return "", err
	}

	if tokenReserved(userid, req.TokenID) {
		txid := nftimx.Mint(string(privateKey), string(userAddress), string(collectionContractAddress), req.TokenID, req.Metadata)
		tokenMarkMinted(userid, req.TokenID)
		return string(txid), nil
	}

	return "", nil
}

func Token(c echo.Context) error {
	var req tokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := nftuser.VerifyUserID(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if !nftuser.UserExists(req.UserID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user " + req.UserID + " doesn't exist"})
	}

	var resMint *tokenMintResponse = nil
	var resSell *tokenSellResponse = nil
	var resBuy *tokenBuyResponse = nil
	var resTransfer *tokenTransferResponse = nil

	if req.Mint != nil {
		resMint = new(tokenMintResponse)
		resMint.MintID, _ = tokenMint(req.UserID, req.Mint, resMint)
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
