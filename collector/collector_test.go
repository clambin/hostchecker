package collector_test

import (
	"bytes"
	"fmt"
	"github.com/clambin/hostchecker/checker"
	"github.com/clambin/hostchecker/checker/mocks"
	"github.com/clambin/hostchecker/collector"
	"github.com/clambin/hostchecker/config"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCollector_Collect(t *testing.T) {
	checker1 := &mocks.Checker{}
	checker2 := &mocks.Checker{}

	c := collector.New(config.Targets{HTTP: []config.HTTPTarget{
		{Name: "tls", URL: "https://localhost"},
		{Name: "down", URL: "http://localhost"},
	}})
	require.Len(t, c.Targets, 2)
	c.Targets[0] = checker1
	c.Targets[1] = checker2

	checker1.On("Check").Return(&checker.Stats{
		Target: config.HTTPTarget{
			Name: "tls",
			URL:  "https://localhost",
		},
		Up:             true,
		Latency:        100 * time.Millisecond,
		IsTLS:          true,
		CertificateAge: 24 * time.Hour,
	}, nil)
	checker2.On("Check").Return(&checker.Stats{
		Target: config.HTTPTarget{
			Name: "down",
			URL:  "http://localhost",
		},
		Up: false,
	}, nil)

	const expected = `# HELP hostchecker_certificate_expiry Number of days before the HTTPS certificate expires
# TYPE hostchecker_certificate_expiry gauge
hostchecker_certificate_expiry{site_name="tls",site_url="https://localhost"} 1
# HELP hostchecker_site_latency_seconds Time to check the site, in seconds
# TYPE hostchecker_site_latency_seconds gauge
hostchecker_site_latency_seconds{site_name="tls",site_url="https://localhost"} 0.1
# HELP hostchecker_site_up Set to 1 if the site is up
# TYPE hostchecker_site_up gauge
hostchecker_site_up{site_name="down",site_url="http://localhost"} 0
hostchecker_site_up{site_name="tls",site_url="https://localhost"} 1
`
	assert.NoError(t, testutil.CollectAndCompare(c, bytes.NewBufferString(expected)))
}

func TestCollector_Collect_Failure(t *testing.T) {
	chk := &mocks.Checker{}

	c := collector.New(config.Targets{HTTP: []config.HTTPTarget{
		{Name: "bad", URL: "http://localhost"},
	}})
	require.Len(t, c.Targets, 1)
	c.Targets[0] = chk

	chk.On("Check").Return(&checker.Stats{
		Target: config.HTTPTarget{Name: "bad", URL: "http://localhost"},
	}, fmt.Errorf("fail"))

	err := testutil.CollectAndCompare(c, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `Desc{fqName: "hostchecker_error", help: "Error reaching site http://localhost", constLabels: {}, variableLabels: []}`)
}

type fakeChecker struct{}

func (f fakeChecker) Check() (*checker.Stats, error) {
	return &checker.Stats{
		Target: config.HTTPTarget{
			Name: "test",
			URL:  "https://localhost:8080",
		},
		Up:      true,
		Latency: 100 * time.Millisecond,
	}, nil
}

func BenchmarkCollector_Collect(b *testing.B) {
	c := collector.New(config.Targets{HTTP: []config.HTTPTarget{
		{Name: "tls", URL: "https://localhost"},
	}})
	require.Len(b, c.Targets, 1)
	c.Targets[0] = &fakeChecker{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := testutil.CollectAndCount(c)
		if count != 2 {
			b.Fatalf("got %d metrics", count)
		}
	}
}
