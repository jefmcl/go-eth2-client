package http_test

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/jefmcl/go-eth2-client/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	status := nethttp.StatusTeapot
	data := []byte("data")
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(data)
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := http.New(ctx, http.WithAddress(srv.URL))

	require.NotNil(t, err)
	require.Equal(t, "failed to confirm node connection: failed to fetch genesis: failed to request genesis: GET failed with status 418: data", err.Error())

	var httpError http.Error
	require.True(t, errors.As(err, &httpError))
	require.Equal(t, status, httpError.StatusCode)
	require.Equal(t, data, httpError.Data)
	require.Equal(t, nethttp.MethodGet, httpError.Method)
	require.Equal(t, "/eth/v1/beacon/genesis", httpError.Endpoint)
}

func TestClientShouldSendExtraHeadersWhenProvided(t *testing.T) {
	authorizationHeader := "Authorization"
	authorizationToken := "Bearer token"
	data := []byte("data")
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if r.Header.Get(authorizationHeader) != authorizationToken {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_, _ = w.Write(data)
			return
		}
		w.WriteHeader(nethttp.StatusTeapot)
		_, _ = w.Write(data)
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := http.New(ctx,
		http.WithAddress(srv.URL),
		http.WithExtraHeaders(map[string]string{authorizationHeader: authorizationToken}),
	)

	require.Error(t, err)
	var httpError http.Error
	require.True(t, errors.As(err, &httpError))
	require.Equal(t, nethttp.StatusTeapot, httpError.StatusCode)
}
