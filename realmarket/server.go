package realmarket

import (
	"github.com/gin-gonic/gin"
)

func WebsocketServerStart() {
	r := gin.Default()
	r.GET("/realorderbook", realOrderBookHandler)
	r.GET("/realmarkethistory", realMarketHistoryHandler)
	r.OPTIONS("/order", CORSMiddleware())
	r.POST("/order", CORSMiddleware(), addOrderHandler)
	r.PUT("/order", CORSMiddleware(), modifyOrderHandler)
	r.DELETE("/order", CORSMiddleware(), cancelOrderHandler)
	r.GET("/order", CORSMiddleware(), getOrderHandler)
	r.Run(":8000") // listen and serve on localhost:8000
}
