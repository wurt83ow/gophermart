package models

import "time"

type Key string

type DataОrder struct {
	UUID    string    `db:"id" json:"-"`
	Number  string    `db:"number" json:"number"`
	Status  string    `db:"status" json:"status"`
	Date    time.Time `db:"date" json:"-"`
	DateRFC string    `db:"date_rfc" json:"uploaded_at"`
	Accrual float64   `db:"accrual" json:"accrual,omitempty"`
	UserID  string    `db:"user_id" json:"-"`
}

type DataUser struct {
	UUID  string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Email string `db:"name" json:"email"`
	Hash  []byte `db:"name" json:"hash"`
}

type RequestUser struct {
	Email    string `json:"login"`
	Password string `json:"password"`
}

type ResponseUser struct {
	Response string `json:"response,omitempty"`
}

type ExtRespOrder struct {
	Order   string  `db:"order" json:"order"`
	Status  string  `db:"status" json:"status"`
	Accrual float64 `db:"accrual" json:"accrual,omitempty"`
}

type BDOrder struct {
	Order string `db:"order" `
}
