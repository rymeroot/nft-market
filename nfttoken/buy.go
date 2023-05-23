package nfttoken

import (
	"errors"
	"nft-market/nftimx"
	"nft-market/storage"
)

type tokenBuyRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
}

type tokenBuyResponse struct {
	BuyID string `json:"buy_id,omitempty"`
	Error string `json:"error,omitempty"`
}

func verifyTokenBuyRequest(req *tokenBuyRequest) error {
	// TODO: verify formatting
	if req.TokenID == "" {
		return errors.New("token ID missing")
	}
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

	privateKey, err := storage.GetUserPrivateKey(userid)
	if err != nil {
		res.Error = "failed to get user private key"
		return err
	}
	starkKey, err := storage.GetUserStarkPrivateKey(userid)
	if err != nil {
		res.Error = "failed to get user private key"
		return err
	}
	sellingID, err := storage.GetTokenSellingID(req.TokenID)
	if err != nil {
		res.Error = "failed to get token selling ID"
		return err
	}

	err = storage.MoveToken(req.TokenID, userid)
	if err != nil {
		res.Error = "failed to transfer token to new owner"
		return err
	}

	buyID, err := nftimx.Buy(string(privateKey), string(starkKey), string(sellingID))
	if err != nil {
		// move token back to old owner
		_ = storage.MoveToken(req.TokenID, userid)
		res.Error = "failed to create buy trade on IMX"
		return err
	}

	tokenMarkSelling(userid, req.TokenID, "-1")
	res.BuyID = string(buyID)
	return nil
}
