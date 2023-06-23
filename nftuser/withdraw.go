package nftuser

import (
	"errors"
	"github.com/alitto/pond"
	"log"
	"nft-market/nftimx"
	"nft-market/storage"
	"time"
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

var WorkerPool *pond.WorkerPool

// NOTE: run withdraw finalizing in async go routine since it can take many hours to confirm (as per IMX documentation)
func UserWithdrawFinalize(userid string) {

	WorkerPool.Submit(
		func() {
			privateKey, err := storage.GetUserPrivateKey(userid)
			if err != nil {
				return
			}
			starkKey, err := storage.GetUserStarkPrivateKey(userid)
			if err != nil {
				return
			}

			withdrawID, err := storage.GetUserWithdrawID(userid)
			if err != nil {
				log.Printf("failed to get withdraw ID")
				return
			}
			// TODO: check IMX docs if withdraw ID could be negative
			if withdrawID < 1 {
				log.Printf("invalid withdraw ID")
				return
			}

			for {
				withdrawState, _ := nftimx.WithdrawGetState(withdrawID)
				if withdrawState == "confirmed" {
					break
				}
				// TODO: use exponential backoff? (could be problematic when starting up with a lot of withdrawals)
				time.Sleep(time.Hour)
			}

			err = nftimx.WithdrawFinalize(string(privateKey), string(starkKey))
			if err != nil {
				return
			}

			storage.UserWithdrawFinalize(userid)

			return
		},
	)
}

func userWithdraw(userid string, req *userWithdrawRequest, res *userWithdrawResponse) error {
	if err := verifyUserWithdrawRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if storage.UserWithdrawInProgress(userid) {
		res.Error = "withdraw operation is already in progress"
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

	withdrawID, err := nftimx.WithdrawPrepare(string(privateKey), string(starkKey), req.Amount)
	if err != nil {
		res.Error = "failed to prepare withdraw operation in IMX"
		return err
	}

	err = storage.SetUserWithdrawID(userid, withdrawID)
	if err != nil {
		res.Error = "failed to set user withdraw ID"
		return err
	}

	UserWithdrawFinalize(userid)

	return nil
}
