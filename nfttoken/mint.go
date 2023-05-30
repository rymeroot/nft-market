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

	err := storage.CreateToken(userid, collectionID, res)
	if err != nil {
		return errors.New("failed to reserve token")
	}

	// TODO: here parallel request from another user theoretically could hijack the reservation

	err = storage.SetTokenOwner(res, userid)
	if err != nil {
		storage.RemoveToken(userid, collectionID, res)
		return errors.New("failed to reserve token (writing user id)")
	}

	err = storage.SetTokenCollection(res, collectionID)
	if err != nil {
		storage.RemoveToken(userid, collectionID, res)
		return errors.New("failed to reserve token (writing collection id)")
	}

	err = storage.SetTokenIndex(tokenID)
	if err != nil {
		storage.RemoveToken(userid, collectionID, res)
		return errors.New("failed to reserve token (writing index)")
	}

	return nil
}

func tokenReserve(userid string, collectionID string) (string, error) {
	tokenID := uint256.NewInt(1)

	if err := storage.GetTokenIndex(tokenID); err == nil {
		tokenID = new(uint256.Int).Add(tokenID, uint256.NewInt(1))
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
