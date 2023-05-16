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
	"strconv"
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
	SellingID    string `json:"selling_id,omitempty"`
}

type tokenBuyRequest struct {
	TokenID string `json:"token_id"`
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

func verifyTokenSellRequest(req *tokenSellRequest) error {
	// TODO: verify formatting
	if req.SellingID != "" {
		return nil
	}

	if req.CollectionID == "" {
		return errors.New("collection ID missing")
	}
	if req.TokenID == "" {
		return errors.New("token ID missing")
	}
	if req.Price == "" {
		return errors.New("price is missing")
	}

	return nil
}

func verifyTokenBuyRequest(req *tokenBuyRequest) error {
	// TODO: verify formatting
	if req.TokenID == "" {
		return errors.New("token ID missing")
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

func tokenMarkMinted(userid string, tokenid string, imxtokenid string) bool {
	err := os.WriteFile("tokens/"+tokenid+"/minted", []byte(imxtokenid), 0644)
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

func tokenMint(userid string, req *tokenMintRequest, res *tokenMintResponse) error {
	if err := verifyTokenMintRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if !nftcollection.CollectionExists(userid, req.CollectionID) {
		res.Error = "collection " + req.CollectionID + " doesn't exist"
		return errors.New(res.Error)
	}

	collectionPath := userid + "/collections/" + req.CollectionID
	collectionContractAddress, err := os.ReadFile(collectionPath + "/contract_address")
	if err != nil {
		res.Error = "failed to read collection contract address"
		return err
	}
	privateKey, err := os.ReadFile(userid + "/private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	userAddress, err := os.ReadFile(userid + "/address")
	if err != nil {
		res.Error = "failed to read user address"
		return err
	}

	if req.TokenID == "" {
		log.Printf("reserving token\n")
		res.TokenID, _ = tokenReserve(userid, req.CollectionID)
		return nil
	}

	if !tokenReserved(userid, req.TokenID) {
		res.Error = "wrong token ID"
		return err
	}

	if tokenMinted(userid, req.TokenID) {
		res.Error = "token already minted"
		return err
	}

	if tokenReserved(userid, req.TokenID) {
		// TODO: verify if token is reserved by userid
		imxTokenID := nftimx.Mint(string(privateKey), string(userAddress), string(collectionContractAddress), req.TokenID, req.Metadata)
		tokenMarkMinted(userid, req.TokenID, imxTokenID)
		res.MintID = imxTokenID
		return nil
	}

	return nil
}

func tokenMarkSelling(userid string, tokenid string, sellingID string) bool {
	path := "tokens/" + tokenid + "/selling"

	if sellingID == "-1" {
		// cancelling sell order
		err := os.Remove(path)
		if err != nil {
			return false
		}
		return true
	}

	err := os.WriteFile(path, []byte(sellingID), 0644)
	if err != nil {
		return false
	}

	return true
}

func tokenSelling(userid string, tokenid string) bool {
	if _, err := os.Stat("tokens/" + tokenid + "/selling"); err != nil {
		return false
	}

	return true
}

func tokenSell(userid string, req *tokenSellRequest, res *tokenSellResponse) error {
	if err := verifyTokenSellRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if !nftcollection.CollectionExists(userid, req.CollectionID) {
		res.Error = "collection " + req.CollectionID + " doesn't exist"
		return errors.New(res.Error)
	}

	if !tokenMinted(userid, req.TokenID) {
		// TODO: verify that token is minted by userid
		res.Error = "token not minted"
		return errors.New(res.Error)
	}

	if tokenSelling(userid, req.TokenID) && req.SellingID == "" {
		// TODO: verify that token is selling by userid
		res.Error = "token already on sale"
		return errors.New(res.Error)
	}

	collectionPath := userid + "/collections/" + req.CollectionID
	collectionContractAddress, err := os.ReadFile(collectionPath + "/contract_address")
	if err != nil {
		res.Error = "failed to read collection contract address"
		return err
	}
	privateKey, err := os.ReadFile(userid + "/private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	userAddress, err := os.ReadFile(userid + "/address")
	if err != nil {
		res.Error = "failed to read user address"
		return err
	}
	starkKey, err := os.ReadFile(userid + "/stark_private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	imxTokenID, err := os.ReadFile("tokens/" + req.TokenID + "/minted")
	if err != nil {
		res.Error = "failed to read minted token ID"
		return err
	}

	if req.SellingID != "" {
		sellID, err := nftimx.CancelSale(string(privateKey), string(starkKey), req.SellingID)
		if err != nil {
			res.Error = "failed to cancel sell order on IMX"
			return err
		}

		tokenMarkSelling(userid, req.TokenID, "-1")
		res.SellID = string(sellID)
		return nil
	}

	listingPriceInWei, _ := strconv.ParseUint(req.Price, 10, 64)
	sellID, err := nftimx.Sell(string(privateKey), string(userAddress), string(starkKey), string(collectionContractAddress), string(imxTokenID), listingPriceInWei)
	if err != nil {
		res.Error = "failed to create sell order on IMX"
		return err
	}

	tokenMarkSelling(userid, req.TokenID, string(sellID))
	res.SellID = string(sellID)
	return nil
}

func tokenBuy(userid string, req *tokenBuyRequest, res *tokenBuyResponse) error {
	if err := verifyTokenBuyRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if !tokenSelling(userid, req.TokenID) {
		// TODO: verify that token is selling by userid
		res.Error = "token is not on sale"
		return errors.New(res.Error)
	}

	privateKey, err := os.ReadFile(userid + "/private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	starkKey, err := os.ReadFile(userid + "/stark_private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	sellingID, err := os.ReadFile("tokens/" + req.TokenID + "/selling")
	if err != nil {
		res.Error = "failed to read token selling ID"
		return err
	}

	buyID, err := nftimx.Buy(string(privateKey), string(starkKey), string(sellingID))
	if err != nil {
		res.Error = "failed to create buy trade on IMX"
		return err
	}

	tokenMarkSelling(userid, req.TokenID, "-1")
	res.BuyID = string(buyID)
	return nil
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
		// TODO: move token data to new user
		resBuy = new(tokenBuyResponse)
		err := tokenBuy(req.UserID, req.Buy, resBuy)
		if err != nil {
			log.Printf("error buying a token: %v", err)
		}
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
