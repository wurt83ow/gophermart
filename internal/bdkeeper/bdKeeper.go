package bdkeeper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

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

// Close implements storage.Keeper.
func (*BDKeeper) Close() bool {
	panic("unimplemented")
}

// LoadOrders implements storage.Keeper.
func (*BDKeeper) LoadOrders() (map[string]models.DataОrder, error) {
	panic("unimplemented")
}

// LoadUsers implements storage.Keeper.
func (*BDKeeper) LoadUsers() (map[string]models.DataUser, error) {
	panic("unimplemented")
}

// Ping implements storage.Keeper.
func (*BDKeeper) Ping() bool {
	panic("unimplemented")
}

// SaveBatch implements storage.Keeper.
func (*BDKeeper) SaveBatch(map[string]models.DataОrder) error {
	panic("unimplemented")
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
		`INSERT INTO dataurl (
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
	FROM dataurl d	 
	WHERE
		d.original_url = $1`,
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
