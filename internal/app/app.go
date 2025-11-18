package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"app/internal/config"
	"app/internal/controllers"
	"app/internal/controllers/gen"
	"app/internal/repository/cache/redis"
	"app/internal/repository/storage/postgres"
	"app/internal/usecase/pr_usecase"
	"app/internal/usecase/stats_usecase"
	"app/internal/usecase/team_usecase"
	"app/internal/usecase/user_usecase"
	"app/pkg/closer"
	"app/pkg/logger"
	"app/pkg/txmanager"
)

type Server struct {
	closer       *closer.Closer
	router       *gin.Engine
	pgPool       *pgxpool.Pool
	config       *config.Config
	httpServer   *http.Server
	logger       logger.Logger
}

func NewServer(cfg *config.Config, logger logger.Logger) *Server {

	c := closer.NewCloser()

	pgPool, err := cfg.Storage.ConnectionToPostgres(logger)
	if err != nil {
		logger.Fatalw("Connect to PostgreSQL", "error", err)
		return nil
	}
	c.Add(func(ctx context.Context) error {
		logger.Infow("Closing PostgreSQL pool")
		pgPool.Close()
		return nil
	})

	redisClient, err := cfg.Storage.ConnectionToRedis(logger)
	if err != nil {
		logger.Fatalw("Connect to Redis", "error", err)
		return nil
	}
	c.Add(func(ctx context.Context) error {
		logger.Infow("Closing Redis client")
		return redisClient.Close()
	})

	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	
	httpServer := &http.Server{
		Addr: 		fmt.Sprintf(":%d", cfg.PublicServer.Port),
		Handler:    router,
	}

	txManager := txmanager.NewTransactor(pgPool, []*pgxpool.Pool{pgPool})

	teamStorage := postgres.NewTeamStorage(txManager, logger)
	userStorage := postgres.NewUserStorage(txManager, logger)
	prStorage := postgres.NewPRStorage(txManager, logger)
	statsCache := redis.NewStatsCache(redisClient, logger)

	prUseCase := pr_usecase.NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, txManager, logger)
	userUseCase := user_usecase.NewUserUseCase(userStorage, txManager, teamStorage, logger)
	teamUseCase := team_usecase.NewTeamUseCase(teamStorage, userStorage, txManager, logger)
	statsUseCase := stats_usecase.NewStatsUseCase(statsCache, userStorage)

	pullRequestController := controllers.NewPullRequestController(prUseCase)
	userController := controllers.NewUserController(userUseCase)
	teamController := controllers.NewTeamController(teamUseCase)
	statsController := controllers.NewStatsController(statsUseCase)

	controller := controllers.NewController(userController, teamController, statsController, pullRequestController)

	gen.RegisterHandlers(router, controller)

	return &Server{
		closer: 		c,
		router:       	router,
		pgPool: 		pgPool,
		config: 		cfg,
		httpServer:   	httpServer,
		logger: 		logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	s.closer.Add(func(ctx context.Context) error {
		s.logger.Infow("Shutting down HTTP server")
		return s.httpServer.Shutdown(ctx)
	})
	
	go func() {
		s.logger.Infow("Starting HTTP server",
			"address", s.httpServer.Addr,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalw("HTTP server error",
				"error", err)
		}
	}()
	
	return nil
}


func (s *Server) Shutdown(ctx context.Context) error {
	return s.closer.Close(ctx)
}
