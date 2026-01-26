package campaign

import (
	"beldur/internal/id"
	"beldur/pkg/logger"
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}

type harness struct {
	saver      *mockSaver
	finder     *mockFinder
	updater    *mockUpdater
	transactor *mockTransactor
	svc        *UseCase
}

func newHarness() *harness {
	h := &harness{
		saver:      new(mockSaver),
		finder:     new(mockFinder),
		updater:    new(mockUpdater),
		transactor: new(mockTransactor),
	}
	h.svc = NewUseCase(h.saver, h.finder, h.updater, h.transactor)
	return h
}

func TestCreateCampaign_Success(t *testing.T) {
	h := newHarness()

	req := CreationRequest{
		Name:        "a test campaign",
		Description: "a test description",
	}
	masterId := id.PlayerId(10)

	h.saver.
		On("Save", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*Campaign)
			c.id = id.CampaignId(rand.Int())
		}).
		Return(nil)

	resp, err := h.svc.CreateNewCampaign(context.Background(), req, masterId)

	assert.NoError(t, err)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.Description, resp.Description)
	assert.Equal(t, string(StatusCreated), resp.Status)
	assert.NotEmpty(t, resp.AccessCode)

	h.saver.AssertExpectations(t)
}

func TestJoinCampaign_Success(t *testing.T) {
	h := newHarness()

	authCode := "ABCDEF"
	req := JoinRequest{Code: authCode}
	campaignId := id.CampaignId(10)
	playerId := id.PlayerId(10)

	returnCampaign := &Campaign{
		id:          campaignId,
		name:        "test name",
		description: "test description",
		status:      StatusCreated,
		createdAt:   time.Now(),
		master:      id.PlayerId(1),
		players:     map[id.PlayerId]struct{}{id.PlayerId(1): {}, id.PlayerId(2): {}, id.PlayerId(3): {}},
	}

	h.finder.
		On("FindById", mock.Anything, campaignId).
		Return(returnCampaign, nil)

	h.finder.
		On("FindAuthCode", mock.Anything, campaignId).
		Return(req.Code, nil)

	h.updater.
		On("Update", mock.Anything, mock.Anything).
		Return(nil)

	resp, err := h.svc.JoinCampaign(context.Background(), req, campaignId, playerId)

	assert.NoError(t, err)
	assert.Equal(t, returnCampaign.name, resp.Name)
	assert.Equal(t, returnCampaign.description, resp.Description)
	assert.Equal(t, string(returnCampaign.status), resp.Status)

	h.finder.AssertExpectations(t)
	h.updater.AssertExpectations(t)
}

func TestJoinCampaign_Failure(t *testing.T) {
	type tc struct {
		name         string
		reqCode      string
		storedCode   string
		campaignId   id.CampaignId
		status       StatusCampaign
		playerId     id.PlayerId
		players      map[id.PlayerId]struct{}
		expectErrIs  error
		expectUpdate bool
	}

	tests := []tc{
		{
			name:         "wrong access code",
			reqCode:      "ABCDEF",
			storedCode:   "AAAAAA",
			campaignId:   id.CampaignId(10),
			status:       StatusCreated,
			playerId:     id.PlayerId(10),
			players:      map[id.PlayerId]struct{}{id.PlayerId(1): {}, id.PlayerId(2): {}, id.PlayerId(3): {}},
			expectErrIs:  ErrWrongAccessCode,
			expectUpdate: false,
		},
		{
			name:         "player has already joined",
			reqCode:      "ABCDEF",
			storedCode:   "ABCDEF",
			campaignId:   id.CampaignId(10),
			status:       StatusCreated,
			playerId:     id.PlayerId(2),
			players:      map[id.PlayerId]struct{}{id.PlayerId(1): {}, id.PlayerId(2): {}, id.PlayerId(3): {}},
			expectErrIs:  ErrPlayerAlreadyInCampaign,
			expectUpdate: false,
		},
		{
			name:         "campaign is already started",
			reqCode:      "ABCDEF",
			storedCode:   "ABCDEF",
			campaignId:   id.CampaignId(10),
			status:       StatusStarted,
			playerId:     id.PlayerId(29),
			players:      map[id.PlayerId]struct{}{id.PlayerId(1): {}, id.PlayerId(2): {}, id.PlayerId(3): {}},
			expectErrIs:  ErrCampaignAlreadyStarted,
			expectUpdate: false,
		},
		{
			name:         "campaign is finished",
			reqCode:      "ABCDEF",
			storedCode:   "ABCDEF",
			campaignId:   id.CampaignId(10),
			status:       StatusFinished,
			playerId:     id.PlayerId(29),
			players:      map[id.PlayerId]struct{}{id.PlayerId(1): {}, id.PlayerId(2): {}, id.PlayerId(3): {}},
			expectErrIs:  ErrCampaignFinished,
			expectUpdate: false,
		},
		{
			name:         "campaign is cancelled",
			reqCode:      "ABCDEF",
			storedCode:   "ABCDEF",
			campaignId:   id.CampaignId(10),
			status:       StatusCancelled,
			playerId:     id.PlayerId(29),
			players:      map[id.PlayerId]struct{}{id.PlayerId(1): {}, id.PlayerId(2): {}, id.PlayerId(3): {}},
			expectErrIs:  ErrCampaignCancelled,
			expectUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness()

			req := JoinRequest{Code: tt.reqCode}

			returnCampaign := &Campaign{
				id:          tt.campaignId,
				name:        "test name",
				description: "test description",
				status:      tt.status,
				createdAt:   time.Now(),
				master:      id.PlayerId(1),
				players:     tt.players,
			}

			h.finder.
				On("FindById", mock.Anything, tt.campaignId).
				Return(returnCampaign, nil)

			h.finder.
				On("FindAuthCode", mock.Anything, tt.campaignId).
				Return(tt.storedCode, nil)

			if tt.expectUpdate {
				h.updater.
					On("Update", mock.Anything, mock.Anything).
					Return(nil)
			} else {
				h.updater.
					On("Update", mock.Anything, mock.Anything).
					Maybe().
					Return(nil)
			}

			_, err := h.svc.JoinCampaign(context.Background(), req, tt.campaignId, tt.playerId)

			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.expectErrIs)

			h.finder.AssertExpectations(t)
			h.updater.AssertExpectations(t)
		})
	}
}

type mockSaver struct {
	mock.Mock
}

func (m *mockSaver) Save(ctx context.Context, campaign *Campaign, accessCode string) error {
	args := m.Called(ctx, campaign, accessCode)
	return args.Error(0)
}

type mockFinder struct {
	mock.Mock
}

func (m *mockFinder) FindById(ctx context.Context, campaignId id.CampaignId) (*Campaign, error) {
	args := m.Called(ctx, campaignId)
	// If you ever want to simulate "not found", return (nil, someErr) in the expectation.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Campaign), args.Error(1)
}

func (m *mockFinder) FindAuthCode(ctx context.Context, campaignId id.CampaignId) (string, error) {
	args := m.Called(ctx, campaignId)
	return args.String(0), args.Error(1)
}

func (m *mockFinder) FindAll(ctx context.Context) ([]*Campaign, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Campaign), args.Error(1)
}

type mockUpdater struct {
	mock.Mock
}

func (m *mockUpdater) Update(ctx context.Context, campaign *Campaign) error {
	args := m.Called(ctx, campaign)
	return args.Error(0)
}

type mockTransactor struct {
	mock.Mock
}

func (m *mockTransactor) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
