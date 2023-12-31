package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	authz "github.com/wurt83ow/gophermart/internal/authorization"
	"github.com/wurt83ow/gophermart/internal/middleware"
	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/storage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Storage interface {
	InsertOrder(k string, v models.DataОrder) (models.DataОrder, error)
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
	GetOrder(k string) (models.DataОrder, error)
	GetUser(k string) (models.DataUser, error)
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

type BaseController struct {
	storage Storage
	options Options
	log     Log
}

func NewBaseController(storage Storage, options Options, log Log) *BaseController {
	instance := &BaseController{
		storage: storage,
		options: options,
		log:     log,
	}

	return instance
}

func (h *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Get("/ping", h.getPing)

	// group where the middleware authorization is needed
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuthzMiddleware(h.storage, h.log))

		// r.Post("/", h.shortenURL)
		// r.Get("/api/user/urls", h.getUserURLs)

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
	fmt.Println("99999999999999999999999")
	_, err := h.storage.GetUser(regReq.Email)
	if err == nil {
		h.log.Info("the user is already registered: ", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	Hash := authz.GetHash(regReq.Email, regReq.Password)

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

	if bytes.Equal(user.Hash, authz.GetHash(rb.Email, rb.Password)) {
		freshToken := authz.CreateJWTTokenForUser(user.UUID)
		http.SetCookie(w, authz.AuthCookie("jwt-token", freshToken))
		http.SetCookie(w, authz.AuthCookie("Authorization", freshToken))

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
