package team_usecase

import (
	"app/internal/domain"
	repoerrors "app/internal/repository/errs"
	"app/internal/repository/models"
	"app/internal/repository/storage/mock"
	"app/internal/usecase/errs"
	loggermock "app/pkg/logger/mock"
	txmock "app/pkg/txmanager/mock"
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)


func TestTeamUseCase_GetTeamByName_Success(t *testing.T) {
    Convey("GetTeamByName success", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockLogger := loggermock.NewMockLogger(ctrl)
        mockLogger.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
        mockLogger.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLogger)
        ctx := context.Background()

        teamID := domain.TeamID(1)
        userID1 := domain.UserID("1")
        userID2 := domain.UserID("2")

        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            GetTeamByName(ctx, "alpha").
            Return(&models.Team{ID: teamID, TeamName: "alpha"}, nil)

        mockTeam.EXPECT().
            GetUsersByTeam(ctx, teamID).
            Return([]models.User{
                {ID: userID1},
                {ID: userID2},
            }, nil)

        team, err := uc.GetTeamByName(ctx, "alpha")

        So(err, ShouldBeNil)
        So(team, ShouldNotBeNil)
        So(team.ID.Int64(), ShouldEqual, int64(1))
        So(len(team.Users), ShouldEqual, 2)
    })
}


func TestTeamUseCase_GetTeamByName_NotFound(t *testing.T) {
    Convey("GetTeamByName team not found", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockLogger := loggermock.NewMockLogger(ctrl)
        mockLogger.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLogger)
        ctx := context.Background()

        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            GetTeamByName(ctx, "beta").
            Return(nil, repoerrors.ErrNotFound)

        team, err := uc.GetTeamByName(ctx, "beta")

        So(err, ShouldNotBeNil)
        So(err, ShouldEqual, errs.ErrTeamNotFound)
        So(team, ShouldBeNil)
    })
}


func TestTeamUseCase_GetTeamByName_UnexpectedError(t *testing.T) {
    Convey("GetTeamByName unexpected storage error", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockLogger := loggermock.NewMockLogger(ctrl)
        mockLogger.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLogger)
        ctx := context.Background()

        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            GetTeamByName(ctx, "gamma").
            Return(nil, errors.New("db error"))

        team, err := uc.GetTeamByName(ctx, "gamma")

        So(err.Error(), ShouldEqual, "db error")
        So(team, ShouldBeNil)
    })
}


func TestTeamUseCase_GetTeamByName_UsersError(t *testing.T) {
    Convey("GetTeamByName users retrieval error", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockLogger := loggermock.NewMockLogger(ctrl)
        mockLogger.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLogger)
        ctx := context.Background()


        teamID := domain.TeamID(1)

        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            GetTeamByName(ctx, "delta").
            Return(&models.Team{ID: teamID, TeamName: "delta"}, nil)

        mockTeam.EXPECT().
            GetUsersByTeam(ctx, teamID).
            Return(nil, errors.New("users error"))

        team, err := uc.GetTeamByName(ctx, "delta")

        So(err.Error(), ShouldEqual, "users error")
        So(team, ShouldBeNil)
    })
}
