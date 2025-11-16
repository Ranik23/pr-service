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


func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	Convey("GetUserByID success", t, func() {
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

		userStorage.EXPECT().
			GetUserByID(ctx, userID).
			Return(&models.User{ID: userID}, nil)

		u, err := uc.GetUserByID(ctx, userID)

		So(err, ShouldBeNil)
		So(u.ID.String(), ShouldEqual, userID.String())
	})
}

func TestUserUseCase_GetUserByID_InvalidID(t *testing.T) {
	Convey("GetUserByID invalid user ID", t, func() {
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

		u, err := uc.GetUserByID(ctx, "")

		So(err, ShouldEqual, errs.ErrInvalidUserID)
		So(u, ShouldBeNil)
	})
}	

func TestUserUseCase_GetUserByID_NotFound(t *testing.T) {
	Convey("GetUserByID not found", t, func() {
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

		userID := domain.UserID("u2")

		userStorage.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, repoerrors.ErrNotFound)

		u, err := uc.GetUserByID(ctx, userID)

		So(err, ShouldEqual, errs.ErrUserNotFound)
		So(u, ShouldBeNil)
	})
}


func TestUserUseCase_GetUserByID_Unexpected(t *testing.T) {
	Convey("GetUserByID unexpected storage error", t, func() {
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

		userStorage.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("db error"))

		u, err := uc.GetUserByID(ctx, userID)

		So(err.Error(), ShouldEqual, "db error")
		So(u, ShouldBeNil)
	})
}
