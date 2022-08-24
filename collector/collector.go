package collector

import (
	"context"
	"github.com/clambin/hostchecker/checker"
	"github.com/clambin/hostchecker/config"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/semaphore"
	"net/http"
)

// Collector checks the status of a list of Targets and exposes the results as Prometheus scrape metrics
type Collector struct {
	// Targets to check
	Targets []checker.Checker
	// MaxConcurrentChecks limits the number of checks to perform concurrently. Defaults to MaxConcurrentChecks
	MaxConcurrentChecks int64
}

var _ prometheus.Collector = &Collector{}

var (
	metricUp = prometheus.NewDesc(
		prometheus.BuildFQName("hostchecker", "site", "up"),
		"Set to 1 if the site is up",
		[]string{"site_url", "site_name"},
		nil,
	)
	metricLatency = prometheus.NewDesc(
		prometheus.BuildFQName("hostchecker", "site", "latency_seconds"),
		"Time to check the site, in seconds",
		[]string{"site_url", "site_name"},
		nil,
	)
	metricCertAge = prometheus.NewDesc(
		prometheus.BuildFQName("hostchecker", "certificate", "expiry"),
		"Number of days before the HTTPS certificate expires",
		[]string{"site_url", "site_name"},
		nil,
	)
)

// MaxConcurrentChecks specifies the maximum number of sites that can be checked concurrently
const MaxConcurrentChecks = 3

// New creates a new Collectors for the provided targets
func New(targets config.Targets) *Collector {
	c := &Collector{MaxConcurrentChecks: MaxConcurrentChecks}

	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for _, target := range targets.HTTP {
		c.Targets = append(c.Targets, &checker.HTTPChecker{
			Target:     target,
			HTTPClient: httpClient,
		})
	}
	return c
}

// Describe implements the Prometheus Collector interface
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricUp
	ch <- metricLatency
	ch <- metricCertAge
}

// Collect implements the Prometheus Collector interface. It checks each target and exposes the results as Prometheus
// metrics.  Collect limits the number of checks that are performed concurrently, as specified by ConcurrentChecks.
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	maxJobs := semaphore.NewWeighted(c.MaxConcurrentChecks)
	ctx := context.Background()

	for _, site := range c.Targets {
		_ = maxJobs.Acquire(ctx, 1)

		go func(site checker.Checker) {
			c.collectTarget(ch, site)
			maxJobs.Release(1)
		}(site)
	}

	_ = maxJobs.Acquire(ctx, c.MaxConcurrentChecks)
}

func (c Collector) collectTarget(ch chan<- prometheus.Metric, target checker.Checker) {
	stats, err := target.Check()

	if err != nil {
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("hostchecker_error",
			"Error reaching site "+stats.Target.URL, nil, nil),
			err)
		return
	}

	if !stats.Up {
		ch <- prometheus.MustNewConstMetric(metricUp, prometheus.GaugeValue, 0, stats.Target.URL, stats.Target.Name)
		return
	}

	ch <- prometheus.MustNewConstMetric(metricUp, prometheus.GaugeValue, 1.0, stats.Target.URL, stats.Target.Name)
	ch <- prometheus.MustNewConstMetric(metricLatency, prometheus.GaugeValue, stats.Latency.Seconds(), stats.Target.URL, stats.Target.Name)

	if stats.IsTLS {
		ch <- prometheus.MustNewConstMetric(metricCertAge, prometheus.GaugeValue, stats.CertificateAge.Hours()/24, stats.Target.URL, stats.Target.Name)

	}
}
