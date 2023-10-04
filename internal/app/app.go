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
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		return err
	}

	var keeper storage.Keeper
	if option.DataBaseDSN() != "" {
		keeper = bdkeeper.NewBDKeeper(option.DataBaseDSN, nLogger)
		defer keeper.Close()
	}

	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	extcontr := controllers.NewExtController(memoryStorage, nLogger)

	var allTask []*workerpool.Task
	pool := workerpool.NewPool(allTask, 10, extcontr, memoryStorage) //!!! вынести в config кол. воркеров

	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	basecontr := controllers.NewBaseController(memoryStorage, option, nLogger, authz)
	reqLog := middleware.NewReqLog(nLogger)
	pool.RunBackground()

	r := chi.NewRouter()
	r.Use(reqLog.RequestLogger)
	// r.Use(middleware.GzipMiddleware)

	r.Mount("/", basecontr.Route())

	flagRunAddr := option.RunAddr()
	nLogger.Info("Running server", zap.String("address", flagRunAddr))

	return http.ListenAndServe(flagRunAddr, r)
}
