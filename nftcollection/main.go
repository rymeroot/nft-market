package nftcollection

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"nft-market/nftuser"
	"os"
)

type collectionCreateRequest struct {
	ContractAddress string `json:"contract_address"`
	Name            string `json:"name"`
	Description     string `json:"description"`
}

type collectionUpdateRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type collectionListRequest struct {
	Matching string `json:"matching"`
}

type collectionInfoRequest struct {
	ID string `json:"id"`
}

type collectionRequest struct {
	UserID string                   `json:"userid"`
	Create *collectionCreateRequest `json:"create,omitempty"`
	Update *collectionUpdateRequest `json:"update,omitempty"`
	List   *collectionListRequest   `json:"list,omitempty"`
	Info   *collectionInfoRequest   `json:"info,omitempty"`
}

type collectionCreateResponse struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

type collectionUpdateResponse struct {
	Error string `json:"error,omitempty"`
}

type collectionTokenInfo struct {
	ID       string `json:"id"`
	Metadata string `json:"metadata"`
}

type collectionInfoResponse struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"name,omitempty"`
	Description string                `json:"description,omitempty"`
	Tokens      []collectionTokenInfo `json:"tokens,omitempty"`
	Error       string                `json:"error,omitempty"`
}

type collectionResponse struct {
	Create *collectionCreateResponse `json:"create,omitempty"`
	Update *collectionUpdateResponse `json:"update,omitempty"`
	List   []collectionInfoResponse  `json:"list,omitempty"`
	Info   *collectionInfoResponse   `json:"info,omitempty"`
}

func CollectionExists(userid string, collectionid string) bool {
	if _, err := os.Stat(userid + "/collections/" + collectionid); err != nil {
		return false
	}
	return true
}

func verifyCollectionCreateRequest(req *collectionCreateRequest) error {
	// TODO: verify formatting
	if req.ContractAddress == "" || req.Name == "" || req.Description == "" {
		return errors.New("collection contract address, name or description missing")
	}
	return nil
}

func collectionCreate(userid string, req *collectionCreateRequest, res *collectionCreateResponse) error {
	if err := verifyCollectionCreateRequest(req); err != nil {
		res.Error = err.Error()
		return err
	}

	h := sha256.New()
	h.Write([]byte(userid + req.ContractAddress + req.Name + req.Description))
	collectionID := hex.EncodeToString(h.Sum(nil))
	collectionPath := userid + "/collections/" + collectionID

	if CollectionExists(userid, collectionID) {
		res.Error = "collection " + collectionID + " already exists"
		return errors.New(res.Error)
	}

	privateKey, err := os.ReadFile(userid + "/private_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	publicKey, err := os.ReadFile(userid + "/public_key")
	if err != nil {
		res.Error = "failed to read user private key"
		return err
	}
	log.Printf("public key: '%v', private key: '%v'\n", string(publicKey), string(privateKey))

	/*
		// TODO: create collection on immutablex here
		// use only one admin contract and project?
		// but then need to keep track of minted token id, which also requires user to handle token id in IPFS metadata
		// also then we have to first reserve a token, give user the ID so that user could create metadata and then mint
		ctx, cfg, imxClient := nftimx.Connect()
		l1signer, err := ethereum.NewSigner(string(privateKey), cfg.ChainID)
		imxCreateCollectionRequest := api.NewCreateCollectionRequest(
			req.ContractAddress, req.Name,
			string(publicKey),4169)
		imxCreateCollectionResponse, err := imxClient.CreateCollection(ctx, l1signer, imxCreateCollectionRequest)
		imxCollection, _ := json.MarshalIndent(imxCreateCollectionResponse, "", "    ")
		log.Printf("Created new collection, response: ", string(imxCollection))
	*/

	err = os.MkdirAll(collectionPath, os.ModePerm)
	if err != nil {
		res.Error = "failed to create collection"
		return err
	}

	err = os.WriteFile(collectionPath+"/contract_address", []byte(req.ContractAddress), 0644)
	if err != nil {
		_ = os.RemoveAll(collectionPath)
		res.Error = "failed to create collection contract"
		return err
	}

	err = os.WriteFile(collectionPath+"/name", []byte(req.Name), 0644)
	if err != nil {
		_ = os.RemoveAll(collectionPath)
		res.Error = "failed to create collection name"
		return err
	}

	err = os.WriteFile(collectionPath+"/description", []byte(req.Description), 0644)
	if err != nil {
		_ = os.RemoveAll(collectionPath)
		res.Error = "failed to create collection description"
		return err
	}

	res.ID = collectionID
	res.Error = ""
	return nil
}

func Collection(c echo.Context) error {
	var req collectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := nftuser.VerifyUserID(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if !nftuser.UserExists(req.UserID) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "user " + req.UserID + " doesn't exist"})
	}

	var resCreate *collectionCreateResponse = nil
	var resUpdate *collectionUpdateResponse = nil
	var resList []collectionInfoResponse = nil
	var resInfo *collectionInfoResponse = nil

	if req.Create != nil {
		resCreate = new(collectionCreateResponse)
		_ = collectionCreate(req.UserID, req.Create, resCreate)
	}

	if req.Update != nil {
		resUpdate = new(collectionUpdateResponse)
		resUpdate.Error = "all good"
	}

	if req.List != nil {
		collectionInfo := collectionInfoResponse{
			ID:          "id1",
			Name:        "name 1",
			Description: "description 1",
		}
		resList = append(resList, collectionInfo)

		collectionInfo = collectionInfoResponse{
			ID:          "id2",
			Name:        "name 2",
			Description: "description 2",
		}
		resList = append(resList, collectionInfo)
	}

	if req.Info != nil {
		resInfo = new(collectionInfoResponse)

		if req.Info.ID != "" {
			resInfo = &collectionInfoResponse{
				ID:          req.Info.ID,
				Name:        "name for " + req.Info.ID,
				Description: "description for " + req.Info.ID,
			}
		} else {
			resInfo = &collectionInfoResponse{
				Error: "wrong collection ID",
			}
		}
	}

	res := collectionResponse{
		Create: resCreate,
		Update: resUpdate,
		List:   resList,
		Info:   resInfo,
	}

	pretty := c.QueryParam("pretty") == "true"
	if pretty {
		return c.JSONPretty(http.StatusOK, res, "    ")
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
