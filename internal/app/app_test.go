//go:build integration

package app

import (
	"beldur/pkg/auth/jwt"
	"beldur/pkg/db/postgres"
	"beldur/pkg/httperr"
	"beldur/pkg/logger"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testServerAddress = "http://localhost:5555"
	testPool          *pgxpool.Pool
	fiberApp          *FiberApp
)

func TestMain(m *testing.M) {
	logger.Init()
	ctx := context.Background()

	pg, err := tcpostgres.Run(
		ctx,
		"postgres:14-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithOrderedInitScripts("../../schema.sql"),
	)
	if err != nil {
		panic(err)
	}

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		panic(err)
	}
	cfg.MaxConns = 4

	testPool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		panic(err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		if err := testPool.Ping(ctx); err == nil {
			break
		}
		if time.Now().After(deadline) {
			panic("database not ready")
		}
		time.Sleep(200 * time.Millisecond)
	}

	buildFiberApp()

	go fiberApp.Listen("5555")

	time.Sleep(time.Second)

	code := m.Run()

	// teardown
	testPool.Close()
	_ = pg.Terminate(ctx)
	_ = fiberApp.app.Shutdown()
	os.Exit(code)
}

func endpoint(endpoint string) string {
	return testServerAddress + endpoint
}

func buildFiberApp() {
	secret := []byte("test-secret")
	expiration, _ := time.ParseDuration("168h")
	issuer := "baldur-test"
	jwtService := jwt.NewService(secret, expiration, issuer)

	transactor, querier := postgres.NewTransactor(testPool)

	fiberApp = NewTest(Deps{
		JwtService: jwtService,
		Transactor: transactor,
		QProvider:  querier,
	})
}

// DoJSON is an helper generic function that sends an http request and does the routine checking
// Success: decode response into T (works for 2xx and will still decode if body exists)
func DoJSONOK[T any](t *testing.T, client *http.Client, method, path string, reqBody any) (*http.Response, T) {
	t.Helper()

	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		require.NoError(t, err)
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, endpoint(path), body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var out T
	if len(b) > 0 {
		require.NoError(t, json.Unmarshal(b, &out))
	}

	return resp, out
}

// Failure: decode response into httperr.Response (use for 4xx/5xx cases)
func DoJSONFail(t *testing.T, client *http.Client, method, path string, reqBody any) (*http.Response, httperr.Response) {
	t.Helper()

	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		require.NoError(t, err)
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, endpoint(path), body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var out httperr.Response
	if len(b) > 0 {
		require.NoError(t, json.Unmarshal(b, &out))
	}

	return resp, out
}
