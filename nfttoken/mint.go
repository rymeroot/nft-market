package nfttoken

import (
	"errors"
	"github.com/holiman/uint256"
	"log"
	"nft-market/nftimx"
	"nft-market/storage"
	"os"
)

type tokenMintRequest struct {
	CollectionID string `json:"collection_id"`
	TokenID      string `json:"token_id"`
	Metadata     string `json:"metadata"`
}

type tokenMintResponse struct {
	MintID  string `json:"mint_id,omitempty"`
	TokenID string `json:"token_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

func verifyTokenMintRequest(req *tokenMintRequest) error {
	// TODO: verify formatting
	if req.CollectionID == "" {
		return errors.New("collection ID missing")
	}
	return nil
}

func tokenMarkMinted(userid string, tokenid string, imxtokenid string) bool {
	err := os.WriteFile("tokens/"+tokenid+"/minted", []byte(imxtokenid), 0644)
	if err != nil {
		return false
	}

	return true
}

func tokenMinted(userid string, tokenid string) bool {
	if _, err := os.Stat("tokens/" + tokenid + "/minted"); err != nil {
		return false
	}

	return true
}

func tokenReserved(userid string, tokenid string) bool {
	// TODO: check if reserved for userid
	if _, err := os.Stat("tokens/" + tokenid); err != nil {
		return false
	}
	return true
}

func tokenSaveReservation(userid string, collectionID string, tokenID *uint256.Int) error {
	res := tokenID.String()[2:]

	err := os.MkdirAll("tokens/"+res, os.ModePerm)
	if err != nil {
		return errors.New("failed to reserve token")
	}

	// TODO: here parallel request from another user theoretically could hijack the reservation

	err = os.WriteFile("tokens/"+res+"/user_id", []byte(userid), 0644)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token (writing user id)")
	}

	err = os.WriteFile("tokens/"+res+"/collection_id", []byte(collectionID), 0644)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token (writing collection id)")
	}

	err = os.MkdirAll(userid+"/collections/"+collectionID+"/"+res, os.ModePerm)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token in collection")
	}

	err = os.WriteFile("tokens/index", tokenID.Bytes(), 0644)
	if err != nil {
		_ = os.RemoveAll("tokens/" + res)
		return errors.New("failed to reserve token (writing index)")
	}

	return nil
}

func tokenReserve(userid string, collectionID string) (string, error) {
	tokenID := uint256.NewInt(0)

	if _, err := os.Stat("tokens/"); err != nil {
		err = os.MkdirAll("tokens/", os.ModePerm)
		if err != nil {
			return "", errors.New("failed to reserve token (cannot create storage)")
		}
	}

	if _, err := os.Stat("tokens/index"); err == nil {
		bytes, err := os.ReadFile("tokens/index")
		if err != nil {
			return "", errors.New("failed to reserve token (unreadable index)")
		}
		tokenID = new(uint256.Int).Add(tokenID.SetBytes(bytes), uint256.NewInt(1))
	}

	if err := tokenSaveReservation(userid, collectionID, tokenID); err != nil {
		return "", err
	}

	res := tokenID.String()[2:]
	log.Printf("new token reservation: '%v'", res)
	return res, nil
}

func tokenMint(userid string, req *tokenMintRequest, res *tokenMintResponse) error {
	if err := verifyTokenMintRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	if tokenMinted(userid, req.TokenID) {
		res.Error = "token already minted"
		return errors.New(res.Error)
	}

	if !storage.CollectionExists(userid, req.CollectionID) {
		res.Error = "collection " + req.CollectionID + " doesn't exist"
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

	if req.TokenID == "" {
		log.Printf("reserving token\n")
		res.TokenID, _ = tokenReserve(userid, req.CollectionID)
		return nil
	}

	if !tokenReserved(userid, req.TokenID) {
		res.Error = "wrong token ID"
		return err
	}

	// TODO: verify if token is reserved by userid
	imxTokenID := nftimx.Mint(string(privateKey), string(userAddress), string(collectionContractAddress), req.TokenID, req.Metadata)
	tokenMarkMinted(userid, req.TokenID, imxTokenID)
	res.MintID = imxTokenID
	return nil
}
