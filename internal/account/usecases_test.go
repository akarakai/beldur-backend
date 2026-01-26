package account

import (
	"beldur/internal/id"
	"beldur/internal/player"
	"beldur/pkg/auth"
	"beldur/pkg/logger"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func strPtr(s string) *string { return &s }

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}

type registerAccountCase struct {
	testName string
	input    CreateAccountRequest

	savedAcc    *Account
	savedPlayer *player.Player
	token       string

	saveErr   error
	playerErr error
	issueErr  error

	wantErr error
}

func TestRegisterAccount_Success(t *testing.T) {
	ctx := context.Background()
	fixedNow := time.Date(2026, 1, 25, 12, 0, 0, 0, time.UTC)

	tests := []registerAccountCase{
		{
			testName: "successfully register account with email",
			input: CreateAccountRequest{
				Username: "username123",
				Password: "password123",
				Email:    strPtr("beautifulEmail@gmail.com"),
			},
			savedAcc: &Account{
				Id:        1,
				Username:  "username123",
				Password:  "megahash",
				CreatedAt: fixedNow,
			},
			savedPlayer: &player.Player{
				Id:   1,
				Name: "username123",
			},
			token: "token",
		},
		{
			testName: "successfully register account without email",
			input: CreateAccountRequest{
				Username: "username123",
				Password: "password123",
				Email:    nil,
			},
			savedAcc: &Account{
				Id:        1,
				Username:  "username123",
				Password:  "megahash",
				CreatedAt: fixedNow,
			},
			savedPlayer: &player.Player{
				Id:   1,
				Name: "username123",
			},
			token: "token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			saver := new(MockSaver)
			uniquePlayer := new(MockUniquePlayerCreator)
			transactor := new(MockTransactor)
			issuer := new(MockTokenIssuer)

			svc := NewAccountRegistration(transactor, saver, uniquePlayer, issuer)

			transactor.
				On("WithTransaction", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					txCtx := args.Get(0).(context.Context)
					fn := args.Get(1).(func(context.Context) error)
					err := fn(txCtx)
					assert.NoError(t, err)
				}).
				Return(nil).
				Once()

			saver.
				On("Save", mock.Anything, mock.AnythingOfType("*account.Account")).
				Run(func(args mock.Arguments) {
					in := args.Get(1).(*Account)

					// reflect fields set by RegisterAccount
					tc.savedAcc.Email = in.Email
					tc.savedAcc.Password = in.Password
					tc.savedAcc.Username = in.Username
				}).
				Return(tc.savedAcc, tc.saveErr).
				Once()

			uniquePlayer.
				On("CreateUniquePlayer", mock.Anything, mock.AnythingOfType("*player.Player"), tc.savedAcc.Id).
				Return(tc.savedPlayer, tc.playerErr).
				Once()

			issuer.
				On("Issue", mock.Anything, mock.MatchedBy(func(c auth.Claims) bool {
					return c.Subject == tc.savedAcc.Id && c.PlayerID == tc.savedPlayer.Id
				})).
				Return(tc.token, tc.issueErr).
				Once()

			resp, token, err := svc.RegisterAccount(ctx, tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.token, token)

			assert.Equal(t, int(tc.savedAcc.Id), resp.AccountID)
			assert.Equal(t, tc.input.Username, resp.AccountName)
			assert.Equal(t, tc.savedAcc.CreatedAt, resp.CreatedAt)

			assert.Equal(t, int(tc.savedPlayer.Id), resp.Player.PlayerID)
			assert.Equal(t, tc.savedPlayer.Name, resp.Player.Name)

			transactor.AssertExpectations(t)
			saver.AssertExpectations(t)
			uniquePlayer.AssertExpectations(t)
			issuer.AssertExpectations(t)
		})
	}
}

func TestRegisterAccount_Failure(t *testing.T) {
	ctx := context.Background()

	tests := []registerAccountCase{
		{
			testName: "malformed email",
			input: CreateAccountRequest{
				Username: "username123",
				Password: "password123",
				Email:    strPtr("beautifulEmailgmail.com"),
			},
			wantErr: ErrInvalidEmailFormat,
		},
		{
			testName: "too short username",
			input: CreateAccountRequest{
				Username: "usr",
				Password: "password123",
				Email:    nil,
			},
			wantErr: ErrInvalidUsername,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			saver := new(MockSaver)
			uniquePlayer := new(MockUniquePlayerCreator)
			issuer := new(MockTokenIssuer)
			transactor := new(MockTransactor) // Use mock instead of FnTransactor

			svc := NewAccountRegistration(transactor, saver, uniquePlayer, issuer)

			_, token, err := svc.RegisterAccount(ctx, tc.input)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Empty(t, token)

			// Verify transaction was never started for validation errors
			transactor.AssertNotCalled(t, "WithTransaction", mock.Anything, mock.Anything)
			saver.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
			uniquePlayer.AssertNotCalled(t, "CreateUniquePlayer", mock.Anything, mock.Anything, mock.Anything)
			issuer.AssertNotCalled(t, "Issue", mock.Anything, mock.Anything)
		})
	}
}

type loginUseCase struct {
	testName string
	input    UsernamePasswordLoginRequest

	accFinderReturn    *Account
	playerFinderReturn *player.Player

	returnIssuer    string
	returnErrIssuer error

	tokenResponse string
	errResponse   error
}

func TestLogin_Success(t *testing.T) {
	hash, _ := HashPassword("password123")

	tests := []loginUseCase{
		{
			testName: "success default",
			input: UsernamePasswordLoginRequest{
				Username: "username123",
				Password: "password123",
			},
			accFinderReturn: &Account{
				Id:       1,
				Username: "username123",
				Password: hash,
			},
			playerFinderReturn: &player.Player{
				Id:   1,
				Name: "username123",
			},
			returnIssuer:    "jwt-token",
			returnErrIssuer: nil,
			tokenResponse:   "jwt-token",
			errResponse:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			accFinder := new(MockFinder)
			accUpdater := new(MockUpdater)
			playerFinder := new(MockPlayerFinder)
			tokenIssuer := new(MockTokenIssuer)

			svc := NewUsernamePasswordLogin(accFinder, accUpdater, playerFinder, tokenIssuer)

			// accFinder expectation
			accFinder.
				On("FindByUsername", mock.Anything, tc.input.Username).
				Return(tc.accFinderReturn, nil).
				Once()

			// playerFinder expectation (only meaningful when acc != nil and password matches)
			playerFinder.
				On("FindByAccountId", mock.Anything, tc.accFinderReturn.Id).
				Return(tc.playerFinderReturn, nil).
				Once()

			// last access update expectation
			accUpdater.
				On("UpdateLastAccess", mock.Anything, tc.accFinderReturn.Id).
				Return(nil).
				Once()

			// token issuer expectation with claim verification
			tokenIssuer.
				On("Issue", mock.Anything, mock.MatchedBy(func(c auth.Claims) bool {
					return c.Subject == tc.accFinderReturn.Id && c.PlayerID == tc.playerFinderReturn.Id
				})).
				Return(tc.returnIssuer, tc.returnErrIssuer).
				Once()

			// act
			tok, err := svc.Login(context.Background(), tc.input)

			// assert
			if tc.errResponse != nil {
				assert.ErrorIs(t, err, tc.errResponse)
				assert.Empty(t, tok)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.tokenResponse, tok)
			}

			// verify all expectations were met
			accFinder.AssertExpectations(t)
			playerFinder.AssertExpectations(t)
			accUpdater.AssertExpectations(t)
			tokenIssuer.AssertExpectations(t)
		})
	}
}

// ---- mocks ----

// FnTransactor is a simple transactor stub that executes the fn.
// Use this in tests where RegisterAccount may start a transaction even on validation failures.
type FnTransactor struct{}

func (FnTransactor) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type MockPlayerFinder struct {
	mock.Mock
}

func (m *MockPlayerFinder) FindByUsername(ctx context.Context, username string) (*player.Player, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerFinder) FindById(ctx context.Context, playerId id.PlayerId) (*player.Player, error) {
	args := m.Called(ctx, playerId)
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *MockPlayerFinder) FindByAccountId(ctx context.Context, accountId id.AccountId) (*player.Player, error) {
	args := m.Called(ctx, accountId)
	return args.Get(0).(*player.Player), args.Error(1)
}

type MockFinder struct{ mock.Mock }
type MockSaver struct{ mock.Mock }
type MockUpdater struct{ mock.Mock }
type MockUniquePlayerCreator struct{ mock.Mock }
type MockTransactor struct{ mock.Mock }
type MockTokenIssuer struct{ mock.Mock }

func (m *MockTransactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockTokenIssuer) Issue(ctx context.Context, claims auth.Claims) (string, error) {
	args := m.Called(ctx, claims)
	return args.String(0), args.Error(1)
}

func (m *MockUniquePlayerCreator) CreateUniquePlayer(ctx context.Context, pl *player.Player, accId id.AccountId) (*player.Player, error) {
	args := m.Called(ctx, pl, accId)
	var playe *player.Player
	if p := args.Get(0); p != nil {
		playe = p.(*player.Player)
	}
	return playe, args.Error(1)
}

func (m *MockUpdater) UpdateLastAccess(ctx context.Context, accountId id.AccountId) error {
	args := m.Called(ctx, accountId)
	return args.Error(0)
}

func (m *MockUpdater) Update(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockSaver) Save(ctx context.Context, account *Account) (*Account, error) {
	args := m.Called(ctx, account)
	var acc *Account
	if v := args.Get(0); v != nil {
		acc = v.(*Account)
	}
	return acc, args.Error(1)
}

func (m *MockFinder) FindByUsername(ctx context.Context, username string) (*Account, error) {
	args := m.Called(ctx, username)
	var acc *Account
	if v := args.Get(0); v != nil {
		acc = v.(*Account)
	}
	return acc, args.Error(1)
}

func (m *MockFinder) FindById(ctx context.Context, accountId id.AccountId) (*Account, error) {
	args := m.Called(ctx, accountId)
	var acc *Account
	if v := args.Get(0); v != nil {
		acc = v.(*Account)
	}
	return acc, args.Error(1)
}
