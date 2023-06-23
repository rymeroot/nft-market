package storage

import (
	"errors"
	"github.com/holiman/uint256"
	"os"
	"strconv"
)

const Prefix = "data/"
const UserDir = "users/"
const TokenDir = "tokens/"

func StorageExists() bool {
	if _, err := os.Stat(Prefix + UserDir); err != nil {
		return false
	}
	return true
}

func StorageCreate() error {
	err := os.MkdirAll(Prefix+UserDir, os.ModePerm)
	if err != nil {
		return errors.New("failed to create user storage")
	}
	return nil
}

func UserExists(userid string) bool {
	if _, err := os.Stat(Prefix + UserDir + userid); err != nil {
		return false
	}
	return true
}

func CollectionExists(userid string, collectionid string) bool {
	if _, err := os.Stat(Prefix + UserDir + userid + "/collections/" + collectionid); err != nil {
		return false
	}
	return true
}

func GetUserPrivateKey(userid string) ([]byte, error) {
	return os.ReadFile(Prefix + UserDir + userid + "/private_key")
}

func GetUserPublicKey(userid string) ([]byte, error) {
	return os.ReadFile(Prefix + UserDir + userid + "/public_key")
}

func GetUserAddress(userid string) ([]byte, error) {
	return os.ReadFile(Prefix + UserDir + userid + "/address")
}

func GetUserStarkPrivateKey(userid string) ([]byte, error) {
	return os.ReadFile(Prefix + UserDir + userid + "/stark_private_key")
}

func GetUserCollectionContractAddress(userid string, collectionid string) ([]byte, error) {
	return os.ReadFile(Prefix + UserDir + userid + "/collections/" + collectionid + "/contract_address")
}

func UserWithdrawInProgress(userid string) bool {
	if _, err := os.Stat(Prefix + UserDir + userid + "/withdraw"); err != nil {
		return false
	}
	return true
}

func UserWithdrawFinalize(userid string) bool {
	if err := os.Remove(Prefix + UserDir + userid + "/withdraw"); err != nil {
		return false
	}
	return true
}

func GetUserWithdrawID(userid string) (int32, error) {
	bytes, err := os.ReadFile(Prefix + UserDir + userid + "/withdraw")
	if err != nil {
		return -1, err
	}

	id, err := strconv.ParseInt(string(bytes), 10, 32)
	if err != nil {
		return -1, err
	}

	return int32(id), nil
}

func SetUserWithdrawID(userid string, withdrawID int32) error {
	return os.WriteFile(Prefix+UserDir+userid+"/withdraw", []byte(strconv.FormatInt(int64(withdrawID), 10)), 0644)
}

func GetTokenIndex(tokenID *uint256.Int) error {
	if _, err := os.Stat(Prefix + "tokens/index"); err == nil {
		bytes, err := os.ReadFile(Prefix + "tokens/index")
		if err != nil {
			return errors.New("failed to read token index")
		}
		tokenID.SetBytes(bytes)
	}

	return nil
}

func SetTokenIndex(tokenID *uint256.Int) error {
	if _, err := os.Stat(Prefix + TokenDir); err != nil {
		err = os.MkdirAll(Prefix+TokenDir, os.ModePerm)
		if err != nil {
			return errors.New("failed to create token storage")
		}
	}
	return os.WriteFile(Prefix+"tokens/index", tokenID.Bytes(), 0644)
}

func TokenSelling(userid string, tokenid string) bool {
	if _, err := os.Stat(Prefix + TokenDir + tokenid + "/selling"); err != nil {
		return false
	}
	return true
}

func GetTokenSellingID(tokenid string) ([]byte, error) {
	return os.ReadFile(Prefix + TokenDir + tokenid + "/selling")
}

func GetTokenSellingList(userid string, list []string) error {
	entries, err := os.ReadDir(Prefix + TokenDir)
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
	return os.ReadFile(Prefix + TokenDir + tokenid + "/user_id")
}

func SetTokenOwner(tokenid string, userid string) error {
	return os.WriteFile(Prefix+TokenDir+tokenid+"/user_id", []byte(userid), 0644)
}

func GetTokenCollection(tokenid string) ([]byte, error) {
	return os.ReadFile(Prefix + TokenDir + tokenid + "/collection_id")
}

func SetTokenCollection(tokenid string, collectionid string) error {
	return os.WriteFile(Prefix+TokenDir+tokenid+"/collection_id", []byte(collectionid), 0644)
}

func TokenMinted(userid string, tokenid string) bool {
	if _, err := os.Stat(Prefix + TokenDir + tokenid + "/minted"); err != nil {
		return false
	}
	return true
}

func GetTokenMintedID(tokenid string) ([]byte, error) {
	return os.ReadFile(Prefix + TokenDir + tokenid + "/minted")
}

func SetTokenMintedID(userid string, tokenid string, imxtokenid string) error {
	return os.WriteFile(Prefix+TokenDir+tokenid+"/minted", []byte(imxtokenid), 0644)
}

func CreateToken(userid string, collectionid string, tokenid string) error {
	err := os.MkdirAll(Prefix+TokenDir+tokenid, os.ModePerm)
	if err != nil {
		return errors.New("failed to reserve token")
	}

	// TODO: convert this to symlink
	err = os.MkdirAll(Prefix+UserDir+userid+"/collections/"+collectionid+"/"+tokenid, os.ModePerm)
	if err != nil {
		RemoveToken(userid, collectionid, tokenid)
		return errors.New("failed to reserve token in collection")
	}

	return nil
}

func RemoveToken(userid string, collectionid string, tokenid string) {
	_ = os.RemoveAll(Prefix + TokenDir + tokenid)
	_ = os.RemoveAll(Prefix + UserDir + userid + "/collections/" + collectionid + "/" + tokenid)
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
	_, err = os.Stat(Prefix + UserDir + string(from) + tokenPath + tokenid)
	if err != nil {
		_ = SetTokenOwner(tokenid, string(from))
		return err
	}
	err = os.MkdirAll(Prefix+UserDir+to+tokenPath, os.ModePerm)
	if err != nil {
		_ = SetTokenOwner(tokenid, string(from))
		return err
	}

	err = os.Rename(Prefix+UserDir+string(from)+tokenPath+tokenid, Prefix+UserDir+to+tokenPath+tokenid)
	if err != nil {
		_ = SetTokenOwner(tokenid, string(from))
		return err
	}

	return nil
}
