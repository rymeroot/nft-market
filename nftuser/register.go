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
	"log"
	"nft-market/storage"
	"os"
)

type userRegisterRequest struct {
	Email      string `json:"email"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type userRegisterResponse struct {
	UserID string `json:"userid"`
	Error  string `json:"error,omitempty"`
}

func verifyUserRegisterRequest(req *userRegisterRequest) error {
	// TODO: check formats
	//if req.Email == "" || req.PublicKey == "" || req.PrivateKey == "" {
	if req.Email == "" {
		return errors.New("please provide email")
	}
	return nil
}

func userRegister(req *userRegisterRequest, res *userRegisterResponse) error {
	if err := verifyUserRegisterRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	userid, err := userCreate(req)
	if err != nil {
		res.Error = err.Error()
		return err
	}

	res.UserID = userid
	return nil
}

func failWith(msg string, err error) (string, error) {
	log.Printf(msg+": %v", err)
	return "", errors.New(msg)
}

func userCreate(req *userRegisterRequest) (string, error) {
	h := sha256.New()
	//h.Write([]byte(req.Email + req.PublicKey + req.PrivateKey))
	h.Write([]byte(req.Email))
	userid := hex.EncodeToString(h.Sum(nil))

	err := VerifyUserID(userid)
	if err != nil {
		return "", errors.New("failed to verify user id while creating it")
	}

	if storage.UserExists(userid) {
		return "", errors.New("user " + userid + " already registered")
	}

	err = os.MkdirAll(storage.Prefix+storage.UserDir+userid+"/collections/tokens", os.ModePerm)
	if err != nil {
		return "", errors.New("failed to create user infrastructure")
	}

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create user wallet (private key)", err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyString := hexutil.Encode(privateKeyBytes)[2:]
	err = os.WriteFile(storage.Prefix+storage.UserDir+userid+"/private_key", []byte(privateKeyString), 0644)
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create user infrastructure (private key)", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("error casting public key to ECDSA", nil)
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	err = os.WriteFile(storage.Prefix+storage.UserDir+userid+"/public_key", []byte(hexutil.Encode(publicKeyBytes)[4:]), 0644)
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create user infrastructure (public key)", err)
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	err = os.WriteFile(storage.Prefix+storage.UserDir+userid+"/address", []byte(address), 0644)
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create user infrastructure (address)", err)
	}

	privateStarkKey, err := stark.GenerateKey()
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to generate Stark Private Key", err)
	}
	err = os.WriteFile(storage.Prefix+storage.UserDir+userid+"/stark_private_key", []byte(fmt.Sprintf("%x", privateStarkKey)), 0644)
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create user infrastructure (stark private key)", err)
	}
	l2signer, err := stark.NewSigner(privateStarkKey)
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create StarkSigner", err)
	}
	err = os.WriteFile(storage.Prefix+storage.UserDir+userid+"/stark_address", []byte(l2signer.GetAddress()), 0644)
	if err != nil {
		_ = os.RemoveAll(storage.Prefix + storage.UserDir + userid)
		return failWith("failed to create user infrastructure (stark public key)", err)
	}

	// don't call imx for now
	//imxauth := nftimx.Register(privateKeyString, l2signer, req.Email)
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
