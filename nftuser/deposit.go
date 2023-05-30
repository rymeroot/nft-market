package nftuser

import (
	"errors"
	"nft-market/nftimx"
	"nft-market/storage"
)

type userDepositRequest struct {
	Amount string `json:"amount"`
}

type userDepositResponse struct {
	TxID  string `json:"tx_id,omitempty"`
	Error string `json:"error,omitempty"`
}

func verifyUserDepositRequest(req *userDepositRequest) error {
	// TODO: verify formatting
	if req.Amount == "" {
		return errors.New("invalid deposit amount")
	}
	return nil
}

func userDeposit(userid string, req *userDepositRequest, res *userDepositResponse) error {
	if err := verifyUserDepositRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	privateKey, err := storage.GetUserPrivateKey(userid)
	if err != nil {
		res.Error = "failed to get user private key"
		return err
	}

	err = nftimx.Deposit(string(privateKey), req.Amount)
	if err != nil {
		res.Error = "failed to perform IMX deposit operation"
		return err
	}

	return nil
}
