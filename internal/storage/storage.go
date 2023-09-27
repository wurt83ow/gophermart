package storage

import (
	"errors"
	"fmt"
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

// GetBaseConnection implements controllers.Storage.
func (*MemoryStorage) GetBaseConnection() bool {
	panic("unimplemented")
}

// GetOrder implements controllers.Storage.
func (*MemoryStorage) GetOrder(k string) (models.DataОrder, error) {
	panic("unimplemented")
}

func (s *MemoryStorage) GetUser(k string) (models.DataUser, error) {
	s.umx.RLock()
	defer s.umx.RUnlock()

	v, exists := s.users[k]
	if !exists {
		return models.DataUser{}, errors.New("value with such key doesn't exist")
	}

	return v, nil
}

// InsertOrder implements controllers.Storage.
func (*MemoryStorage) InsertOrder(k string, v models.DataОrder) (models.DataОrder, error) {
	panic("unimplemented")
}

func (s *MemoryStorage) InsertUser(k string,
	v models.DataUser) (models.DataUser, error) {

	fmt.Println("8888888888888888888888888888")
	nv, err := s.SaveUser(k, v)
	if err != nil {
		return nv, err
	}

	s.umx.Lock()
	defer s.umx.Unlock()

	s.users[k] = nv

	return nv, nil
}

// SaveOrder implements controllers.Storage.
func (*MemoryStorage) SaveOrder(k string, v models.DataОrder) (models.DataОrder, error) {
	panic("unimplemented")
}

func (s *MemoryStorage) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.SaveUser(k, v)
}
