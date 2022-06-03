package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/api/account"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/emeris-utils/configuration"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
)

var Version = "not specified"

func main() {
	var c Config
	configuration.ReadConfig(&c, "demeris-api", map[string]string{
		"ListenAddr":             ":8000",
		"Debug":                  "false",
		"SentryEnvironment":      "notset",
		"SentrySampleRate":       "1.0",
		"SentryTracesSampleRate": "0.3",
	})

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              c.SentryDSN,
		Release:          Version,
		SampleRate:       c.SentrySampleRate,
		TracesSampleRate: c.SentryTracesSampleRate,
		Environment:      c.SentryEnvironment,
		AttachStacktrace: true,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	defer sentry.Flush(2 * time.Second)

	w := golog.NewBufWriter(
		golog.NewJsonEncoder(golog.DefaultJsonConfig()),
		bufio.NewWriter(os.Stderr),
		golog.DefaultErrorHandler(),
		golog.DEBUG,
	)
	defer w.Flush()
	minLogLevel := golog.INFO
	if c.Debug {
		minLogLevel = golog.DEBUG
	}
	logger := golog.New(
		w,
		golog.NewLevelCheckerOption(minLogLevel),
	)
	golog.SetLogger(logger)

	emerisClient := emeris.NewClient()

	fmt.Printf("Starting server on %s\n", c.ListenAddr)

	r := mux.NewRouter()
	r.HandleFunc("/check/{address}", account.CheckAddress(emerisClient)).Methods("GET")
	err := http.ListenAndServe(c.ListenAddr, r)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	ListenAddr             string `validate:"required"`
	Debug                  bool
	SentryDSN              string
	SentryEnvironment      string
	SentrySampleRate       float64
	SentryTracesSampleRate float64
}

func (c *Config) Validate() error {
	return nil
}
