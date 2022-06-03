package account

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/check"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
)

func CheckAddress(emerisClient *emeris.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		chains, err := emerisClient.Chains(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		addr, err := bech32.HexDecode(vars["address"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprint(w, "Started balance checking")

		go func() {
			errs := check.Balances(context.Background(), emerisClient, chains, addr)

			hub := sentry.GetHubFromContext(r.Context())
			for _, e := range errs {
				var mismatchErr *check.BalanceMismatchErr
				if errors.As(e, &mismatchErr) {
					hub.WithScope(func(scope *sentry.Scope) {
						scope.SetTag("check_name", mismatchErr.CheckName)
						scope.SetTag("chain_name", mismatchErr.ChainName)
						hub.CaptureException(e)
					})
				} else {
					golog.With(
						golog.Err(e),
					).Error(r.Context(), "Error checking balance")
				}
			}
		}()
	}
}
