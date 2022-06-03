// Package check implements the actual checks that can be run. Think of them as
// a sort of "unit tests".
package check

import (
	"context"
	"sync"

	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/pkg/lcd"
)

type BalanceMismatchErr struct {
	CheckName string
	ChainName string
	APIURL    string
	LCDURL    string

	WrappedError error
}

func (e *BalanceMismatchErr) Error() string {
	return e.WrappedError.Error()
}

func Balances(ctx context.Context, emerisClient *emeris.Client, chains []checker.Chain, addr string) []error {
	errsChan := make(chan error)
	wg := sync.WaitGroup{}

	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			lcdClient := lcd.NewClient(chain)
			err := checker.RunBalanceCheck(ctx, addr, emerisClient.Balances, lcdClient.Balances)
			if err != nil {
				errsChan <- &BalanceMismatchErr{
					CheckName:    "balance",
					ChainName:    chain.Name,
					APIURL:       emerisClient.BalancesURL(addr),
					LCDURL:       lcdClient.BalancesURL(addr),
					WrappedError: err,
				}
			}
			wg.Done()
		}(chain)
	}

	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			lcdClient := lcd.NewClient(chain)
			err := checker.RunBalanceCheck(ctx, addr, emerisClient.StakingBalances, lcdClient.StakingBalances)
			if err != nil {
				errsChan <- &BalanceMismatchErr{
					CheckName:    "staking balance",
					ChainName:    chain.Name,
					APIURL:       emerisClient.StakingBalancesURL(addr),
					LCDURL:       lcdClient.StakingBalancesURL(addr),
					WrappedError: err,
				}
			}
			wg.Done()
		}(chain)
	}

	for _, chain := range chains {
		wg.Add(1)
		go func(chain checker.Chain) {
			lcdClient := lcd.NewClient(chain)
			err := checker.RunBalanceCheck(ctx, addr, emerisClient.UnstakingBalances, lcdClient.UnstakingBalances)
			if err != nil {
				errsChan <- &BalanceMismatchErr{
					CheckName:    "unbonding balance",
					ChainName:    chain.Name,
					APIURL:       emerisClient.UnstakingBalancesURL(addr),
					LCDURL:       lcdClient.UnstakingBalancesURL(addr),
					WrappedError: err,
				}
			}
			wg.Done()
		}(chain)
	}

	go func() {
		wg.Wait()
		close(errsChan)
	}()

	var errs []error
	for err := range errsChan {
		errs = append(errs, err)
	}

	return errs
}
