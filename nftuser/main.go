package nftuser

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

type registerRequest struct {
	Email      string `json:"email"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type registerResponse struct {
	UserID string `json:"userid"`
	Error  string `json:"error,omitempty"`
}

func VerifyUserID(userid string) error {
	// TODO: check userid format
	if userid == "" {
		return errors.New("invalid user ID")
	}
	return nil
}

func UserExists(userid string) bool {
	if _, err := os.Stat(userid); err != nil {
		return false
	}
	return true
}

func verifyRegisterRequest(req *registerRequest) error {
	// TODO: check formats
	if req.Email == "" || req.PublicKey == "" || req.PrivateKey == "" {
		return errors.New("need to define email, public and private keys")
	}
	return nil
}

func userCreate(req *registerRequest) (string, error) {
	h := sha256.New()
	h.Write([]byte(req.Email + req.PublicKey + req.PrivateKey))
	userid := hex.EncodeToString(h.Sum(nil))

	err := VerifyUserID(userid)
	if err != nil {
		return "", errors.New("failed to verify user id while creating it")
	}

	if UserExists(userid) {
		return "", errors.New("user " + userid + " already registered")
	}

	err = os.MkdirAll(userid+"/collections/tokens", os.ModePerm)
	if err != nil {
		return "", errors.New("failed to create user infrastructure")
	}

	err = os.WriteFile(userid+"/public_key", []byte(req.PublicKey), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return "", errors.New("failed to create user infrastructure (public key)")
	}

	err = os.WriteFile(userid+"/private_key", []byte(req.PrivateKey), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return "", errors.New("failed to create user infrastructure (private key)")
	}

	return userid, nil
}

func Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := verifyRegisterRequest(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	userid, err := userCreate(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	res := registerResponse{
		UserID: userid,
		Error:  "",
	}

	pretty := c.QueryParam("pretty") == "true"
	if pretty {
		return c.JSONPretty(http.StatusOK, res, "    ")
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
