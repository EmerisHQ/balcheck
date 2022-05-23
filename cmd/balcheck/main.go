package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/api/account"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/gorilla/mux"
)

var listenAdrr = flag.String("listen-addr", ":8081", "address to start http server (default localhost:8081)")

func main() {
	flag.Parse()

	w := golog.NewBufWriter(
		golog.NewJsonEncoder(golog.DefaultJsonConfig()),
		bufio.NewWriter(os.Stderr),
		golog.DefaultErrorHandler(),
		golog.DEBUG,
	)
	defer w.Flush()
	logger := golog.New(
		w,
		golog.NewLevelCheckerOption(golog.DEBUG),
	)
	golog.SetLogger(logger)

	serveAddr := ":8081"

	if listenAdrr != nil && len(*listenAdrr) != 0 {
		serveAddr = *listenAdrr
	}

	emerisClient := emeris.NewClient()

	fmt.Printf("Starting server on %s\n", serveAddr)

	r := mux.NewRouter()
	r.HandleFunc("/check/{address}", account.CheckAddress(emerisClient, w)).Methods("GET")
	err := http.ListenAndServe(serveAddr, r)
	if err != nil {
		panic(err)
	}
}
