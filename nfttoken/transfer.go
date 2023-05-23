package nfttoken

import (
	"errors"
	"nft-market/nftimx"
	"nft-market/storage"
)

type tokenTransferRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
	To           string `json:"to"`
}

type tokenTransferResponse struct {
	TransferID string `json:"transfer_id,omitempty"`
	Error      string `json:"error,omitempty"`
}

func verifyTokenTransferRequest(req *tokenTransferRequest) error {
	// TODO: verify formatting, etc.
	if req.CollectionID == "" {
		return errors.New("collection ID missing")
	}
	if req.TokenID == "" {
		return errors.New("token ID missing")
	}
	if req.To == "" {
		return errors.New("receiver address is missing")
	}

	return nil
}

func tokenTransfer(userid string, req *tokenTransferRequest, res *tokenTransferResponse) error {
	if err := verifyTokenTransferRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if tokenSelling(userid, req.TokenID) {
		res.Error = "token is on sale, cancel first"
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

	err = storage.MoveToken(req.TokenID, req.To)
	if err != nil {
		res.Error = "failed to transfer token to new owner"
		return err
	}

	transferID, err := nftimx.Transfer(string(privateKey), string(starkKey), req.To)
	if err != nil {
		// move token back to old owner
		_ = storage.MoveToken(req.TokenID, userid)
		res.Error = "failed to create buy trade on IMX"
		return err
	}

	res.TransferID = string(transferID)
	return nil
}
