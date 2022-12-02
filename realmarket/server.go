package realmarket

import (
	"github.com/gin-gonic/gin"
)

func WebsocketServerStart() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/orderbook", realOrderBookHandler)
	r.GET("/markethistory", realMarketHistoryHandler)
	r.OPTIONS("/order", CORSMiddleware())
	r.POST("/order", CORSMiddleware(), addOrderHandler)
	r.PUT("/order", CORSMiddleware(), modifyOrderHandler)
	r.DELETE("/order", CORSMiddleware(), cancelOrderHandler)
	r.GET("/order", CORSMiddleware(), getOrderHandler)
	r.StaticFile("/apidoc", "./resources/NUSWAP_APIs.xlsx")
	r.StaticFile("/connector", "./resources/API.py")
	userGroup := r.Group("/user")
	{
		userGroup.OPTIONS("/balance", CORSMiddleware())
		userGroup.POST("/balance", CORSMiddleware(), topupHandler)
		userGroup.GET("/balance", CORSMiddleware(), balanceHandler)
		userGroup.GET("/assets", CORSMiddleware(), assetsHandler)
		userGroup.GET("/pltrend", CORSMiddleware(), PLTrendHandler)
		userGroup.GET("/cumulativepl", CORSMiddleware(), cumulativePLHandler)
		userGroup.GET("/learnstage", CORSMiddleware(), currentLearnStageHandler)
	}
	r.Run(":8000") // listen and serve on localhost:8000
}
