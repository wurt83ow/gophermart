package models

import "time"

type Key string

type DataОrder struct {
	UUID   string    `db:"id" json:"id"`
	Number string    `db:"number" json:"number"`
	Date   time.Time `db:"date" json:"date"`
	Status string    `db:"status" json:"status"`
	UserID string    `db:"user_id" json:"user_id"`
}

type DataUser struct {
	UUID  string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"name" json:"email"`
	Hash  []byte `db:"name" json:"hash"`
}

type RequestUser struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type ResponseUser struct {
	Response string `json:"response,omitempty"`
}
