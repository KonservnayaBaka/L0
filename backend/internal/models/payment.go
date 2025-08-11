package models

type Payment struct {
	OrderUID     string  `json:"order_uid" postgres:"order_uid"`
	Transaction  string  `json:"transaction" postgres:"transaction"`
	RequestID    string  `json:"request_id" postgres:"request_id"`
	Currency     string  `json:"currency" postgres:"currency"`
	Provider     string  `json:"provider" postgres:"provider"`
	Amount       int     `json:"amount" postgres:"amount"`
	PaymentDt    int     `json:"payment_dt" postgres:"payment_dt"`
	Bank         string  `json:"bank" postgres:"bank"`
	DeliveryCost float64 `json:"delivery_cost" postgres:"delivery_cost"`
	GoodsTotal   float64 `json:"goods_total" postgres:"goods_total"`
	CustomFee    float64 `json:"custom_fee" postgres:"custom_fee"`
}
