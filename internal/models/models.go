package models

import "time"

type LawFirm struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ServeRequest struct {
	ID        int       `json:"id" db:"id"`
	LawFirmID int       `json:"law_firm_id" db:"law_firm_id"`
	Defendant string    `json:"defendant" db:"defendant"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
