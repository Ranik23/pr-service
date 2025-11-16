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

func TestCreatePR_UserNotFound(t *testing.T) {
	Convey("CreatePR: user not found", t, func() {
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

		userID := domain.UserID("u1")
		authorID := domain.UserID("u1")
		prID := domain.PRID("p1")
		prName := "PR"

		userStorage.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(nil, repoerrors.ErrNotFound)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		_, err := uc.CreatePR(context.Background(), authorID, prID, prName)

		So(err, ShouldEqual, errs.ErrUserNotFound)
	})
}

func TestCreatePR_AlreadyExists(t *testing.T) {
	Convey("CreatePR: pr already exists", t, func() {
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

		userID := domain.UserID("u1")
		authorID := domain.UserID("u1")
		prID := domain.PRID("p1")
		prName := "PR"

		userStorage.EXPECT().
			GetUserByID(gomock.Any(), userID).
			Return(&models.User{ID: userID}, nil)

		prStorage.EXPECT().
			CreatePullRequest(gomock.Any(), prID, prName, authorID).
			Return(nil, repoerrors.ErrAlreadyExists)

		mocktx.EXPECT().
			WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		_, err := uc.CreatePR(context.Background(), authorID, prID, prName)
		So(err, ShouldEqual, errs.ErrPullRequestAlreadyExists)
	})
}

func TestCreatePR_NoTeam(t *testing.T) {
	Convey("CreatePR: user has no team", t, func() {
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

		userID := domain.UserID("u1")
		authorID := domain.UserID("u1")
		prID := domain.PRID("p1")
		prName := "PR"

		userStorage.EXPECT().
			GetUserByID(gomock.Any(), authorID).
			Return(&models.User{ID: userID}, nil)

		prStorage.EXPECT().
			CreatePullRequest(gomock.Any(), prID, prName, authorID).
			Return(&models.PullRequest{ID: "1"}, nil)

		teamStorage.EXPECT().
			GetTeamByUserID(gomock.Any(), authorID).
			Return(nil, repoerrors.ErrNotFound)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		_, err := uc.CreatePR(context.Background(), authorID, prID, prName)
		So(err, ShouldEqual, errs.ErrUserHasNoTeam)
	})
}

func TestCreatePR_InvalidPRName(t *testing.T) {
	Convey("CreatePR: invalid PR name", t, func() {
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

		authorID := domain.UserID("u1")
		prID := domain.PRID("p1")
		prName := ""

		_, err := uc.CreatePR(context.Background(), authorID, prID, prName)
		So(err, ShouldEqual, errs.ErrInvalidPullRequestName)
	})
}

func TestCreatePR_InvalidAuthorID(t *testing.T) {
	Convey("CreatePR: invalid author ID", t, func() {
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

		authorID := domain.UserID("")
		prID := domain.PRID("p1")
		prName := "few"

		_, err := uc.CreatePR(context.Background(), authorID, prID, prName)
		So(err, ShouldEqual, errs.ErrInvalidUserID)
	})
}

func TestCreatePR_InvalidPRID(t *testing.T) {
	Convey("CreatePR: invalid PR ID", t, func() {
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

		authorID := domain.UserID("u1")
		prID := domain.PRID("")
		prName := "wf"

		_, err := uc.CreatePR(context.Background(), authorID, prID, prName)
		So(err, ShouldEqual, errs.ErrInvalidPullRequestID)
	})
}

func TestCreatePR_Success(t *testing.T) {
	Convey("CreatePR: success", t, func() {
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

		authorID := domain.UserID("u1")
		prID := domain.PRID("p1")
		prName := "pr"
		teamID := domain.TeamID(5)

		userID3 := domain.UserID("u3")
		userID4 := domain.UserID("u4")


		userStorage.EXPECT().
			GetUserByID(gomock.Any(), authorID).
			Return(&models.User{ID: authorID}, nil)

		prStorage.EXPECT().
			CreatePullRequest(gomock.Any(), prID, prName, authorID).
			Return(&models.PullRequest{ID: prID, Name: prName, AuthorID: authorID}, nil)

		teamStorage.EXPECT().
			GetTeamByUserID(gomock.Any(), authorID).
			Return(&models.Team{ID: teamID}, nil)

		teamStorage.EXPECT().
			GetUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: authorID},
				{ID: userID3},
				{ID: userID4},
			}, nil)

		prStorage.EXPECT().
			CreatePRReviewerInstance(gomock.Any(), prID, userID3).
			AnyTimes()

		prStorage.EXPECT().
			CreatePRReviewerInstance(gomock.Any(), prID, userID4).
			AnyTimes()


		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		pr, err := uc.CreatePR(context.Background(), authorID, prID, prName)

		So(err, ShouldBeNil)
		So(pr.ID.String(), ShouldEqual, prID.String())
	})
}


func TestCreatePR_SuccessEmptyTeam(t *testing.T) {
	Convey("CreatePR: success", t, func() {
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

		authorID := domain.UserID("u1")
		prID := domain.PRID("p1")
		prName := "pr"
		teamID := domain.TeamID(5)


		userStorage.EXPECT().
			GetUserByID(gomock.Any(), authorID).
			Return(&models.User{ID: authorID}, nil)

		prStorage.EXPECT().
			CreatePullRequest(gomock.Any(), prID, prName, authorID).
			Return(&models.PullRequest{ID: prID, Name: prName, AuthorID: authorID}, nil)

		teamStorage.EXPECT().
			GetTeamByUserID(gomock.Any(), authorID).
			Return(&models.Team{ID: teamID}, nil)

		teamStorage.EXPECT().
			GetUsersByTeam(gomock.Any(), teamID).
			Return([]models.User{
				{ID: authorID},
			}, nil)

		mocktx.EXPECT().WithTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, iso, mode any, fn func(context.Context) error) error {
				return fn(ctx)
			})

		pr, err := uc.CreatePR(context.Background(), authorID, prID, prName)

		So(err, ShouldBeNil)
		So(pr.ID.String(), ShouldEqual, prID.String())
	})
}

