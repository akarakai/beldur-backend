package campaign

import (
	"beldur/internal/id"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c, err := New("a new campaign", "a beautiful description", id.PlayerId(1))
		require.NoError(t, err)
		require.NotNil(t, c)

		assert.Equal(t, StatusCreated, c.status)
		assert.Equal(t, "a new campaign", c.name)
		assert.Equal(t, "a beautiful description", c.description)
		assert.Equal(t, id.PlayerId(1), c.master)

		assert.NotZero(t, c.createdAt)
		assert.Nil(t, c.startedAt)
		assert.Nil(t, c.finishedAt)
		assert.NotNil(t, c.players)
		assert.Len(t, c.players, 1)
	})

	t.Run("invalid name - too long", func(t *testing.T) {
		name := strings.Repeat("a", MaxNameCharacters+1)
		c, err := New(name, "ok", id.PlayerId(1))
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrInvalidCampaignName)
	})

	t.Run("invalid description - too long", func(t *testing.T) {
		desc := strings.Repeat("d", MaxDescriptionCharacters+1)
		c, err := New("ok", desc, id.PlayerId(1))
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrInvalidCampaignDescription)
	})

	t.Run("boundary values - allowed max lengths", func(t *testing.T) {
		name := strings.Repeat("n", MaxNameCharacters)
		desc := strings.Repeat("d", MaxDescriptionCharacters)

		c, err := New(name, desc, id.PlayerId(1))
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, name, c.name)
		assert.Equal(t, desc, c.description)
	})
}

func TestCampaign_AddPlayer(t *testing.T) {
	newCampaign := func(t *testing.T) *Campaign {
		t.Helper()
		c, err := New("ok name", "ok description", id.PlayerId(1))
		require.NoError(t, err)
		return c
	}

	t.Run("success - adds player", func(t *testing.T) {
		c := newCampaign(t)

		err := c.AddPlayer(id.PlayerId(2))
		require.NoError(t, err)

		_, ok := c.players[id.PlayerId(2)]
		assert.True(t, ok)
		assert.Len(t, c.players, 2)
	})

	t.Run("failure - duplicate player", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))

		err := c.AddPlayer(id.PlayerId(2))
		assert.ErrorIs(t, err, ErrPlayerAlreadyInCampaign)
	})

	t.Run("failure - campaign finished", func(t *testing.T) {
		c := newCampaign(t)

		// start -> finish
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())
		require.NoError(t, c.Finish())

		err := c.AddPlayer(id.PlayerId(3))
		assert.ErrorIs(t, err, ErrCampaignFinished)
	})

	t.Run("failure - campaign cancelled", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.Cancel())

		err := c.AddPlayer(id.PlayerId(2))
		assert.ErrorIs(t, err, ErrCampaignCancelled)
	})

	t.Run("table - multiple distinct players", func(t *testing.T) {
		c := newCampaign(t)

		cases := []struct {
			name     string
			playerID id.PlayerId
			wantErr  error
		}{
			{"add p2", id.PlayerId(2), nil},
			{"add p3", id.PlayerId(3), nil},
			{"add p4", id.PlayerId(4), nil},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				err := c.AddPlayer(tc.playerID)
				if tc.wantErr == nil {
					require.NoError(t, err)
					_, ok := c.players[tc.playerID]
					assert.True(t, ok)
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
				}
			})
		}

		assert.Len(t, c.players, len(cases)+1)
	})
}

func TestCampaign_Start(t *testing.T) {
	newCampaign := func(t *testing.T) *Campaign {
		t.Helper()
		c, err := New("ok", "ok", id.PlayerId(1))
		require.NoError(t, err)
		return c
	}

	t.Run("failure - not enough players", func(t *testing.T) {
		c := newCampaign(t)

		err := c.Start()
		assert.ErrorIs(t, err, ErrNotEnoughPlayersToStart)
		assert.Equal(t, StatusCreated, c.status)
		assert.Nil(t, c.startedAt)
	})

	t.Run("success - with minimum players", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))

		err := c.Start()
		require.NoError(t, err)

		assert.Equal(t, StatusStarted, c.status)
		require.NotNil(t, c.startedAt)
		assert.WithinDuration(t, time.Now(), *c.startedAt, 2*time.Second)
	})

	t.Run("idempotent - already started returns nil and does not change startedAt", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())

		require.NotNil(t, c.startedAt)
		first := *c.startedAt

		// call Start again
		err := c.Start()
		require.NoError(t, err)

		require.NotNil(t, c.startedAt)
		assert.Equal(t, first, *c.startedAt)
		assert.Equal(t, StatusStarted, c.status)
	})

	t.Run("failure - cannot start if finished", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())
		require.NoError(t, c.Finish())

		err := c.Start()
		assert.ErrorIs(t, err, ErrCampaignFinished)
	})

	t.Run("failure - cannot start if cancelled", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.Cancel())

		err := c.Start()
		assert.ErrorIs(t, err, ErrCampaignCancelled)
	})

	t.Run("table - player counts", func(t *testing.T) {
		cases := []struct {
			name             string
			extraPlayerCount int // players added in addition to master
			wantErr          error
		}{
			{"0 extra players (master only) -> error", 0, ErrNotEnoughPlayersToStart},
			{"1 extra player -> ok", 1, nil},
			{"2 extra players -> ok", 2, nil},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				c := newCampaign(t)

				// master is already counted as a player
				for i := 0; i < tc.extraPlayerCount; i++ {
					require.NoError(t, c.AddPlayer(id.PlayerId(2+i)))
				}

				err := c.Start()
				if tc.wantErr == nil {
					require.NoError(t, err)
					assert.Equal(t, StatusStarted, c.status)
					assert.NotNil(t, c.startedAt)
				} else {
					assert.ErrorIs(t, err, tc.wantErr)
					assert.Equal(t, StatusCreated, c.status)
					assert.Nil(t, c.startedAt)
				}
			})
		}
	})

}

func TestCampaign_Finish(t *testing.T) {
	newCampaign := func(t *testing.T) *Campaign {
		t.Helper()
		c, err := New("ok", "ok", id.PlayerId(1))
		require.NoError(t, err)
		return c
	}

	t.Run("failure - cannot finish if not started (created)", func(t *testing.T) {
		c := newCampaign(t)

		err := c.Finish()
		assert.ErrorIs(t, err, ErrCampaignNotStarted)
		assert.Equal(t, StatusCreated, c.status)
		assert.Nil(t, c.finishedAt)
	})

	t.Run("success - finish from started", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())

		err := c.Finish()
		require.NoError(t, err)

		assert.Equal(t, StatusFinished, c.status)
		require.NotNil(t, c.finishedAt)
		assert.WithinDuration(t, time.Now(), *c.finishedAt, 2*time.Second)
	})

	t.Run("failure - finish twice", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())
		require.NoError(t, c.Finish())

		err := c.Finish()
		assert.ErrorIs(t, err, ErrCampaignFinished)
	})

	t.Run("failure - cannot finish if cancelled", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.Cancel())

		err := c.Finish()
		assert.ErrorIs(t, err, ErrCampaignCancelled)
	})
}

func TestCampaign_Cancel(t *testing.T) {
	newCampaign := func(t *testing.T) *Campaign {
		t.Helper()
		c, err := New("ok", "ok", id.PlayerId(1))
		require.NoError(t, err)
		return c
	}

	t.Run("success - cancel from created", func(t *testing.T) {
		c := newCampaign(t)

		err := c.Cancel()
		require.NoError(t, err)

		assert.Equal(t, StatusCancelled, c.status)
		require.NotNil(t, c.finishedAt)
		assert.WithinDuration(t, time.Now(), *c.finishedAt, 2*time.Second)
	})

	t.Run("failure - cancel twice", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.Cancel())

		err := c.Cancel()
		assert.ErrorIs(t, err, ErrCampaignCancelled)
	})

	t.Run("failure - cannot cancel if started", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())

		err := c.Cancel()
		assert.ErrorIs(t, err, ErrCampaignAlreadyStarted)
	})

	t.Run("failure - cannot cancel if finished", func(t *testing.T) {
		c := newCampaign(t)
		require.NoError(t, c.AddPlayer(id.PlayerId(2)))
		require.NoError(t, c.Start())
		require.NoError(t, c.Finish())

		err := c.Cancel()
		assert.ErrorIs(t, err, ErrCampaignFinished)
	})
}
