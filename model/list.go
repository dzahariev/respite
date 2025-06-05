package model

// List holds technical fields
type List struct {
	PageSize int      `json:"page_size,omitempty"`
	Page     int      `json:"page,omitempty"`
	Count    int64    `json:"count"`
	Data     []Object `json:"data"`
}
