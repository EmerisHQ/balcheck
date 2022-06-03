package utils

import (
	"context"
	"sync"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/pkg/lcd"
)

func CheckBalances(ctx context.Context, emerisClient *emeris.Client, chains []checker.Chain, addr string) {
	wg := sync.WaitGroup{}

	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			lcdClient := lcd.NewClient(chain)
			log := golog.With(
				golog.String("check", "balances"),
				golog.String("chain", chain.Name),
				golog.String("api_url", emerisClient.BalancesURL(addr)),
				golog.String("lcd_url", lcdClient.BalancesURL(addr)),
			)
			log.Info(ctx, "started testing")

			err := checker.BalanceCheck(ctx, addr, emerisClient.Balances, lcdClient.Balances)
			if err != nil {
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

			err := checker.BalanceCheck(ctx, addr, emerisClient.StakingBalances, lcdClient.StakingBalances)
			if err != nil {
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

			err := checker.BalanceCheck(ctx, addr, emerisClient.UnstakingBalances, lcdClient.UnstakingBalances)
			if err != nil {
				log.With(golog.Err(err)).Error(ctx, "unbonding balance mismatch")
			}
			wg.Done()
		}(chain)
	}

	wg.Wait()
}
