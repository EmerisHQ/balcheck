package utils

import (
	"context"

	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/pkg/lcd"
)

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
