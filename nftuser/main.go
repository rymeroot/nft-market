package nftuser

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/immutable/imx-core-sdk-golang/imx/signers/stark"
	"github.com/labstack/echo/v4"
	"log"
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
	//if req.Email == "" || req.PublicKey == "" || req.PrivateKey == "" {
	if req.Email == "" {
		return errors.New("please provide email")
	}
	return nil
}

func failWith(msg string, err error) (string, error) {
	log.Printf(msg+": %v", err)
	return "", errors.New(msg)
}

func userCreate(req *registerRequest) (string, error) {
	h := sha256.New()
	//h.Write([]byte(req.Email + req.PublicKey + req.PrivateKey))
	h.Write([]byte(req.Email))
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

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create user wallet (private key)", err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	//err = os.WriteFile(userid+"/private_key", []byte(req.PrivateKey), 0644)
	err = os.WriteFile(userid+"/private_key", []byte(hexutil.Encode(privateKeyBytes)[2:]), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create user infrastructure (private key)", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		_ = os.RemoveAll(userid)
		return failWith("error casting public key to ECDSA", nil)
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	//err = os.WriteFile(userid+"/public_key", []byte(req.PublicKey), 0644)
	err = os.WriteFile(userid+"/public_key", []byte(hexutil.Encode(publicKeyBytes)[4:]), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create user infrastructure (public key)", err)
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	err = os.WriteFile(userid+"/address", []byte(address), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create user infrastructure (address)", err)
	}

	privateStarkKey, err := stark.GenerateKey()
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to generate Stark Private Key", err)
	}
	err = os.WriteFile(userid+"/stark_private_key", []byte(fmt.Sprintf("%x", privateStarkKey)), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create user infrastructure (stark private key)", err)
	}
	l2signer, err := stark.NewSigner(privateStarkKey)
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create StarkSigner", err)
	}
	err = os.WriteFile(userid+"/stark_address", []byte(l2signer.GetAddress()), 0644)
	if err != nil {
		_ = os.RemoveAll(userid)
		return failWith("failed to create user infrastructure (stark public key)", err)
	}

	/*
		// don't call imx for now
			ctx := context.TODO()
			imxConfig := api.NewConfiguration()
			cfg := imx.Config{
				APIConfig:     imxConfig,
				AlchemyAPIKey: "WmzAboIrYGOEDnuUJnxQ8ucuu3jfoR_q",
				Environment:   imx.Sandbox,
			}

			imxClient, err := imx.NewClient(&cfg)
			if err != nil {
				_ = os.RemoveAll(userid)
				log.Printf("failed to create imx client: %v\n", err)
				return "", errors.New("failed to create imx client")
			}
			//defer imxClient.EthClient.Close()

			l1signer, err := ethereum.NewSigner(req.PrivateKey, cfg.ChainID)
			if err != nil {
				log.Printf("failed to create L1Signer: %v\n", err)
				//_ = os.RemoveAll(userid)
				return "", errors.New("failed to create L1Signer")
			}

			_, err = imxClient.RegisterOffchain(ctx, l1signer, l2signer, req.Email)
			if err != nil {
				log.Printf("failed to register user in ImmutableX: %v\n", err)
				//_ = os.RemoveAll(userid)
				return "", errors.New("failed to register user in ImmutableX")
			}
	*/
	/*
		// private key from testing1 account
		adminl1signer, err := ethereum.NewSigner("", cfg.ChainID)
		//adminl1signer, err := ethereum.NewSigner(req.PrivateKey, cfg.ChainID)
		if err != nil {
			log.Printf("failed to create adminL1Signer: %v\n", err)
			_ = os.RemoveAll(userid)
			return "", errors.New("failed to create adminL1Signer")
		}
	*/
	/*
		// for debug
		projectReponse, err := imxClient.GetProject(ctx, adminl1signer, "")
		if err != nil {
			log.Printf("error in GetProject: %v", err)
		}
		val, err := json.MarshalIndent(projectReponse, "", "    ")
		if err != nil {
			log.Printf("error in json marshaling: %v\n", err)
		}
		log.Println("Project details: ", string(val))
	*/
	/*
		// seems creating more than one project doesn't work, imx returns error 500
		response, err := imxClient.CreateProject(ctx, adminl1signer, userid, "Something Inc.", "")
		if err != nil {
			log.Printf("failed to create project: %v\n", err)
			//_ = os.RemoveAll(userid)
			return "", errors.New("failed to create ImmutableX project")
		}
		err = os.WriteFile(userid+"/project_id", []byte(strconv.FormatInt(int64(response.Id), 10)), 0644)
		if err != nil {
			//_ = os.RemoveAll(userid)
			return "", errors.New("failed to create user infrastructure (project_id)")
		}
	*/

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
