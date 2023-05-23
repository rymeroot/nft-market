package storage

import (
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

func GetTokenSellingID(tokenid string) ([]byte, error) {
	return os.ReadFile("tokens/" + tokenid + "/selling")
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

func GetTokenMintedID(tokenid string) ([]byte, error) {
	return os.ReadFile("tokens/" + tokenid + "/minted")
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
