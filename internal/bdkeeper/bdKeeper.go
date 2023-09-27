package bdkeeper

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/wurt83ow/gophermart/internal/models"
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

	fmt.Println("77777777777777", dsn())
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

// SaveOrders implements storage.Keeper.
func (*BDKeeper) SaveOrders(string, models.DataОrder) (models.DataОrder, error) {
	panic("unimplemented")
}

// SaveUser implements storage.Keeper.
func (*BDKeeper) SaveUser(string, models.DataUser) (models.DataUser, error) {
	panic("unimplemented")
}

// UpdateBatch implements storage.Keeper.
func (*BDKeeper) UpdateBatch(...models.DeleteOrder) error {
	panic("unimplemented")
}
