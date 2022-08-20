package sitechecker_test

import (
	"github.com/clambin/hostchecker/config"
	"github.com/clambin/hostchecker/sitechecker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSiteChecker_Check_HTTP(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello world"))
	}))

	checker := sitechecker.HTTPChecker{
		Target: config.HTTPTarget{
			Name:  "test",
			URL:   testServer.URL,
			Codes: []int{http.StatusOK},
		},
		HTTPClient: http.DefaultClient,
	}

	stats, err := checker.Check()
	require.NoError(t, err)
	assert.True(t, stats.Up)
	assert.NotZero(t, stats.Latency)

	testServer.Close()
	stats, err = checker.Check()
	require.NoError(t, err)
	assert.False(t, stats.Up)
}

func TestSiteChecker_Check_BadStatusCode(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello world"))
	}))

	checker := sitechecker.HTTPChecker{
		Target: config.HTTPTarget{
			Name:  "test",
			URL:   testServer.URL,
			Codes: []int{http.StatusAccepted},
		},
		HTTPClient: http.DefaultClient,
	}

	stats, err := checker.Check()
	require.NoError(t, err)
	assert.False(t, stats.Up)
}

func TestSiteChecker_Check_HTTPS(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("Hello world"))
	}))

	checker := sitechecker.HTTPChecker{
		Target: config.HTTPTarget{
			Name:  "test",
			URL:   testServer.URL,
			Codes: []int{http.StatusOK},
		},
		HTTPClient: testServer.Client(),
	}

	stats, err := checker.Check()
	require.NoError(t, err)
	assert.True(t, stats.Up)
	assert.NotZero(t, stats.Latency)
	assert.True(t, stats.IsTLS)
	assert.NotZero(t, stats.CertificateAge)

	testServer.Close()
	stats, err = checker.Check()
	require.NoError(t, err)
	assert.False(t, stats.Up)

}
