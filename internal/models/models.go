package models

type DataОrder struct {
	UUID        string `db:"correlation_id" json:"result"`
	Code        string `db:"code" json:"code"`
	UserID      string `db:"user_id" json:"user_id"`
	DeletedFlag bool   `db:"is_deleted" json:"is_deleted"`
}

type DataUser struct {
	UUID  string `db:"id" json:"user_id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	Hash  []byte `db:"hash"`
}

type DeleteOrder struct {
	UserID  string   `db:"user_id" json:"user_id"`
	OrderID []string `db:"order_id" json:"order_id"`
}
