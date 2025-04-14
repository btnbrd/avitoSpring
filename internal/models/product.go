package models

type ProductType string

const (
	ProductTypeElectronics ProductType = "электроника"
	ProductTypeClothing    ProductType = "одежда"
	ProductTypeFootwear    ProductType = "обувь"
)

type Product struct {
	ID          string      `json:"id"`
	DateTime    string      `json:"dateTime"`
	Type        ProductType `json:"type"`
	ReceptionID string      `json:"receptionId"`
}
