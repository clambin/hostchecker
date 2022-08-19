package sitechecker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"hostchecker/config"
	"net/http"
	"time"
)

//go:generate mockery --name Checker
type Checker interface {
	Check() (*Stats, error)
}

type HTTPChecker struct {
	Target     config.HTTPTarget
	HTTPClient *http.Client
}

var _ Checker = &HTTPChecker{}

type Stats struct {
	Target         config.HTTPTarget
	Up             bool
	Latency        time.Duration
	IsTLS          bool
	CertificateAge time.Duration
}

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

	if hc.isValidHTTPCode(resp.StatusCode) {
		stats.Up = true
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

func (hc HTTPChecker) isValidHTTPCode(code int) bool {
	for _, validCode := range hc.Target.Codes {
		if code == validCode {
			return true
		}
	}
	return false
}
