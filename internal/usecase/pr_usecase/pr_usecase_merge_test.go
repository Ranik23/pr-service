package pr_usecase

import (
	"app/internal/domain"
	cachemock "app/internal/repository/cache/mock"
	repoerrors "app/internal/repository/errs"
	"app/internal/repository/models"
	mock "app/internal/repository/storage/mock"
	"app/internal/usecase/errs"
	mocklog "app/pkg/logger/mock"
	txmock "app/pkg/txmanager/mock"
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestMergePR_NotFound(t *testing.T) {
	Convey("MergePR: PR not found", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
    	mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
    	mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("u1")

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(nil, repoerrors.ErrNotFound)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		err := uc.MergePR(context.Background(), prID)
		So(err, ShouldEqual, errs.ErrPullRequestNotFound)
	})
}

func TestMergePR_Success(t *testing.T) {
	Convey("MergePR: success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLog := mocklog.NewMockLogger(ctrl)
    	mockLog.EXPECT().Errorw(gomock.Any(), gomock.Any()).AnyTimes()
    	mockLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

		prStorage := mock.NewMockPRStorage(ctrl)
		userStorage := mock.NewMockUserStorage(ctrl)
		teamStorage := mock.NewMockTeamStorage(ctrl)
		statsCache := cachemock.NewMockStatsCache(ctrl)
		mocktx := txmock.NewMockTxManager(ctrl)

		uc := NewPRUseCase(prStorage, userStorage, statsCache, teamStorage, mocktx, mockLog)

		prID := domain.PRID("u1")

		prStorage.EXPECT().GetPullRequestByID(gomock.Any(), prID).
			Return(&models.PullRequest{ID: prID}, nil)

		prStorage.EXPECT().
			UpdatePullRequestStatus(gomock.Any(), prID, domain.PRStatusMerged).
			Return(nil)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})


		err := uc.MergePR(context.Background(), prID)
		So(err, ShouldBeNil)
	})
}