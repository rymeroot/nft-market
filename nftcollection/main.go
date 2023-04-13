package nftcollection

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"nft-market/nfttoken"
	"nft-market/nftuser"
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
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

type collectionUpdateResponse struct {
	Error string `json:"error,omitempty"`
}

type collectionInfoResponse struct {
	ID          string               `json:"id,omitempty"`
	Name        string               `json:"name,omitempty"`
	Description string               `json:"description,omitempty"`
	Tokens      []nfttoken.TokenInfo `json:"tokens,omitempty"`
	Error       string               `json:"error,omitempty"`
}

type collectionResponse struct {
	Create *collectionCreateResponse `json:"create,omitempty"`
	Update *collectionUpdateResponse `json:"update,omitempty"`
	List   []collectionInfoResponse  `json:"list,omitempty"`
	Info   *collectionInfoResponse   `json:"info,omitempty"`
}

func Collection(c echo.Context) error {
	var req collectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := nftuser.VerifyUserID(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var resCreate *collectionCreateResponse = nil
	var resUpdate *collectionUpdateResponse = nil
	var resList []collectionInfoResponse = nil
	var resInfo *collectionInfoResponse = nil

	if req.Create != nil {
		resCreate = new(collectionCreateResponse)
		resCreate.ID = "new collection id here"
		if req.Create.Name != "" {
			resCreate.ID = req.Create.Name
		}
		resCreate.Error = ""
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
