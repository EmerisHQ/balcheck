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
	return func(response http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		chains, err := emerisClient.Chains(request.Context())
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			_, _ = response.Write([]byte(err.Error()))
			return
		}

		addr, err := bech32.HexDecode(vars["address"])
		if err != nil {
			response.WriteHeader(http.StatusBadRequest)
			_, _ = response.Write([]byte(err.Error()))
			return
		}

		_, _ = fmt.Fprint(response, "Started balance checking")

		go func() {
			errs := check.Balances(context.Background(), emerisClient, chains, addr)

			hub := sentry.GetHubFromContext(request.Context())
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
					).Error(request.Context(), "Error checking balance")
				}
			}
		}()
	}
}
