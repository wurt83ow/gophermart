package app

import (
	"net/http"

	"github.com/go-chi/chi"
	authz "github.com/wurt83ow/gophermart/internal/authorization"
	"github.com/wurt83ow/gophermart/internal/bdkeeper"
	"github.com/wurt83ow/gophermart/internal/config"
	"github.com/wurt83ow/gophermart/internal/controllers"
	"github.com/wurt83ow/gophermart/internal/logger"
	"github.com/wurt83ow/gophermart/internal/middleware"
	"github.com/wurt83ow/gophermart/internal/storage"
	"github.com/wurt83ow/gophermart/internal/workerpool"
	"go.uber.org/zap"
)

func Run() error {
	// create and initialize a new option instance
	option := config.NewOptions()
	option.ParseFlags()

	// get a new logger
	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		return err
	}

	// initialize the keeper instance
	var keeper storage.Keeper
	if option.DataBaseDSN() != "" {
		keeper = bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
		defer keeper.Close()
	}

	// initialize the storage instance
	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	// create a new controller for creating outgoing requests
	extcontr := controllers.NewExtController(memoryStorage,
		option.AccrualSystemAddress, nLogger)

	// create a new workerpool for concurrency task processing
	var allTask []*workerpool.Task
	pool := workerpool.NewPool(allTask, option.Concurrency,
		extcontr, memoryStorage, nLogger)

	// create a new NewJWTAuthz for user authorization
	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)

	// create a new controller to process incoming requests
	basecontr := controllers.NewBaseController(memoryStorage, option,
		nLogger, authz)

	// get a middleware for logging requests
	reqLog := middleware.NewReqLog(nLogger)

	// start the worker pool in the background
	go pool.RunBackground()

	r := chi.NewRouter()
	r.Use(reqLog.RequestLogger)
	// r.Use(middleware.GzipMiddleware)

	r.Mount("/", basecontr.Route())

	flagRunAddr := option.RunAddr()
	nLogger.Info("Running server", zap.String("address", flagRunAddr))

	return http.ListenAndServe(flagRunAddr, r)
}
