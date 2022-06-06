package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"os"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/check"
	"github.com/emerishq/balcheck/pkg/emeris"
)

var (
	fullAddr = flag.String("addr", "", "address to check (e.g. cosmos1qymla9gh8z2cmrylt008hkre0gry6h92sxgazg)")
	apiUrl   = flag.String("apiurl", "https://api.emeris.com/v1", "emeris api url (default https://api.emeris.com/v1)")
)

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

	emerisClient := emeris.NewClient(*apiUrl)
	chains, err := emerisClient.Chains(ctx)
	if err != nil {
		panic(err)
	}

	errs := check.Balances(ctx, emerisClient, chains, addr)
	for _, e := range errs {
		var mismatchErr *check.BalanceMismatchErr
		if errors.As(e, &mismatchErr) {
			golog.With(
				golog.String("check", mismatchErr.CheckName),
				golog.String("chain", mismatchErr.ChainName),
				golog.String("api_url", mismatchErr.APIURL),
				golog.String("lcd_url", mismatchErr.LCDURL),
				golog.Err(mismatchErr.WrappedError),
			).Error(ctx, "balance mismatch")
		} else {
			golog.With(golog.Err(e)).Error(ctx, "error")
		}
	}
}
