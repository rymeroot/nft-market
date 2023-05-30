package storage

import (
	"errors"
	"github.com/holiman/uint256"
	"os"
)

func UserExists(userid string) bool {
	if _, err := os.Stat(userid); err != nil {
		return false
	}
	return true
}

func CollectionExists(userid string, collectionid string) bool {
	if _, err := os.Stat(userid + "/collections/" + collectionid); err != nil {
		return false
	}
	return true
}

func GetUserPrivateKey(userid string) ([]byte, error) {
	return os.ReadFile(userid + "/private_key")
}

func GetUserPublicKey(userid string) ([]byte, error) {
	return os.ReadFile(userid + "/public_key")
}

func GetUserAddress(userid string) ([]byte, error) {
	return os.ReadFile(userid + "/address")
}

func GetUserStarkPrivateKey(userid string) ([]byte, error) {
	return os.ReadFile(userid + "/stark_private_key")
}

func GetUserCollectionContractAddress(userid string, collectionid string) ([]byte, error) {
	return os.ReadFile(userid + "/collections/" + collectionid + "/contract_address")
}

func GetTokenIndex(tokenID *uint256.Int) error {
	if _, err := os.Stat("tokens/index"); err == nil {
		bytes, err := os.ReadFile("tokens/index")
		if err != nil {
			return errors.New("failed to read token index")
		}
		tokenID.SetBytes(bytes)
	}

	return nil
}

func SetTokenIndex(tokenID *uint256.Int) error {
	if _, err := os.Stat("tokens/"); err != nil {
		err = os.MkdirAll("tokens/", os.ModePerm)
		if err != nil {
			return errors.New("failed to create token storage")
		}
	}
	return os.WriteFile("tokens/index", tokenID.Bytes(), 0644)
}
func TokenSelling(userid string, tokenid string) bool {
	if _, err := os.Stat("tokens/" + tokenid + "/selling"); err != nil {
		return false
	}

	return true
}

func GetTokenSellingID(tokenid string) ([]byte, error) {
	return os.ReadFile("tokens/" + tokenid + "/selling")
}

func GetTokenSellingList(userid string, list []string) error {
	entries, err := os.ReadDir("tokens/")
	if err != nil {
		return errors.New("failed to read token storage")
	}

	for _, token := range entries {
		tokenid := token.Name()
		if !token.IsDir() {
			continue
		}
		if TokenSelling(userid, tokenid) {
			list = append(list, tokenid)
		}
	}

	return nil
}

func GetTokenOwner(tokenid string) ([]byte, error) {
	return os.ReadFile("tokens/" + tokenid + "/user_id")
}

func SetTokenOwner(tokenid string, userid string) error {
	return os.WriteFile("tokens/"+tokenid+"/user_id", []byte(userid), 0644)
}

func GetTokenCollection(tokenid string) ([]byte, error) {
	return os.ReadFile("tokens/" + tokenid + "/collection_id")
}

func SetTokenCollection(tokenid string, collectionid string) error {
	return os.WriteFile("tokens/"+tokenid+"/collection_id", []byte(collectionid), 0644)
}

func GetTokenMintedID(tokenid string) ([]byte, error) {
	return os.ReadFile("tokens/" + tokenid + "/minted")
}

func CreateToken(userid string, collectionid string, tokenid string) error {
	err := os.MkdirAll("tokens/"+tokenid, os.ModePerm)
	if err != nil {
		return errors.New("failed to reserve token")
	}

	// TODO: convert this to symlink
	err = os.MkdirAll(userid+"/collections/"+collectionid+"/"+tokenid, os.ModePerm)
	if err != nil {
		RemoveToken(userid, collectionid, tokenid)
		return errors.New("failed to reserve token in collection")
	}

	return nil
}

func RemoveToken(userid string, collectionid string, tokenid string) {
	_ = os.RemoveAll("tokens/" + tokenid)
	_ = os.RemoveAll(userid + "/collections/" + collectionid + "/" + tokenid)
}

func MoveToken(tokenid string, to string) error {
	from, err := GetTokenOwner(tokenid)
	if err != nil {
		return err
	}
	collection, err := GetTokenCollection(tokenid)
	if err != nil {
		return err
	}

	err = SetTokenOwner(tokenid, to)
	if err != nil {
		return err
	}

	tokenPath := "/collections/" + string(collection)
	_, err = os.Stat(string(from) + tokenPath + tokenid)
	if err != nil {
		_ = SetTokenOwner(tokenid, string(from))
		return err
	}
	err = os.MkdirAll(to+tokenPath, os.ModePerm)
	if err != nil {
		_ = SetTokenOwner(tokenid, string(from))
		return err
	}

	err = os.Rename(string(from)+tokenPath+tokenid, to+tokenPath+tokenid)
	if err != nil {
		_ = SetTokenOwner(tokenid, string(from))
		return err
	}

	return nil
}
