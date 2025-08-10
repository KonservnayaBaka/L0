package models

type Item struct {
	ID          int     `json:"id" postgres:"id"`
	OrderUID    string  `json:"order_uid" postgres:"order_uid"`
	ChrtID      int     `json:"chrt_id" postgres:"chrt_id"`
	TrackNumber string  `json:"track_number" postgres:"track_number"`
	Price       float64 `json:"price" postgres:"price"`
	Rid         string  `json:"rid" postgres:"rid"`
	Name        string  `json:"name" postgres:"name"`
	Sale        int     `json:"sale" postgres:"sale"`
	Size        string  `json:"size" postgres:"size"`
	TotalPrice  float64 `json:"total_price" postgres:"total_price"`
	NmID        int     `json:"nm_id" postgres:"nm_id"`
	Brand       string  `json:"brand" postgres:"brand"`
	Status      int     `json:"status" postgres:"status"`
}
