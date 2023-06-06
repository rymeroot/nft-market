package nftuser

import (
	"errors"
	"nft-market/nftimx"
	"nft-market/storage"
)

type userWithdrawRequest struct {
	Amount string `json:"amount"`
}

type userWithdrawResponse struct {
	TxID  string `json:"tx_id,omitempty"`
	Error string `json:"error,omitempty"`
}

func verifyUserWithdrawRequest(req *userWithdrawRequest) error {
	// TODO: verify formatting
	if req.Amount == "" {
		return errors.New("invalid withdraw amount")
	}
	return nil
}

func userWithdraw(userid string, req *userWithdrawRequest, res *userWithdrawResponse) error {
	if err := verifyUserWithdrawRequest(req); err != nil {
		res.Error = err.Error()
		return err
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

	withdrawID, err := nftimx.WithdrawPrepare(string(privateKey), string(starkKey), req.Amount)
	if err != nil {
		res.Error = "failed to prepare withdraw operation in IMX"
		return err
	}

	withdrawState, err := nftimx.WithdrawGetState(withdrawID)
	if err != nil {
		res.Error = "failed to get withdraw state from IMX"
		return err
	}

	// TODO: run withdraw finalizing in async go routine since it can take many hours to confirm (as per IMX documentation)
	if withdrawState == "confirmed" {
		err = nftimx.WithdrawFinalize(string(privateKey), string(starkKey))
		if err != nil {
			res.Error = "failed to finalize withdraw operation in IMX"
			return err
		}
	}

	return nil
}
