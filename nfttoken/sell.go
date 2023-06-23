package nfttoken

import (
	"errors"
	"nft-market/nftimx"
	"nft-market/storage"
	"os"
	"strconv"
)

type tokenSellRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
	Price        string `json:"price"`
	SellingID    string `json:"selling_id,omitempty"`
}

type tokenSellResponse struct {
	SellID string `json:"sell_id,omitempty"`
	Error  string `json:"error,omitempty"`
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

func tokenMarkSelling(userid string, tokenid string, sellingID string) bool {
	path := storage.TokenDir + tokenid + "/selling"

	if sellingID == "-1" {
		// cancelling sell order
		err := os.Remove(storage.Prefix + path)
		if err != nil {
			return false
		}
		return true
	}

	err := os.WriteFile(storage.Prefix+path, []byte(sellingID), 0644)
	if err != nil {
		return false
	}

	return true
}

func tokenSell(userid string, req *tokenSellRequest, res *tokenSellResponse) error {
	if err := verifyTokenSellRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if !storage.CollectionExists(userid, req.CollectionID) {
		res.Error = "collection " + req.CollectionID + " doesn't exist"
		return errors.New(res.Error)
	}

	if !storage.TokenMinted(userid, req.TokenID) {
		// TODO: verify that token is minted by userid
		res.Error = "token not minted"
		return errors.New(res.Error)
	}

	if storage.TokenSelling(userid, req.TokenID) && req.SellingID == "" {
		// TODO: verify that token is selling by userid
		res.Error = "token already on sale"
		return errors.New(res.Error)
	}

	collectionContractAddress, err := storage.GetUserCollectionContractAddress(userid, req.CollectionID)
	if err != nil {
		res.Error = "failed to read collection contract address"
		return err
	}
	privateKey, err := storage.GetUserPrivateKey(userid)
	if err != nil {
		res.Error = "failed to get user private key"
		return err
	}
	userAddress, err := storage.GetUserAddress(userid)
	if err != nil {
		res.Error = "failed to get user address"
		return err
	}
	starkKey, err := storage.GetUserStarkPrivateKey(userid)
	if err != nil {
		res.Error = "failed to get user private key"
		return err
	}
	imxTokenID, err := storage.GetTokenMintedID(req.TokenID)
	if err != nil {
		res.Error = "failed to get minted token ID"
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
