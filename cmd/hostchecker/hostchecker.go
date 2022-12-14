package main

import (
	"errors"
	"github.com/clambin/go-metrics/server"
	"github.com/clambin/hostchecker/collector"
	"github.com/clambin/hostchecker/config"
	"github.com/clambin/hostchecker/version"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	var (
		debug      bool
		configName string
	)
	a := kingpin.New(filepath.Base(os.Args[0]), "hostchecker")
	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").Short('d').BoolVar(&debug)
	a.Flag("config", "Configuration file").Short('c').Default("/etc/hostchecker/config.yaml").ExistingFileVar(&configName)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	var f *os.File
	if f, err = os.Open(configName); err != nil {
		log.WithError(err).Fatal("failed to open config file")
	}

	var cfg *config.Config
	if cfg, err = config.Read(f); err != nil {
		log.WithError(err).Fatal("failed to load configuration")
	}

	_ = f.Close()

	if debug || cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	c := collector.New(cfg.Targets)
	prometheus.MustRegister(c)

	log.WithField("version", version.BuildVersion).Info("starting metrics server")
	s := server.NewWithHandlers(cfg.Port, []server.Handler{
		{
			Path: "/",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		},
	})
	if err = s.Run(); !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Fatal("failed to start metrics server")
	}
}
