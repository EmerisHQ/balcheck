package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	"sync"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/pkg/lcd"
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

	var errors bool
	wg := sync.WaitGroup{}
	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			defer w.Flush()
			lcdClient := lcd.NewClient(chain)
			log := golog.With(
				golog.String("check", "balances"),
				golog.String("chain", chain.Name),
				golog.String("api_url", emerisClient.BalancesURL(addr)),
				golog.String("lcd_url", lcdClient.BalancesURL(addr)),
			)
			log.Info(ctx, "started testing")

			err := BalanceCheck(ctx, addr, lcdClient, emerisClient)
			if err != nil {
				errors = true
				log.With(golog.Err(err)).Error(ctx, "balance mismatch")
			}

			wg.Done()
		}(chain)
	}

	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			lcdClient := lcd.NewClient(chain)
			log := golog.With(
				golog.String("check", "staking balances"),
				golog.String("chain", chain.Name),
				golog.String("api_url", emerisClient.StakingBalancesURL(addr)),
				golog.String("lcd_url", lcdClient.StakingBalancesURL(addr)),
			)
			log.Info(ctx, "started testing")

			err := StakingBalanceCheck(ctx, addr, lcdClient, emerisClient)
			if err != nil {
				errors = true
				log.With(golog.Err(err)).Error(ctx, "staking balance mismatch")
			}

			wg.Done()
		}(chain)
	}

	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			lcdClient := lcd.NewClient(chain)
			log := golog.With(
				golog.String("check", "unbonding balances"),
				golog.String("chain", chain.Name),
				golog.String("api_url", emerisClient.UnstakingBalancesURL(addr)),
				golog.String("lcd_url", lcdClient.UnstakingBalancesURL(addr)),
			)
			log.Info(ctx, "started testing")

			err := UnstakingBalanceCheck(ctx, addr, lcdClient, emerisClient)
			if err != nil {
				errors = true
				log.With(golog.Err(err)).Error(ctx, "unbonding balance mismatch")
			}

			wg.Done()
		}(chain)
	}
	wg.Wait()

	if errors {
		golog.Fatal(ctx, "some checks failed")
	}
}

func BalanceCheck(
	ctx context.Context,
	addr string,
	lcdClient *lcd.LCDClient,
	emerisClient *emeris.Client,
) error {
	return checker.BalanceCheck(
		ctx,
		addr,
		emerisClient.Balances,
		lcdClient.Balances,
	)
}

func StakingBalanceCheck(
	ctx context.Context,
	addr string,
	lcdClient *lcd.LCDClient,
	emerisClient *emeris.Client,
) error {
	return checker.BalanceCheck(
		ctx,
		addr,
		emerisClient.StakingBalances,
		lcdClient.StakingBalances,
	)
}

func UnstakingBalanceCheck(
	ctx context.Context,
	addr string,
	lcdClient *lcd.LCDClient,
	emerisClient *emeris.Client,
) error {
	return checker.BalanceCheck(
		ctx,
		addr,
		emerisClient.UnstakingBalances,
		lcdClient.UnstakingBalances,
	)
}
