package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/wurt83ow/gophermart/internal/accruel"
	authz "github.com/wurt83ow/gophermart/internal/authorization"
	"github.com/wurt83ow/gophermart/internal/bdkeeper"
	"github.com/wurt83ow/gophermart/internal/config"
	"github.com/wurt83ow/gophermart/internal/controllers"
	"github.com/wurt83ow/gophermart/internal/logger"
	"github.com/wurt83ow/gophermart/internal/middleware"
	"github.com/wurt83ow/gophermart/internal/storage"
	"github.com/wurt83ow/gophermart/internal/workerpool"
)

type Server struct {
	srv *http.Server
	ctx context.Context
	// db  *pgxpool.Pool
}

func NewServer(ctx context.Context) *Server {
	server := new(Server)
	server.ctx = ctx

	return server
}

func (server *Server) Serve() {
	// create and initialize a new option instance
	option := config.NewOptions()
	option.ParseFlags()

	// get a new logger
	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		log.Fatalln(err)
	}

	// initialize the keeper instance
	keeper := initializeKeeper(option.DataBaseDSN, nLogger)
	defer keeper.Close()

	// initialize the storage instance
	memoryStorage := initializeStorage(keeper, nLogger)

	// create a new workerpool for concurrency task processing
	pool := initializeWorkerPool(option.Concurrency, nLogger, option.TaskExecutionInterval)

	// create a new NewJWTAuthz for user authorization
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)

	// create a new controller to process incoming requests
	basecontr := initializeBaseController(memoryStorage, option, nLogger, authz)

	// get a middleware for logging requests
	reqLog := middleware.NewReqLog(nLogger)

	// start the worker pool in the background
	go pool.RunBackground()

	// create a new controller for creating outgoing requests
	extcontr := controllers.NewExtController(memoryStorage, option.AccrualSystemAddress, nLogger)

	// create and start accrual service
	accruelServise := initializeAccrualService(extcontr, pool, memoryStorage, nLogger, option.TaskExecutionInterval)
	accruelServise.Start()

	// create router and mount routes
	r := chi.NewRouter()
	r.Use(reqLog.RequestLogger)
	r.Mount("/", basecontr.Route())

	// configure and start the server
	startServer(r, option.RunAddr())
}

func initializeKeeper(dataBaseDSN func() string, logger *logger.Logger) *bdkeeper.BDKeeper {
	if dataBaseDSN() == "" {
		return nil
	}

	return bdkeeper.NewBDKeeper(dataBaseDSN, logger)
}

func initializeStorage(keeper storage.Keeper, logger *logger.Logger) *storage.MemoryStorage {
	if keeper == nil {
		return nil
	}

	return storage.NewMemoryStorage(keeper, logger)
}

func initializeWorkerPool(concurrency func() string, logger *logger.Logger, interval func() string) *workerpool.Pool {
	var allTask []*workerpool.Task

	return workerpool.NewPool(allTask, concurrency, logger, interval)
}

func initializeBaseController(storage *storage.MemoryStorage, option *config.Options,
	logger *logger.Logger, authz *authz.JWTAuthz,
) *controllers.BaseController {
	return controllers.NewBaseController(storage, option, logger, authz)
}

func initializeAccrualService(extController *controllers.ExtController, pool *workerpool.Pool,
	storage *storage.MemoryStorage, logger *logger.Logger, interval func() string,
) *accruel.AccrualService {
	return accruel.NewAccrualService(extController, pool, storage, logger, interval)
}

func startServer(router chi.Router, address string) {
	const (
		oneMegabyte = 1 << 20
		readTimeout = 3 * time.Second
	)

	server := &http.Server{
		Addr:                         address,
		Handler:                      router,
		ReadHeaderTimeout:            readTimeout,
		WriteTimeout:                 readTimeout,
		IdleTimeout:                  readTimeout,
		ReadTimeout:                  readTimeout,
		MaxHeaderBytes:               oneMegabyte, // 1 MB
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}

func (server *Server) Shutdown() {
	log.Printf("server stopped")

	const shutdownTimeout = 5 * time.Second
	ctxShutDown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer cancel()

	if err := server.srv.Shutdown(ctxShutDown); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			//nolint:gocritic
			log.Fatalf("server Shutdown Failed:%s", err)
		}
	}

	log.Println("server exited properly")
}
