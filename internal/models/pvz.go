package models

type City string

const (
	CityMoscow          City = "Москва"
	CitySaintPetersburg City = "Санкт-Петербург"
	CityKazan           City = "Казань"
)

type PVZ struct {
	ID               string `json:"id"`
	RegistrationDate string `json:"registrationDate"`
	City             City   `json:"city"`
}

type PVZWithDetails struct {
	PVZ        *PVZ
	Receptions []*ReceptionWithProducts
}
