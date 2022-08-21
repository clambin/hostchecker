package checker

import (
	"context"
	"github.com/clambin/hostchecker/config"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// A Checker determines the state of a target host
//
//go:generate mockery --name Checker
type Checker interface {
	Check() (*Stats, error)
}

// Stats contains all metrics determines while checking a target host
type Stats struct {
	// Target host that was checked
	Target config.HTTPTarget
	// Up is true if the host was reachable
	Up bool
	// Latency is the time to receive a response from the target host
	Latency time.Duration
	// IsTLS is true if the host was contacted over HTTPS
	IsTLS bool
	// CertificateAge is the duration until the HTTP server certificate expires
	CertificateAge time.Duration
}

// An HTTPChecker determines the state of a target HTTP(s) host
type HTTPChecker struct {
	Target     config.HTTPTarget
	HTTPClient *http.Client
}

var _ Checker = &HTTPChecker{}

// Check determines the state of an HTTP(s) target
func (hc HTTPChecker) Check() (*Stats, error) {
	stats := &Stats{Target: hc.Target}

	log.WithField("url", hc.Target.URL).Debug("checking site")
	req, err := http.NewRequestWithContext(context.Background(), hc.Target.Method, hc.Target.URL, nil)
	if err != nil {
		return stats, err
	}

	var resp *http.Response
	start := time.Now()
	if resp, err = hc.HTTPClient.Do(req); err != nil {
		log.WithError(err).WithField("url", hc.Target.URL).Debug("target not reachable")
		return stats, nil
	}

	_ = resp.Body.Close()

	log.WithFields(log.Fields{"url": hc.Target.URL, "code": resp.StatusCode}).Info("host checked")

	if stats.Up = hc.Target.Codes.IsValidCode(resp.StatusCode); stats.Up {
		stats.Latency = time.Since(start)
	} else {
		log.Warningf("target responded with unexpected HTTP code: %d", resp.StatusCode)
	}

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		stats.IsTLS = true
		stats.CertificateAge = time.Until(resp.TLS.PeerCertificates[0].NotAfter)
	}

	return stats, nil
}
