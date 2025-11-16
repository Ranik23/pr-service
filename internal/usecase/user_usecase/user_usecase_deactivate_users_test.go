package user_usecase

import (
	"app/internal/domain"
	repoerrors "app/internal/repository/errs"
	"app/internal/repository/models"
	"app/internal/repository/storage/mock"
	"app/internal/usecase/errs"
	loggermock "app/pkg/logger/mock"
	txmock "app/pkg/txmanager/mock"
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestDeactivateUsers_TeamNotFound(t *testing.T) {
	Convey("DeactivateUsers team not found", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
		mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		teamStorage := mock.NewMockTeamStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		txmock := txmock.NewMockTxManager(ctrl)
		uc := NewUserUseCase(userStorage, txmock, teamStorage, mockLog)
		ctx := context.Background()


		teamName := "exampleTeam"

		txmock.EXPECT().
			WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		teamStorage.EXPECT().
			GetTeamByName(ctx, teamName).
			Return(nil, repoerrors.ErrNotFound)

		err := uc.DeactivateUsersByTeamName(ctx, teamName)

		So(err, ShouldEqual, errs.ErrTeamNotFound)
	})
}

func TestDeactivateUsers_NoUsersInTeam(t *testing.T) {
	Convey("DeactivateUsers no users in team", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
		mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		teamStorage := mock.NewMockTeamStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		txmock := txmock.NewMockTxManager(ctrl)
		uc := NewUserUseCase(userStorage, txmock, teamStorage, mockLog)
		ctx := context.Background()

		teamID := domain.TeamID(1)
		teamName := "name"

		txmock.EXPECT().
			WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		teamStorage.EXPECT().
			GetTeamByName(ctx, teamName).
			Return(&models.Team{ID: teamID}, nil)

		teamStorage.EXPECT().
			GetUsersByTeam(ctx, teamID).
			Return([]models.User{}, nil)

		err := uc.DeactivateUsersByTeamName(ctx, teamName)
		So(err, ShouldEqual, errs.ErrNoUsersInTeam)
	})
}