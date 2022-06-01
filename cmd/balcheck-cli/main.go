package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/check"
	"github.com/emerishq/balcheck/pkg/emeris"
)

var (
	bech32Addr = flag.String("addr", "", "bech32 address to check (e.g. cosmos1qymla9gh8z2cmrylt008hkre0gry6h92sxgazg)")
	verbose    = flag.Bool("v", false, "enable verbose mode")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	logLevel := golog.ERROR
	if *verbose {
		logLevel = golog.DEBUG
	}

	w := golog.NewBufWriter(
		golog.NewTextEncoder(golog.DefaultTextConfig()),
		bufio.NewWriter(os.Stderr),
		golog.DefaultErrorHandler(),
		logLevel,
	)
	defer w.Flush()
	logger := golog.New(
		w,
		golog.NewLevelCheckerOption(logLevel),
	)
	golog.SetLogger(logger)

	if bech32Addr == nil || len(*bech32Addr) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	addr, err := bech32.HexDecode(*bech32Addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q is not a valid bech32 address\n", *bech32Addr)
		os.Exit(1)
	}

	emerisClient := emeris.NewClient()
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
