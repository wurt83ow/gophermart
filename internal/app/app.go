package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	authz "github.com/wurt83ow/gophermart/internal/authorization"
	"github.com/wurt83ow/gophermart/internal/bdkeeper"
	"github.com/wurt83ow/gophermart/internal/config"
	"github.com/wurt83ow/gophermart/internal/controllers"
	"github.com/wurt83ow/gophermart/internal/logger"
	"github.com/wurt83ow/gophermart/internal/middleware"
	"github.com/wurt83ow/gophermart/internal/storage"
	"github.com/wurt83ow/gophermart/internal/worker"
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

	ctx := context.Background()
	memoryStorage := storage.NewMemoryStorage(keeper, nLogger)

	extcontr := controllers.NewExtController(nLogger)
	extcontr.GetData()
	worker := worker.NewWorker(extcontr, memoryStorage, nLogger)

	authz := authz.NewJWTAuthz(option.JWTSigningKey(), nLogger)
	basecontr := controllers.NewBaseController(memoryStorage, option, nLogger, authz)
	reqLog := middleware.NewReqLog(nLogger)

	worker.Start(ctx)
	r := chi.NewRouter()
	r.Use(reqLog.RequestLogger)
	// r.Use(middleware.GzipMiddleware)

	r.Mount("/", basecontr.Route())

	flagRunAddr := option.RunAddr()
	nLogger.Info("Running server", zap.String("address", flagRunAddr))

	return http.ListenAndServe(flagRunAddr, r)
}
