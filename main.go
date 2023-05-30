package main

import (
	"github.com/labstack/echo/v4"
	"nft-market/nftcollection"
	"nft-market/nfttoken"
	"nft-market/nftuser"
)

func main() {
	e := echo.New()
	e.POST("/user", nftuser.User)
	e.POST("/collection", nftcollection.Collection)
	e.POST("/token", nfttoken.Token)
	e.Logger.Fatal(e.Start(":8080"))
}
