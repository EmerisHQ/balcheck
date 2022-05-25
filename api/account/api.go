package account

import (
	"fmt"
	"net/http"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/pkg/lcd"
	"github.com/gorilla/mux"
)

func CheckAddress(emerisClient *emeris.Client, w *golog.BufWriter) func(http.ResponseWriter,
	*http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		vars := mux.Vars(request)
		chains, err := emerisClient.Chains(ctx)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(err.Error()))
			return
		}

		addr, err := bech32.HexDecode(vars["address"])
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte(err.Error()))
			return
		}

		fmt.Fprint(response, "Started balance checking")

		for _, chain := range chains {
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

				err := checker.BalanceCheck(ctx, addr, emerisClient.Balances, lcdClient.Balances)
				if err != nil {
					log.With(golog.Err(err)).Error(ctx, "balance mismatch")
				}
			}(chain)
		}

		for _, chain := range chains {
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
			}(chain)
		}

		for _, chain := range chains {
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
			}(chain)
		}
	}
}
