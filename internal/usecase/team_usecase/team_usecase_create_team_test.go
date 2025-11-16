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

func TestTeamUseCase_CreateTeam_Success(t *testing.T) {
    Convey("CreateTeam success", t, func() {
		ctrl := gomock.NewController(t)
    	defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)

    	mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
    	mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLog)
        ctx := context.Background()

        teamID := domain.TeamID(1)
        userID := domain.UserID("u1")

        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            CreateTeam(ctx, "Alpha").
            Return(&models.Team{ID: teamID, TeamName: "Alpha"}, nil)

        mockUser.EXPECT().
            CreateUser(ctx, userID, "User One").
            Return(&models.User{ID: userID}, nil).
            AnyTimes()

        mockTeam.EXPECT().
            CreateUserTeamInstance(ctx, teamID, userID).
            Return(nil)

        team, err := uc.CreateTeam(ctx, "Alpha", []domain.TeamUser{{ID: userID, Name: "User One"}})

        So(err, ShouldBeNil)
        So(team, ShouldNotBeNil)
		So(team.ID, ShouldEqual, int64(1))
		So(team.TeamName, ShouldEqual, "Alpha")
    })
}


func TestTeamUseCase_CreateTeam_AlreadyExists(t *testing.T) {
    Convey("Team already exists", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)
        mockLog := loggermock.NewMockLogger(ctrl)

        mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLog)
        ctx := context.Background()



        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(_ context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            CreateTeam(ctx, "Alpha").
            Return(nil, repoerrors.ErrAlreadyExists)

        team, err := uc.CreateTeam(ctx, "Alpha", []domain.TeamUser{{ID: "u1", Name: "User One"}})

        So(team, ShouldBeNil)
        So(errors.Is(err, errs.ErrTeamAlreadyExists), ShouldBeTrue)
    })
}

func TestTeamUseCase_CreateTeam_UserAlreadyHasTeam(t *testing.T) {
    Convey("User already has team", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)
        mockLog := loggermock.NewMockLogger(ctrl)

        mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLog)
        ctx := context.Background()

        teamID := domain.TeamID(1)
        userID := domain.UserID("1")

        mockTx.EXPECT().
            WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
            DoAndReturn(func(_ context.Context, iso, mode any, fn func(context.Context) error) error {
                return fn(ctx)
            })

        mockTeam.EXPECT().
            CreateTeam(ctx, "A").
            Return(&models.Team{ID: teamID, TeamName: "A"}, nil)

        mockUser.EXPECT().
            CreateUser(ctx, userID, "User One").
            Return(nil, repoerrors.ErrAlreadyExists)

        team, err := uc.CreateTeam(ctx, "A", []domain.TeamUser{{ID: userID, Name: "User One"}})

        So(team, ShouldBeNil)
        So(errors.Is(err, errs.ErrUserAlreadyHasTeam), ShouldBeTrue)
    })
}


func TestTeamUseCase_CreateTeam_InvalidTeamName(t *testing.T) {
    Convey("CreateTeam invalid team name", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)
        mockLog := loggermock.NewMockLogger(ctrl)

        mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLog)
        ctx := context.Background()

        team, err := uc.CreateTeam(ctx, "", []domain.TeamUser{{ID: "u1", Name: "User One"}})

        So(team, ShouldBeNil)
        So(errors.Is(err, errs.ErrInvalidTeamName), ShouldBeTrue)
    })
}  


func TestTeamUseCase_CreateTeam_NoUsersProvided(t *testing.T) {
    Convey("CreateTeam no users provided", t, func() {
        ctrl := gomock.NewController(t)
        defer ctrl.Finish()

        mockTeam := mock.NewMockTeamStorage(ctrl)
        mockUser := mock.NewMockUserStorage(ctrl)
        mockTx := txmock.NewMockTxManager(ctrl)
        mockLog := loggermock.NewMockLogger(ctrl)

        mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()

        uc := NewTeamUseCase(mockTeam, mockUser, mockTx, mockLog)
        ctx := context.Background()

        team, err := uc.CreateTeam(ctx, "Alpha", []domain.TeamUser{})

        So(team, ShouldBeNil)
        So(errors.Is(err, errs.ErrNoUsersProvided), ShouldBeTrue)
    })
}

