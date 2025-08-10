package models

type Delivery struct {
	OrderUID string `json:"order_uid" postgres:"order_uid"`
	Name     string `json:"name" postgres:"name"`
	Phone    string `json:"phone" postgres:"phone"`
	Zip      int    `json:"zip" postgres:"zip"`
	City     string `json:"city" postgres:"city"`
	Address  string `json:"address" postgres:"address"`
	Region   string `json:"region" postgres:"region"`
	Email    string `json:"email" postgres:"email"`
}
