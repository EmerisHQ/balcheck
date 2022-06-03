// Package checker contains the "engine" that can run checks. It is agnostic to
// the logic of "how to fetch data", it only contains the logic for "how to
// compare the data".
package checker

import (
	"context"
	"fmt"

	"github.com/damianopetrungaro/golog"
)

type BalanceProviderFunc func(ctx context.Context, addr string) (Balances, error)

func RunBalanceCheck(
	ctx context.Context,
	addr string,
	expected BalanceProviderFunc,
	actual BalanceProviderFunc,
) error {
	actualBalances, err := actual(ctx, addr)
	if err != nil {
		return fmt.Errorf("fetching actual balances: %w", err)
	}

	expectedBalances, err := expected(ctx, addr)
	if err != nil {
		return fmt.Errorf("fetching expected balances: %w", err)
	}

	err = expectedBalances.Contains(actualBalances)
	return err
}

type Chain struct {
	Name        string
	LCDEndpoint string
	HRP         string
}

type Balances map[string]string

func (b Balances) Contains(other Balances) error {
	for k, v := range b {
		if v != other[k] {
			return fmt.Errorf("%s: %s != %s", k, v, other[k])
		}
		golog.With(
			golog.String("key", k),
			golog.String("value", v),
		).Debug(context.Background(), "balance correct")
	}
	return nil
}
