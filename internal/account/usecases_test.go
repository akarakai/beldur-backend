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

func strPtr(s string) *string {
	return &s
}

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}

type registerAccountCases struct {
	testName    string
	input       CreateAccountRequest
	savedAcc    *Account
	savedPlayer *player.Player
}

func TestRegisterAccount_Success(t *testing.T) {
	ctx := context.Background()

	saver := new(MockSaver)
	uniquePlayer := new(MockUniquePlayerCreator)
	transactor := new(MockTransactor)
	issuer := new(MockTokenIssuer)

	svc := NewAccountRegistration(transactor, saver, uniquePlayer, issuer)

	tests := []registerAccountCases{
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
				CreatedAt: time.Now(),
			},
			savedPlayer: &player.Player{
				Id:   1,
				Name: "username123",
			},
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
				CreatedAt: time.Now(),
			},
			savedPlayer: &player.Player{
				Id:   1,
				Name: "username123",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
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
					test.savedAcc.Email = in.Email
					test.savedAcc.Password = in.Password
					test.savedAcc.Username = in.Username
				}).
				Return(test.savedAcc, nil).
				Once()
			uniquePlayer.
				On("CreateUniquePlayer", mock.Anything, mock.AnythingOfType("*player.Player"), test.savedAcc.Id).
				Return(test.savedPlayer, nil).
				Once()
			issuer.
				On("Issue", mock.Anything, mock.MatchedBy(func(c auth.Claims) bool {
					return c.Subject == test.savedAcc.Id && c.PlayerID == test.savedPlayer.Id
				})).
				Return("token", nil).
				Once()

			resp, token, err := svc.RegisterAccount(ctx, test.input)
			assert.NoError(t, err)
			assert.Equal(t, "token", token)

			assert.Equal(t, int(test.savedAcc.Id), resp.AccountID)
			assert.Equal(t, test.input.Username, resp.AccountName)
			assert.Equal(t, test.savedAcc.CreatedAt, resp.CreatedAt)

			assert.Equal(t, int(test.savedPlayer.Id), resp.Player.PlayerID)
			assert.Equal(t, test.savedPlayer.Name, resp.Player.Name)

			transactor.AssertExpectations(t)
			saver.AssertExpectations(t)
			uniquePlayer.AssertExpectations(t)
			issuer.AssertExpectations(t)

		})
	}
}

type MockFinder struct {
	mock.Mock
}

type MockSaver struct {
	mock.Mock
}

type MockUpdater struct {
	mock.Mock
}

type MockUniquePlayerCreator struct {
	mock.Mock
}

type MockTransactor struct {
	mock.Mock
}

func (m *MockTransactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

type MockTokenIssuer struct {
	mock.Mock
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
