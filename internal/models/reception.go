package models

type ReceptionStatus string

const (
	ReceptionStatusInProgress ReceptionStatus = "in_progress"
	ReceptionStatusClose      ReceptionStatus = "close"
)

type Reception struct {
	ID       string          `json:"id"`
	DateTime string          `json:"dateTime"`
	PVZID    string          `json:"pvzId"`
	Status   ReceptionStatus `json:"status"`
}
