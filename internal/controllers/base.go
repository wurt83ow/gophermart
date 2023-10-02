package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	authz "github.com/wurt83ow/gophermart/internal/authorization"
	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var keyUserID models.Key = "userID"

type IExternalClient interface {
	GetData() (string, error)
}

type Storage interface {
	InsertOrder(k string, v models.DataОrder) (models.DataОrder, error)
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
	GetUser(k string) (models.DataUser, error)
	GetUserOrders(userID string) []models.DataОrder
	SaveOrder(k string, v models.DataОrder) (models.DataОrder, error)
	SaveUser(k string, v models.DataUser) (models.DataUser, error)
	GetBaseConnection() bool
}

type Options interface {
	ParseFlags()
	RunAddr() string
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type Authz interface {
	JWTAuthzMiddleware(authz.Storage, authz.Log) func(http.Handler) http.Handler
	GetHash(email string, password string) []byte
	CreateJWTTokenForUser(userid string) string
	AuthCookie(name string, token string) *http.Cookie
}

type BaseController struct {
	storage Storage
	options Options
	log     Log
	authz   Authz
}

func NewBaseController(storage Storage, options Options, log Log, authz Authz) *BaseController {
	instance := &BaseController{
		storage: storage,
		options: options,
		log:     log,
		authz:   authz,
	}

	return instance
}

func (h *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)
	r.Get("/ping", h.getPing)

	// group where the middleware authorization is needed
	r.Group(func(r chi.Router) {
		r.Use(h.authz.JWTAuthzMiddleware(h.storage, h.log))

		r.Post("/api/user/orders", h.createOrder)
		r.Get("/api/user/orders", h.getUserOrders)

	})

	return r
}

func (h *BaseController) Register(w http.ResponseWriter, r *http.Request) {

	regReq := models.RequestUser{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&regReq); err != nil {
		h.log.Info("cannot decode request JSON body: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := h.storage.GetUser(regReq.Email)
	if err == nil {
		h.log.Info("the user is already registered: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	Hash := h.authz.GetHash(regReq.Email, regReq.Password)

	// save the user to the storage
	dataUser := models.DataUser{UUID: uuid.New().String(), Email: regReq.Email, Hash: Hash, Name: regReq.Name}

	_, err = h.storage.InsertUser(regReq.Email, dataUser)

	if err != nil {
		if err == storage.ErrConflict {
			w.WriteHeader(http.StatusConflict) //code 409
		} else {
			w.WriteHeader(http.StatusBadRequest) // code 400
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	h.log.Info("sending HTTP 201 response")
}

func (h *BaseController) Login(w http.ResponseWriter, r *http.Request) {
	var rb models.RequestUser
	if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUser(rb.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if bytes.Equal(user.Hash, h.authz.GetHash(rb.Email, rb.Password)) {
		freshToken := h.authz.CreateJWTTokenForUser(user.UUID)
		http.SetCookie(w, h.authz.AuthCookie("jwt-token", freshToken))
		http.SetCookie(w, h.authz.AuthCookie("Authorization", freshToken))

		w.Header().Set("Authorization", freshToken)
		err := json.NewEncoder(w).Encode(models.ResponseUser{
			Response: "success",
		})

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		return
	}

	err = json.NewEncoder(w).Encode(models.ResponseUser{
		Response: "incorrect email/password",
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// POST
func (h *BaseController) createOrder(w http.ResponseWriter, r *http.Request) {

	metod := zap.String("method", r.Method)
	userID, StatusOK := r.Context().Value(keyUserID).(string)
	if !StatusOK || userID == "" {
		// user is not authenticated
		w.WriteHeader(http.StatusUnauthorized) //code 401
		h.log.Info("user is not authenticated, request status 401: %v", metod)
		return
	}

	// set the correct header for the data type
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		// invalid request format
		w.WriteHeader(http.StatusBadRequest) //code 400
		h.log.Info("invalid request format, request status 400: %v", metod)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	orderNum := string(body)

	ord, err := strconv.Atoi(orderNum)
	if err != nil || !Valid(ord) {
		// incorrect order number format
		w.WriteHeader(http.StatusUnprocessableEntity) //code 422
		h.log.Info("incorrect order number format, request status 422: %v", metod)
		return
	}

	curDate := time.Now()
	status := "NEW"

	// save full url to storage with the key received earlier
	order, err := h.storage.InsertOrder(orderNum, models.DataОrder{
		Number: orderNum, Date: curDate, Status: status, UserID: userID})

	if err != nil {
		if err == storage.ErrConflict {
			// The order number has already been uploaded
			if order.UserID == userID {
				// this user
				w.WriteHeader(http.StatusOK) //code 200
			} else {
				// another user
				w.WriteHeader(http.StatusConflict) //code 409
				h.log.Info(`The order number has already been uploaded 
					another user, request status 409: %v`, metod)
			}
			return
		} else {
			// internal server error
			w.WriteHeader(http.StatusInternalServerError) //code 500
			h.log.Info("internal server error, request status 500: %v", metod)
			return
		}
	}

	// new order number accepted for processing
	w.WriteHeader(http.StatusAccepted) //code 202
}

// GET
func (h *BaseController) getUserOrders(w http.ResponseWriter, r *http.Request) {
	metod := zap.String("method", r.Method)

	userID, ok := r.Context().Value(keyUserID).(string)
	if !ok {
		// user is not authorized
		w.WriteHeader(http.StatusUnauthorized) //401
		h.log.Info("user is not authenticated, request status 401: %v", metod)
		return
	}

	orders := h.storage.GetUserOrders(userID)
	if len(orders) == 0 {
		// no information to answer
		w.WriteHeader(http.StatusNoContent) // 204
		h.log.Info("no information to answer, request status 204: %v", metod)
		return
	}

	// serialize the server response
	enc := json.NewEncoder(w)
	if err := enc.Encode(orders); err != nil {
		// Internal Server Error
		w.WriteHeader(http.StatusInternalServerError) //code 500
		h.log.Info("Internal Server Error: ", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) //code 200
}

// GET
func (h *BaseController) getPing(w http.ResponseWriter, r *http.Request) {
	if !h.storage.GetBaseConnection() {
		h.log.Info("got status internal server error")
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	w.WriteHeader(http.StatusOK) // 200
	h.log.Info("sending HTTP 200 response")
}

// Valid check number is valid or not based on Luhn algorithm
func Valid(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
