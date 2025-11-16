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
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_UpdateUserActivity_SuccessTrue(t *testing.T) {
	Convey("UpdateUserActivity success true", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
		mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		tx := txmock.NewMockTxManager(ctrl)
		uc := NewUserUseCase(userStorage, tx, teamStorage, mockLog) 
		ctx := context.Background()

		tx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})
		
		userID := domain.UserID("u1")

		userActivity := domain.UserActivityStatus(true)

		userStorage.EXPECT().
			GetUserByID(ctx, userID).
			Return(&models.User{ID: userID}, nil)

		userStorage.EXPECT().
			UpdateActivity(ctx, userID, userActivity).
			Return(nil)

		_, err := uc.UpdateUserActivity(ctx, userID, domain.UserStatusActive)

		So(err, ShouldBeNil)
	})
}

func TestUserUseCase_UpdateUserActivity_NotFound(t *testing.T) {
	Convey("UpdateUserActivity user not found", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
		mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		tx := txmock.NewMockTxManager(ctrl)
		uc := NewUserUseCase(userStorage, tx, teamStorage, mockLog) 
		ctx := context.Background()

		tx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		userID := domain.UserID("u1")
		userActivity := domain.UserActivityStatus(true)

		userStorage.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, repoerrors.ErrNotFound)

		_, err := uc.UpdateUserActivity(ctx, userID, userActivity)

		So(err, ShouldEqual, errs.ErrUserNotFound)
	})
}

func TestUserUseCase_UpdateUserActivity_UpdateFails(t *testing.T) {
	Convey("UpdateUserActivity update fails", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
		mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
		mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		teamStorage := mock.NewMockTeamStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		tx := txmock.NewMockTxManager(ctrl)
		uc := NewUserUseCase(userStorage, tx, teamStorage, mockLog) 
		ctx := context.Background()

		tx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		userID := domain.UserID("u1")
		userActivity := domain.UserActivityStatus(true)

		userStorage.EXPECT().
			GetUserByID(ctx, userID).
			Return(&models.User{ID: userID}, nil)

		userStorage.EXPECT().
			UpdateActivity(ctx, userID, userActivity).
			Return(errors.New("update error"))

		_, err := uc.UpdateUserActivity(ctx, userID, userActivity)

		So(err.Error(), ShouldEqual, "update error")
	})
}
