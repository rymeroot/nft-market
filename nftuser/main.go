package nftuser

import (
	"errors"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"nft-market/storage"
)

type userRequest struct {
	UserID   string               `json:"userid,omitempty"`
	Register *userRegisterRequest `json:"register,omitempty"`
	Deposit  *userDepositRequest  `json:"deposit,omitempty"`
	Withdraw *userWithdrawRequest `json:"withdraw,omitempty"`
}

type userResponse struct {
	Register *userRegisterResponse `json:"register,omitempty"`
	Deposit  *userDepositResponse  `json:"deposit,omitempty"`
	Withdraw *userWithdrawResponse `json:"withdraw,omitempty"`
}

func VerifyUserID(userid string) error {
	// TODO: check userid format
	if userid == "" {
		return errors.New("invalid user ID")
	}
	return nil
}

func User(c echo.Context) error {
	var req userRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := VerifyUserID(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if !storage.UserExists(req.UserID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user " + req.UserID + " doesn't exist"})
	}

	var resRegister *userRegisterResponse = nil
	var resDeposit *userDepositResponse = nil
	var resWithdraw *userWithdrawResponse = nil

	if req.Register != nil {
		resRegister = new(userRegisterResponse)
		err := userRegister(req.Register, resRegister)
		if err != nil {
			log.Printf("error creating in minting: %v", err)
		}
	}

	if req.Deposit != nil {
		resDeposit = new(userDepositResponse)
		err := userDeposit(req.UserID, req.Deposit, resDeposit)
		if err != nil {
			log.Printf("error performing deposit: %v", err)
		}
	}

	if req.Withdraw != nil {
		resWithdraw = new(userWithdrawResponse)
		err := userWithdraw(req.UserID, req.Withdraw, resWithdraw)
		if err != nil {
			log.Printf("error performing withdraw: %v", err)
		}
	}

	res := userResponse{
		Register: resRegister,
		Deposit:  resDeposit,
		Withdraw: resWithdraw,
	}

	pretty := c.QueryParam("pretty") == "true"
	if pretty {
		return c.JSONPretty(http.StatusOK, res, "    ")
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
