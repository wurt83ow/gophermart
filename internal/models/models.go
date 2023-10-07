package models

import "time"

type Key string

type DataОrder struct {
	UUID    string    `db:"id" json:"-"`
	Number  string    `db:"number" json:"number"`
	Status  string    `db:"status" json:"status"`
	Date    time.Time `db:"date" json:"-"`
	DateRFC string    `db:"date_rfc" json:"uploaded_at"`
	Accrual float32   `db:"accrual" json:"accrual,omitempty"`
	UserID  string    `db:"user_id" json:"-"`
}

type DataОrderBD struct {
	UserID      string    `db:"user_id" json:"-"`
	Number      string    `db:"number" json:"number"`
	Date        time.Time `db:"date" json:"-"`
	Accrual     float32   `db:"accrual" json:"accrual,omitempty"`
	UserAccrual float32   `db:"user_accrual" json:"user_accrual,omitempty"`
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
	Accrual float32 `db:"accrual" json:"accrual,omitempty"`
}

type BDOrder struct {
	Order string `db:"order" `
}

type BDAccrual struct {
	Accrual float32 `db:"accrual" json:"accrual,omitempty"`
}

type RequestWithdraw struct {
	UserID string  `db:"user_id" json:"user_id"`
	Order  string  `db:"order" json:"order"`
	Sum    float32 `db:"sum" json:"sum"`
}
