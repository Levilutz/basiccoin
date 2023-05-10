package p2p

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/levilutz/basiccoin/src/utils"
)

func Mount(r gin.IRouter, pn *P2pNetwork) {
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, VersionResp{
			utils.Constants.AppVersion,
			time.Now().UnixMicro(),
			utils.Constants.RuntimeID,
		})
	})

	r.POST("/hello", func(c *gin.Context) {
		var json HelloReq
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if pn.HasPeer(json.Addr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "peer known"})
			return
		}
		go func() {
			err := pn.RetryAddPeer(json.Addr, false)
			if err != nil {
				fmt.Printf("failed to add peer %s: %s", json.Addr, err.Error())
			}
		}()
		c.Status(200)
	})

	r.GET("/peers", func(c *gin.Context) {
		peers := pn.GetAddrsIds()
		c.JSON(http.StatusOK, PeersResp{
			Peers: peers,
		})
	})
}
