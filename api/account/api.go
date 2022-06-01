package account

import (
	"context"
	"fmt"
	"net/http"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/emeris"
	"github.com/emerishq/balcheck/utils"
	"github.com/gorilla/mux"
)

func CheckAddress(emerisClient *emeris.Client, w *golog.BufWriter) func(http.ResponseWriter,
	*http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx := context.Background()
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

		utils.CheckBalances(emerisClient, chains, w, addr, false)
	}
}
