package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/check"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/gorilla/mux"
)

func CheckAddress(emerisClient *emeris.Client) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		chains, err := emerisClient.Chains(request.Context())
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

		go func() {
			check.Balances(context.Background(), emerisClient, chains, addr)
		}()
	}
}
