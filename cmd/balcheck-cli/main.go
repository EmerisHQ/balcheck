package main

import (
	"bufio"
	"context"
	"flag"
	"os"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/utils"
)

var fullAddr = flag.String("addr", "", "address to check (e.g. cosmos1qymla9gh8z2cmrylt008hkre0gry6h92sxgazg)")

func main() {
	flag.Parse()

	ctx := context.Background()

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

	if fullAddr == nil || len(*fullAddr) == 0 {
		logger.Fatal(ctx, "missing address")
	}
	addr, err := bech32.HexDecode(*fullAddr)
	if err != nil {
		panic(err)
	}

	emerisClient := emeris.NewClient()
	chains, err := emerisClient.Chains(ctx)
	if err != nil {
		panic(err)
	}

	utils.CheckBalances(emerisClient, chains, w, addr, true)
}
