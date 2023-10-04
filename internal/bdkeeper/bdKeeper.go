package bdkeeper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Info(string, ...zapcore.Field)
}

type BDKeeper struct {
	conn *sql.DB
	log  Log
}

func NewBDKeeper(dsn func() string, log Log) *BDKeeper {
	addr := dsn()
	if addr == "" {
		log.Info("database dsn is empty")
		return nil
	}

	conn, err := sql.Open("pgx", dsn())
	if err != nil {

		log.Info("Unable to connection to database: ", zap.Error(err))
		return nil
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		log.Info("error getting driver: ", zap.Error(err))
		return nil
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Info("error getting getwd: ", zap.Error(err))
	}

	// fix error test path
	mp := dir + "/migrations"
	var path string
	if _, err := os.Stat(mp); err != nil {
		path = "../../"
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%smigrations", path),
		"postgres",
		driver)

	if err != nil {
		log.Info("Error creating migration instance : ", zap.Error(err))
	}

	err = m.Up()
	if err != nil {
		log.Info("Error while performing migration: ", zap.Error(err))
	}

	log.Info("Connected!")

	return &BDKeeper{
		conn: conn,
		log:  log,
	}
}

func (bdk *BDKeeper) GetOpenOrders() ([]string, error) {
	ctx := context.Background()

	// get orders from bd
	rows, err := bdk.conn.QueryContext(ctx, `
	SELECT number
	FROM public.orders
	WHERE status <> 'INVALID'
	AND status <> 'PROCESSED'
	LIMIT 100
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := make([]string, 0)
	for rows.Next() {

		var order string
		err := rows.Scan(order)
		if err != nil {
			log.Fatal(err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (bdk *BDKeeper) LoadOrders() (storage.StorageOrders, error) {
	ctx := context.Background()

	// get orders from bd
	rows, err := bdk.conn.QueryContext(ctx, `
	SELECT 
		o.id,
		o.number,
		o.status,
		o.date,
		'' AS date_rfc,				
		COALESCE(s.accrual, 0) AS accrual,
		o.user_id	
	FROM orders AS o
	LEFT JOIN savings_account AS s 
	ON o.id = s.id_order_in
		AND o.date = s.processed_at`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	data := make(storage.StorageOrders)
	for rows.Next() {
		record := models.DataОrder{}

		s := reflect.ValueOf(&record).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Fatal(err)
		}
		data[record.Number] = record
	}

	if err = rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

func (bdk *BDKeeper) LoadUsers() (storage.StorageUsers, error) {
	ctx := context.Background()

	// get users from bd
	rows, err := bdk.conn.QueryContext(ctx, `SELECT id, name, email, hash FROM users`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	data := make(storage.StorageUsers)
	for rows.Next() {
		record := models.DataUser{}

		s := reflect.ValueOf(&record).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Fatal(err)
		}
		data[record.Email] = record
	}

	if err = rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

func (bdk *BDKeeper) SaveOrder(key string, order models.DataОrder) (models.DataОrder, error) {
	ctx := context.Background()

	var id string
	if order.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = order.UUID
	}
	_, err := bdk.conn.ExecContext(ctx,
		`INSERT INTO orders (
			id,
			number,
			date,
			status,
			user_id)
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING user_id`,
		id, order.Number, order.Date, order.Status, order.UserID)

	row := bdk.conn.QueryRowContext(ctx, `
	SELECT
		d.id,
		d.number,
		d.date,
		d.status,
		d.user_id	 
	FROM orders d	 
	WHERE
		d.number = $1`,
		order.Number,
	)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataОrder
	nerr := row.Scan(&m.UUID, &m.Number, &m.Date, &m.Status, &m.UserID)
	if nerr != nil {
		bdk.log.Info("row scan error: ", zap.Error(err))
		return order, nerr
	}

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			bdk.log.Info("unique field violation on column: ", zap.Error(err))

			return m, storage.ErrConflict
		}
		return m, err
	}

	return m, nil
}

func (bdk *BDKeeper) SaveUser(key string, data models.DataUser) (models.DataUser, error) {
	ctx := context.Background()

	var id string
	if data.UUID == "" {
		neuuid := uuid.New()
		id = neuuid.String()
	} else {
		id = data.UUID
	}

	_, err := bdk.conn.ExecContext(ctx,
		`INSERT INTO users (
			id,
			email,
			hash,
			name)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		id, data.Email, data.Hash, data.Name)

	var (
		cond string
		hash []byte
	)

	if data.Hash != nil {
		cond = "AND u.hash = $2"
		hash = data.Hash
	}

	stmt := fmt.Sprintf(`
	SELECT
		u.id,
		u.email,
		u.hash,
		u.name  	 
	FROM users u	 
	WHERE
		u.email = $1 %s`, cond)
	row := bdk.conn.QueryRowContext(ctx, stmt, data.Email, hash)

	// read the values from the database record into the corresponding fields of the structure
	var m models.DataUser
	nerr := row.Scan(&m.UUID, &m.Email, &m.Hash, &m.Name)
	if nerr != nil {
		return data, nerr
	}

	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			bdk.log.Info("unique field violation on column: ", zap.Error(err))

			return m, storage.ErrConflict
		}
		return m, err
	}

	return m, nil
}

func (bdk *BDKeeper) UpdateOrderStatus(result []models.ExtRespOrder) error {

	ctx := context.Background()

	valueStrings := make([]string, 0, len(result))
	valueArgs := make([]interface{}, 0, len(result)*2)

	for i, v := range result {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, v.Order)
		valueArgs = append(valueArgs, v.Status)
	}

	stmt := fmt.Sprintf(
		`WITH _data (number, status) 
		AS (VALUES %s)
		UPDATE orders AS o
		SET o.status = _data.status
		FROM _data
		WHERE o.number = _data.number`,
		strings.Join(valueStrings, ","))
	_, err := bdk.conn.ExecContext(ctx, stmt, valueArgs...)

	if err != nil {
		return err
	}

	return nil
}

func (bdk *BDKeeper) InsertAccruel(result []models.ExtRespOrder) error {
	ctx := context.Background()

	valueStrings := make([]string, 0, len(result))
	valueArgs := make([]interface{}, 0, len(result)*2)

	for i, v := range result {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, v.Order)
		valueArgs = append(valueArgs, v.Accrual)
	}

	stmt := fmt.Sprintf(
		`WITH _data (number, accrual) 
			AS (VALUES %s)
		INSERT INTO savings_account (user_id, processed_at, id_order_in,  accrual)
		SELECT orders.user_id, current_timestamp, _data.number, _data.accrual   
		FROM _data 
		INNER JOIN orders 
			ON _data.number = orders.number`,
		strings.Join(valueStrings, ","))
	_, err := bdk.conn.ExecContext(ctx, stmt, valueArgs...)

	if err != nil {
		return err
	}

	return nil
}

func (bdk *BDKeeper) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := bdk.conn.PingContext(ctx); err != nil {
		return false
	}

	return true
}

func (bdk *BDKeeper) Close() bool {
	bdk.conn.Close()

	return true
}
