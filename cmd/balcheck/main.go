package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/api/account"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/emeris-utils/configuration"
	"github.com/gorilla/mux"
)

func main() {
	var c Config
	configuration.ReadConfig(&c, "demeris-api", map[string]string{
		"ListenAddr": ":8000",
		"Debug":      "false",
	})

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
	r.HandleFunc("/check/{address}", account.CheckAddress(emerisClient, w)).Methods("GET")
	err := http.ListenAndServe(c.ListenAddr, r)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	ListenAddr string `validate:"required"`
	Debug      bool
}

func (c *Config) Validate() error {
	return nil
}
