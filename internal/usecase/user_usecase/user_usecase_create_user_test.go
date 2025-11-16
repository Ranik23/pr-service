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

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	Convey("CreateUser success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
    	mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
    	mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		userStorage := mock.NewMockUserStorage(ctrl)
		tx := txmock.NewMockTxManager(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		uc := NewUserUseCase(userStorage, tx, teamStorage, mockLog)
		ctx := context.Background()

		tx.EXPECT().
			WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		userID := domain.UserID("U1")

		userStorage.EXPECT().
			CreateUser(ctx, userID, "Name").
			Return(&models.User{ID: userID}, nil)

		u, err := uc.CreateUser(ctx, userID, "Name")

		So(err, ShouldBeNil)
		So(u.ID.String(), ShouldEqual, userID.String())
	})
}


func TestUserUseCase_CreateUser_AlreadyExists(t *testing.T) {
	Convey("CreateUser already exists", t, func() {
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
			CreateUser(ctx, userID, "Name").
			Return(nil, repoerrors.ErrAlreadyExists)

		u, err := uc.CreateUser(ctx, userID, "Name")

		So(err, ShouldEqual, errs.ErrUserAlreadyExists)
		So(u, ShouldBeNil)
	})
}



func TestUserUseCase_CreateUser_Unexpected(t *testing.T) {
	Convey("CreateUser unexpected storage error", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
    	mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
    	mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		userStorage := mock.NewMockUserStorage(ctrl)
		tx := txmock.NewMockTxManager(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		uc := NewUserUseCase(userStorage, tx, teamStorage, mockLog)
		ctx := context.Background()

		tx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, a, b any, fn func(context.Context) error) error {
				return fn(ctx)
			})
		
		userID := domain.UserID("u1")

		userStorage.EXPECT().
			CreateUser(ctx, userID, "Name").
			Return(nil, errors.New("db error"))

		u, err := uc.CreateUser(ctx, userID, "Name")

		So(err.Error(), ShouldEqual, "db error")
		So(u, ShouldBeNil)
	})
}

func TestUserUseCase_CreateUser_EmptyID(t *testing.T) {
	Convey("CreateUser empty ID", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := loggermock.NewMockLogger(ctrl)
    	mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
    	mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		userStorage := mock.NewMockUserStorage(ctrl)
		tx := txmock.NewMockTxManager(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		uc := NewUserUseCase(userStorage, tx, teamStorage, mockLog)
		ctx := context.Background()

		u, err := uc.CreateUser(ctx, "", "Name")

		So(err, ShouldEqual, errs.ErrInvalidUserID)
		So(u, ShouldBeNil)
	})
}
