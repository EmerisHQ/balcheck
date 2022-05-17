package lcd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emerishq/balcheck/pkg/bech32"
	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/httpx"
)

type LCDClient struct {
	Endpoint string
	HRP      string
	HTTP     *httpx.Client
}

func NewClient(c checker.Chain) *LCDClient {
	return &LCDClient{
		Endpoint: c.LCDEndpoint,
		HRP:      c.HRP,
		HTTP:     httpx.NewClient(),
	}
}

func (c *LCDClient) Bech32Address(addr string) string {
	bech32Addr, err := bech32.HexEncode(c.HRP, addr)
	if err != nil {
		panic(fmt.Sprintf("cannot encode with bech32: %s", err))
	}
	return bech32Addr
}

func (c *LCDClient) BalancesURL(addr string) string {
	bech32Addr := c.Bech32Address(addr)
	return fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", c.Endpoint, bech32Addr)
}

func (c *LCDClient) Balances(ctx context.Context, addr string) (checker.Balances, error) {
	url := c.BalancesURL(addr)

	var lcdBalanceResponse lcdBalanceResponse
	res, err := c.HTTP.GetJson(ctx, url, &lcdBalanceResponse)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	// TODO: handle res pagination for accounts that have many denoms
	balances := make(checker.Balances)
	for _, b := range lcdBalanceResponse.Balances {
		balances[b.Denom] = b.Amount
	}
	return balances, nil
}

type lcdBalanceResponse struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
	Pagination struct {
		NextKey *string `json:"next_key"`
		Total   string  `json:"total"`
	} `json:"pagination"`
}

func (c *LCDClient) StakingBalancesURL(addr string) string {
	bech32Addr := c.Bech32Address(addr)
	return fmt.Sprintf("%s/cosmos/staking/v1beta1/delegations/%s", c.Endpoint, bech32Addr)
}

func (c *LCDClient) StakingBalances(ctx context.Context, addr string) (checker.Balances, error) {
	url := c.StakingBalancesURL(addr)
	var stakingRes StakingResponse
	res, err := c.HTTP.GetJson(ctx, url, &stakingRes)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		// staking balances returns 404 if the address never made a tx in the
		// chain
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	balances := make(checker.Balances)
	for _, d := range stakingRes.DelegationResponses {
		validatorAddress, err := bech32.HexDecode(d.Delegation.ValidatorAddress)
		if err != nil {
			panic(fmt.Sprintf("cannot decode validator address: %s", err))
		}
		balances[validatorAddress] = d.Balance.Amount
	}

	return balances, nil
}

type StakingResponse struct {
	DelegationResponses []struct {
		Delegation struct {
			ValidatorAddress string `json:"validator_address,omitempty"`
		} `json:"delegation,omitempty"`
		Balance struct {
			Amount string `json:"amount,omitempty"`
		} `json:"balance,omitempty"`
	} `json:"delegation_responses,omitempty"`
}

func (c *LCDClient) UnstakingBalancesURL(addr string) string {
	bech32Addr := c.Bech32Address(addr)
	return fmt.Sprintf("%s/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations", c.Endpoint, bech32Addr)
}

func (c *LCDClient) UnstakingBalances(ctx context.Context, addr string) (checker.Balances, error) {
	url := c.UnstakingBalancesURL(addr)
	var unstakingRes UnstakingResponse
	res, err := c.HTTP.GetJson(ctx, url, &unstakingRes)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	balances := make(checker.Balances)
	for _, d := range unstakingRes.UnbondingResponses {
		validatorAddress, err := bech32.HexDecode(d.ValidatorAddress)
		if err != nil {
			panic(fmt.Sprintf("cannot decode validator address: %s", err))
		}
		for _, entry := range d.Entries {
			key := fmt.Sprintf("%s_%s", validatorAddress, entry.CreationHeight)
			balances[key] = entry.Balance
		}
	}

	return balances, nil
}

type UnstakingResponse struct {
	UnbondingResponses []struct {
		ValidatorAddress string `json:"validator_address,omitempty"`
		Entries          []struct {
			Balance        string `json:"balance,omitempty"`
			CreationHeight string `json:"creation_height,omitempty"`
		} `json:"entries,omitempty"`
	} `json:"unbonding_responses,omitempty"`
}
