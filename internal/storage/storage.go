package storage

import (
	"errors"
	"sync"

	"github.com/wurt83ow/gophermart/internal/models"
	"go.uber.org/zap"
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
	omx    sync.RWMutex
	umx    sync.RWMutex
	orders StorageOrders
	users  StorageUsers
	keeper Keeper
	log    Log
}

type Keeper interface {
	LoadOrders() (StorageOrders, error)
	LoadUsers() (StorageUsers, error)
	SaveOrder(string, models.DataОrder) (models.DataОrder, error)
	SaveUser(string, models.DataUser) (models.DataUser, error)
	SaveBatch(StorageOrders) error

	Ping() bool
	Close() bool
}

func NewMemoryStorage(keeper Keeper, log Log) *MemoryStorage {
	orders := make(StorageOrders)
	users := make(StorageUsers)

	if keeper != nil {
		var err error
		orders, err = keeper.LoadOrders()
		if err != nil {
			log.Info("cannot load url data: ", zap.Error(err))
		}

		users, err = keeper.LoadUsers()
		if err != nil {
			log.Info("cannot load user data: ", zap.Error(err))
		}
	}

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

func (s *MemoryStorage) InsertOrder(k string,
	v models.DataОrder) (models.DataОrder, error) {

	nv, err := s.SaveOrder(k, v)
	if err != nil {
		return nv, err
	}

	s.omx.Lock()
	defer s.omx.Unlock()

	s.orders[k] = nv

	return nv, nil
}

func (s *MemoryStorage) InsertUser(k string,
	v models.DataUser) (models.DataUser, error) {

	nv, err := s.SaveUser(k, v)
	if err != nil {
		return nv, err
	}

	s.umx.Lock()
	defer s.umx.Unlock()

	s.users[k] = nv

	return nv, nil
}

func (s *MemoryStorage) SaveOrder(k string, v models.DataОrder) (models.DataОrder, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.SaveOrder(k, v)
}

func (s *MemoryStorage) SaveUser(k string, v models.DataUser) (models.DataUser, error) {
	if s.keeper == nil {
		return v, nil
	}

	return s.keeper.SaveUser(k, v)
}
