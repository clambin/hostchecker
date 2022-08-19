package collector

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/semaphore"
	"hostchecker/config"
	"hostchecker/sitechecker"
	"net/http"
)

type Collector struct {
	Targets             []sitechecker.Checker
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

const MaxConcurrentChecks = 3

func New(targets config.Targets) *Collector {
	c := &Collector{MaxConcurrentChecks: MaxConcurrentChecks}

	for _, target := range targets.Http {
		c.Targets = append(c.Targets, &sitechecker.HTTPChecker{
			Target:     target,
			HTTPClient: http.DefaultClient,
		})
	}
	return c
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- metricUp
	ch <- metricLatency
	ch <- metricCertAge
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	maxJobs := semaphore.NewWeighted(c.MaxConcurrentChecks)
	ctx := context.Background()

	for _, site := range c.Targets {
		_ = maxJobs.Acquire(ctx, 1)

		go func(site sitechecker.Checker) {
			if stats, err := site.Check(); err == nil {
				c.collectSiteStats(ch, stats.Target, stats)
			} else {
				ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("hostchecker_error",
					"Error reaching site "+stats.Target.URL, nil, nil),
					err)
			}
			maxJobs.Release(1)
		}(site)
	}

	_ = maxJobs.Acquire(ctx, c.MaxConcurrentChecks)
}

func (c Collector) collectSiteStats(ch chan<- prometheus.Metric, site config.HTTPTarget, stats *sitechecker.Stats) {
	if !stats.Up {
		ch <- prometheus.MustNewConstMetric(metricUp, prometheus.GaugeValue, 0, site.URL, site.Name)
		return
	}

	ch <- prometheus.MustNewConstMetric(metricUp, prometheus.GaugeValue, 1.0, site.URL, site.Name)
	ch <- prometheus.MustNewConstMetric(metricLatency, prometheus.GaugeValue, stats.Latency.Seconds(), site.URL, site.Name)
	if stats.IsTLS {
		ch <- prometheus.MustNewConstMetric(metricCertAge, prometheus.GaugeValue, stats.CertificateAge.Hours()/24, site.URL, site.Name)

	}
}
