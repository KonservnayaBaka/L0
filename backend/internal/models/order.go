package models

import (
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" postgres:"order_uid"` //В тестовых данных UUID странный, решил избежать использование uuid.UUID
	TrackNumber       string    `json:"track_number" postgres:"track_number"`
	Entry             string    `json:"entry" postgres:"entry"`
	Locale            string    `json:"locale" postgres:"locale"`
	InternalSignature string    `json:"internal_signature" postgres:"internal_signature"`
	CustomerID        string    `json:"customer_id" postgres:"customer_id"`
	DeliveryService   string    `json:"delivery_service" postgres:"delivery_service"`
	Shardkey          string    `json:"shardkey" postgres:"shardkey"`
	SmID              int       `json:"sm_id" postgres:"sm_id"`
	DateCreated       time.Time `json:"date_created" postgres:"date_created"`
	OofShard          string    `json:"oof_shard" postgres:"oof_shard"`
}
