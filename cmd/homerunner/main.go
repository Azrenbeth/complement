package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matrix-org/complement/internal/config"
	"github.com/matrix-org/complement/internal/docker"
	"github.com/sirupsen/logrus"
)

type Config struct {
	HomeserverLifetimeMins int
	Port                   int
	VersionCheckIterations int
	KeepBlueprints         []string
}

func NewConfig() *Config {
	cfg := &Config{
		HomeserverLifetimeMins: 30,
		Port:                   54321,
		VersionCheckIterations: 100,
		KeepBlueprints:         strings.Split(os.Getenv("HOMERUNNER_KEEP_BLUEPRINTS"), " "),
	}
	if val, _ := strconv.Atoi(os.Getenv("HOMERUNNER_LIFETIME_MINS")); val != 0 {
		cfg.HomeserverLifetimeMins = val
	}
	if val, _ := strconv.Atoi(os.Getenv("HOMERUNNER_PORT")); val != 0 {
		cfg.Port = val
	}
	if val, _ := strconv.Atoi(os.Getenv("HOMERUNNER_VER_CHECK_ITERATIONS")); val != 0 {
		cfg.VersionCheckIterations = val
	}
	return cfg
}

func cleanup(c *Config) {
	cfg := &config.Complement{
		BaseImageURI:           "nothing",
		DebugLoggingEnabled:    true,
		VersionCheckIterations: c.VersionCheckIterations,
		KeepBlueprints:         c.KeepBlueprints,
		BestEffort:             true,
	}
	builder, err := docker.NewBuilder(cfg)
	if err != nil {
		logrus.WithError(err).Fatalf("failed to run cleanup")
	}
	builder.Cleanup()
}

func main() {
	cfg := NewConfig()
	rt, err := NewRuntime(cfg)
	if err != nil {
		logrus.Fatalf("failed to setup new runtime: %s", err)
	}
	cleanup(cfg)

	srv := &http.Server{
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		Handler:      Routes(rt, cfg),
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
	}
	logrus.Infof("Homerunner listening on :%d with config %+v", cfg.Port, cfg)

	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatalf("ListenAndServe failed: %s", err)
	}
}
