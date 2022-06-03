package emeris

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/damianopetrungaro/golog"
	"github.com/emerishq/balcheck/pkg/checker"
	"github.com/emerishq/balcheck/pkg/httpx"
)

type Client struct {
	BaseURL string
	HTTP    *httpx.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL: "https://api.emeris.com",
		HTTP:    httpx.NewClient(),
	}
}

func (c *Client) BalancesURL(addr string) string {
	return fmt.Sprintf("%s/v1/account/%s/balance", c.BaseURL, addr)
}

func (c *Client) Balances(ctx context.Context, addr string) (checker.Balances, error) {
	url := c.BalancesURL(addr)
	golog.With(
		golog.String("url", url),
	).Info(ctx, "fetching balances from emeris")

	var balancesRes BalancesResponse
	res, err := c.HTTP.GetJson(ctx, url, &balancesRes)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	balances := make(checker.Balances)
	for _, b := range balancesRes.Balances {
		// Native denoms
		if strings.HasSuffix(b.Amount, b.BaseDenom) {
			denom := b.BaseDenom
			amount := strings.SplitN(b.Amount, denom, 2)[0]
			balances[denom] = amount
			continue
		}

		// IBC denoms
		amountDenom := strings.SplitN(b.Amount, "ibc/", 2)
		amount := amountDenom[0]
		denom := "ibc/" + amountDenom[1]
		balances[denom] = amount
	}
	return balances, nil
}

type BalancesResponse struct {
	Balances []struct {
		BaseDenom string `json:"base_denom"`
		Amount    string `json:"amount"`
	} `json:"balances"`
}

func (c *Client) Chains(ctx context.Context) ([]checker.Chain, error) {
	url := c.BaseURL + "/v1/chains"

	golog.With(
		golog.String("url", url),
	).Info(ctx, "fetching chains from emeris")

	var chainsRes ChainsResponse
	res, err := c.HTTP.GetJson(ctx, url, &chainsRes)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	chains := make([]checker.Chain, 0, len(chainsRes.Chains))
	for _, c := range chainsRes.Chains {
		log := golog.With(
			golog.String("chain", c.ChainName),
		)

		if !c.Enabled || !c.Online {
			log.With(
				golog.Bool("enabled", c.Enabled),
				golog.Bool("online", c.Online),
			).Warn(ctx, "skipping chain")
			continue
		}

		if len(c.PublicNodeEndpoints.CosmosAPI) == 0 {
			log.Warn(ctx, "no LCD endpoints configured")
			continue
		}

		chains = append(chains, checker.Chain{
			Name:        c.ChainName,
			LCDEndpoint: c.PublicNodeEndpoints.CosmosAPI[0],
			HRP:         c.NodeInfo.Bech32Config.MainPrefix,
		})
	}

	return chains, nil
}

type ChainsResponse struct {
	Chains []struct {
		ChainName string `json:"chain_name"`

		Enabled bool `json:"enabled"`
		Online  bool `json:"online"`

		PublicNodeEndpoints struct {
			CosmosAPI []string `json:"cosmos_api"`
		} `json:"public_node_endpoints"`
		NodeInfo struct {
			Bech32Config struct {
				MainPrefix string `json:"main_prefix"`
			} `json:"bech32_config"`
		} `json:"node_info"`
	} `json:"chains"`
}

func (c *Client) StakingBalancesURL(addr string) string {
	return fmt.Sprintf("%s/v1/account/%s/stakingbalances", c.BaseURL, addr)
}

func (c *Client) StakingBalances(ctx context.Context, addr string) (checker.Balances, error) {
	url := c.StakingBalancesURL(addr)
	var stakingRes StakingBalancesResponse
	res, err := c.HTTP.GetJson(ctx, url, &stakingRes)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	balances := make(checker.Balances)
	for _, b := range stakingRes.StakingBalances {
		// strip useless decimals
		amount := strings.SplitN(b.Amount, ".", 2)[0]
		balances[b.ValidatorAddress] = amount
	}
	return balances, nil
}

type StakingBalancesResponse struct {
	StakingBalances []struct {
		ValidatorAddress string `json:"validator_address"`
		Amount           string `json:"amount"`
	} `json:"staking_balances"`
}

func (c *Client) UnstakingBalancesURL(addr string) string {
	return fmt.Sprintf("%s/v1/account/%s/unbondingdelegations", c.BaseURL, addr)
}

func (c *Client) UnstakingBalances(ctx context.Context, addr string) (checker.Balances, error) {
	url := c.UnstakingBalancesURL(addr)
	var unstakingRes UnstakingBalancesResponse
	res, err := c.HTTP.GetJson(ctx, url, &unstakingRes)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	balances := make(checker.Balances)
	for _, d := range unstakingRes.UnbondingDelegations {
		for _, entry := range d.Entries {
			key := fmt.Sprintf("%s_%d", d.ValidatorAddress, entry.CreationHeight)
			balances[key] = entry.Balance
		}
	}
	return balances, nil
}

type UnstakingBalancesResponse struct {
	UnbondingDelegations []struct {
		ValidatorAddress string `json:"validator_address"`
		Entries          []struct {
			Balance        string `json:"balance"`
			CreationHeight int    `json:"creation_height"`
		} `json:"entries"`
	} `json:"unbonding_delegations"`
}
