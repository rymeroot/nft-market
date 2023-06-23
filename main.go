package main

import (
	"github.com/alitto/pond"
	"github.com/labstack/echo/v4"
	"log"
	"nft-market/nftcollection"
	"nft-market/nfttoken"
	"nft-market/nftuser"
	"nft-market/storage"
	"os"
)

func main() {
	nftuser.WorkerPool = pond.New(100, 1000)
	if !storage.StorageExists() {
		err := storage.StorageCreate()
		if err != nil {
			log.Panic("failed to create storage")
			return
		}
	}
	entries, err := os.ReadDir(storage.Prefix + storage.UserDir)
	if err != nil {
		log.Panic("failed to read storage")
		return
	}

	// retrieve list of withdrawals in progress and run finalize
	for _, user := range entries {
		userid := user.Name()
		if !user.IsDir() {
			continue
		}
		if storage.UserWithdrawInProgress(userid) {
			nftuser.UserWithdrawFinalize(userid)
		}
	}

	e := echo.New()
	e.POST("/user", nftuser.User)
	e.POST("/collection", nftcollection.Collection)
	e.POST("/token", nfttoken.Token)
	e.Logger.Fatal(e.Start(":8080"))
}
