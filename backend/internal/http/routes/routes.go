package routes

import (
	"github.com/gin-gonic/gin"
	"wbL0/internal/http/handler/orderHandler"
)

func InitRoutes(r *gin.Engine, orderHandler orderHandler.OrderHandler) {
	orderGroup := r.Group("/order")
	{
		orderGroup.GET("/:orderUID", orderHandler.GetOrderInfo)
	}
}
