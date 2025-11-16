package integration_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"app/internal/app"
	"app/internal/config"
	"app/internal/repository/cache"
	"app/internal/repository/cache/mock"
	"app/internal/repository/storage"
	"app/internal/repository/storage/postgres"
	"app/internal/usecase"
	"app/internal/usecase/pr_usecase"
	"app/internal/usecase/stats_usecase"
	"app/internal/usecase/team_usecase"
	"app/internal/usecase/user_usecase"
	"app/pkg/txmanager"
	"app/tests/integration/testutil"
)

type TestSuite struct {
	suite.Suite

	pool 		*pgxpool.Pool

	userUseCase 	user_usecase.UserUseCase
	teamUseCase 	team_usecase.TeamUseCase
	prUseCase   	pr_usecase.PullRequestUseCase
	statsUseCase 	stats_usecase.StatsUseCase
	usecase   		usecase.UseCase

	userStorage 	storage.UserStorage
	teamStorage 	storage.TeamStorage
	prStorage   	storage.PRStorage
	statsCache   	cache.StatsCache

	psqlContainer 	*testutil.PostgreSQLContainer
}

func (s *TestSuite) SetupSuite() {

	logger, err := app.SetupLogger(config.LoggingConfig{
		Level:             "info",
		Mode:              "dev",
		Encoding:          "console",
		DisableCaller:     true,
		DisableStacktrace: true,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		TimestampKey: 		"timestamp",
		CapitalizeLevel:    true,
	})
	s.Require().NoError(err)

	cfg, err := config.LoadConfig("../../configs/", "../../.env")
	s.Require().NoError(err)

	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Duration(cfg.PublicServer.ShutdownTimeout)*time.Second)
	defer ctxCancel()

	psqlContainer, err := testutil.NewPostgreSQLContainer(ctx)
	s.Require().NoError(err)

	s.psqlContainer = psqlContainer

	err = testutil.RunMigrations(psqlContainer.GetDSN(), "../../migrations")
	s.Require().NoError(err)

	poolConfig, err := pgxpool.ParseConfig(psqlContainer.GetDSN())
	s.Require().NoError(err)

	poolConfig.MaxConns = int32(cfg.Storage.Postgres.Pool.MaxConnections)
	poolConfig.MinConns = int32(cfg.Storage.Postgres.Pool.MinConnections)
	poolConfig.MaxConnLifetime = time.Duration(cfg.Storage.Postgres.Pool.MaxLifeTime) * time.Second
	poolConfig.MaxConnIdleTime = time.Duration(cfg.Storage.Postgres.Pool.MaxIdleTime) * time.Second
	poolConfig.HealthCheckPeriod = time.Duration(cfg.Storage.Postgres.Pool.HealthCheckPeriod) * time.Second

	pgPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	s.Require().NoError(err)

	s.pool = pgPool

	ctrl := gomock.NewController(s.T())

	txManager 	:= txmanager.NewTransactor(pgPool, []*pgxpool.Pool{pgPool})
	teamStorage := postgres.NewTeamStorage(txManager, logger)
	userStorage := postgres.NewUserStorage(txManager, logger)
	prStorage 	:= postgres.NewPRStorage(txManager, logger)

	statsCache := mock.NewMockStatsCache(ctrl)
	statsCache.EXPECT().IncrementAssignCountByUserID(gomock.Any(), gomock.Any()).AnyTimes()
	statsCache.EXPECT().DecrementAssignCountByUserID(gomock.Any(), gomock.Any()).AnyTimes()

	prUseCase 	 := pr_usecase.NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, txManager, logger)
	userUseCase  := user_usecase.NewUserUseCase(userStorage, txManager, teamStorage, logger)
	teamUseCase  := team_usecase.NewTeamUseCase(teamStorage, userStorage, txManager, logger)
	statsUseCase := stats_usecase.NewStatsUseCase(statsCache, userStorage)

	usecase := usecase.NewUseCase(userUseCase, teamUseCase, prUseCase, statsUseCase)

	s.userUseCase = userUseCase
	s.teamUseCase = teamUseCase
	s.prUseCase = prUseCase
	s.usecase = usecase
	s.userStorage = userStorage
	s.teamStorage = teamStorage
	s.prStorage = prStorage
	s.statsCache = statsCache
	s.statsUseCase = statsUseCase
}

func (s *TestSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)
	defer func() {
        if err = db.Close(); err != nil {
            s.T().Fatalf("failed to close db: %v", err)
        }   
    }()

	_, err = db.Exec(`
        TRUNCATE TABLE users, teams, user_teams, pull_requests, pr_reviewers RESTART IDENTITY CASCADE;
    `)
	s.Require().NoError(err)
}

func (s *TestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.psqlContainer.Terminate(ctx))

}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
