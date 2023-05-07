package p2p

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/levilutz/basiccoin/src/utils"
)

type VersionResp struct {
	Version     string `json:"version"`
	CurrentTime int64  `json:"currentTime"`
}

func Mount(r gin.IRouter) {
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, VersionResp{
			utils.Constants.AppVersion,
			time.Now().UnixMicro(),
		})
	})
}
