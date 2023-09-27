package storage

import (
	"errors"
	"sync"

	"github.com/wurt83ow/gophermart/internal/models"
	"go.uber.org/zap/zapcore"
)

// ErrConflict indicates a data conflict in the store.
var ErrConflict = errors.New("data conflict")

type StorageOrders = map[string]models.DataОrder
type StorageUsers = map[string]models.DataUser

type Log interface {
	Info(string, ...zapcore.Field)
}

type MemoryStorage struct {
	dmx    sync.RWMutex
	umx    sync.RWMutex
	orders StorageOrders
	users  StorageUsers
	keeper Keeper
	log    Log
}

type Keeper interface {
	LoadOrders() (StorageOrders, error)
	LoadUsers() (StorageUsers, error)
	SaveOrders(string, models.DataОrder) (models.DataОrder, error)
	SaveUser(string, models.DataUser) (models.DataUser, error)
	SaveBatch(StorageOrders) error
	UpdateBatch(...models.DeleteOrder) error
	Ping() bool
	Close() bool
}

func NewMemoryStorage(keeper Keeper, log Log) *MemoryStorage {
	orders := make(StorageOrders)
	users := make(StorageUsers)

	// if keeper != nil {
	// 	var err error
	// 	orders, err = keeper.Load()
	// 	if err != nil {
	// 		log.Info("cannot load url data: ", zap.Error(err))
	// 	}

	// 	users, err = keeper.LoadUsers()
	// 	if err != nil {
	// 		log.Info("cannot load user data: ", zap.Error(err))
	// 	}
	// }

	return &MemoryStorage{
		orders: orders,
		users:  users,
		keeper: keeper,
		log:    log,
	}
}
