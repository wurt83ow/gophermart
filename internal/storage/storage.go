package storage

import (
	"errors"
	"sort"
	"sync"
	"time"

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
	GetOpenOrders() ([]string, error)
	UpdateOrderStatus(result []models.ExtRespOrder) error
	InsertAccruel(map[string]models.ExtRespOrder) error
	ExecuteWithdraw(models.RequestWithdraw) error
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

func (s *MemoryStorage) UpdateOrderStatus(result []models.ExtRespOrder) error {
	err := s.keeper.UpdateOrderStatus(result)
	if err != nil {
		return err
	}

	for _, v := range result {
		o, exists := s.orders[v.Order]

		if exists {
			s.orders[v.Order] = models.DataОrder{
				UUID:    o.UUID,
				Number:  o.Number,
				Status:  v.Status,
				Date:    o.Date,
				DateRFC: o.DateRFC,
				Accrual: v.Accrual,
				UserID:  o.UserID,
			}
		}
	}

	return nil
}

func (s *MemoryStorage) InsertAccruel(orders map[string]models.ExtRespOrder) error {
	return s.keeper.InsertAccruel(orders)
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

func (s *MemoryStorage) GetOpenOrders() ([]string, error) {
	orders, err := s.keeper.GetOpenOrders()
	if err != nil {
		return nil, err
	}

	return orders, nil
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

func (s *MemoryStorage) GetUserOrders(userID string) []models.DataОrder {
	// var orders []models.DataОrder
	orders := make([]models.DataОrder, 0)

	s.omx.RLock()
	defer s.omx.RUnlock()

	for _, o := range s.orders {
		if o.UserID != userID {
			continue
		}

		o.DateRFC = o.Date.Format(time.RFC3339)
		// o.Accrual = 0
		// o.Status = "PROCESSED"
		orders = append(orders, o)
	}

	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].Date.After(orders[j].Date)
	})

	return orders
}

func (s *MemoryStorage) ExecuteWithdraw(withdraw models.RequestWithdraw) error {
	return s.keeper.ExecuteWithdraw(withdraw)
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

func (s *MemoryStorage) GetBaseConnection() bool {
	if s.keeper == nil {
		return false
	}

	return s.keeper.Ping()
}
