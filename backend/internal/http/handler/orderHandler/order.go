package orderHandler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"wbL0/internal/models"
	"wbL0/internal/service/orderService"
)

type OrderHandler struct {
	service orderService.OrderServiceInterface
	log     *slog.Logger
}

func NewOrderHandler(service orderService.OrderServiceInterface, log *slog.Logger) *OrderHandler {
	return &OrderHandler{service: service, log: log}
}

// GetOrderInfo godoc
// @Summary      Get information about order
// @Description  Get full information about order to UID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        orderUID   path      string  true  "order UID"
// @Success      200 {object} models.OrderResponse
// @Failure      400 {object} map[string]string "orderUID param is empty"
// @Failure      500 {object} map[string]string "failed to get order"
// @Router       /order/{orderUID} [get]
func (h *OrderHandler) GetOrderInfo(c *gin.Context) {
	orderUID := c.Param("orderUID")
	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderUID param is empty"})
		return
	}

	ctx := c.Request.Context()

	fo, err := h.service.GetOrder(ctx, orderUID)
	if err != nil {
		if errors.Is(err, models.ErrOrderNotFound) {
			h.log.Error("failed to get order", "orderUID", orderUID, "err", err.Error())
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		h.log.Error("failed to get order", "orderUID", orderUID, "err", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	response := models.OrderResponse{
		OrderUID:          fo.Order.OrderUID,
		TrackNumber:       fo.Order.TrackNumber,
		Entry:             fo.Order.Entry,
		Locale:            fo.Order.Locale,
		InternalSignature: fo.Order.InternalSignature,
		CustomerID:        fo.Order.CustomerID,
		DeliveryService:   fo.Order.DeliveryService,
		Shardkey:          fo.Order.Shardkey,
		SmID:              fo.Order.SmID,
		DateCreated:       fo.Order.DateCreated,
		OofShard:          fo.Order.OofShard,
		Delivery: models.DeliveryDTO{
			Name:    fo.Delivery.Name,
			Phone:   fo.Delivery.Phone,
			Zip:     fmt.Sprintf("%d", fo.Delivery.Zip),
			City:    fo.Delivery.City,
			Address: fo.Delivery.Address,
			Region:  fo.Delivery.Region,
			Email:   fo.Delivery.Email,
		},
		Payment: models.PaymentDTO{
			Transaction:  fo.Payment.Transaction,
			RequestID:    fo.Payment.RequestID,
			Currency:     fo.Payment.Currency,
			Provider:     fo.Payment.Provider,
			Amount:       fo.Payment.Amount,
			PaymentDt:    fo.Payment.PaymentDt,
			Bank:         fo.Payment.Bank,
			DeliveryCost: fo.Payment.DeliveryCost,
			GoodsTotal:   fo.Payment.GoodsTotal,
			CustomFee:    fo.Payment.CustomFee,
		},
	}

	for _, item := range fo.Items {
		response.Items = append(response.Items, models.ItemDTO{
			ChrtID:      item.ChrtID,
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			Rid:         item.Rid,
			Name:        item.Name,
			Sale:        item.Sale,
			Size:        item.Size,
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}

	c.JSON(http.StatusOK, response)
}
