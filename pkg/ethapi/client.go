// Package ethapi provides Etherscan and Alchemy API clients.
package ethapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jaydasondee/wbc/pkg/models"
)

type RateLimiter struct {
	mu       sync.Mutex
	last     time.Time
	interval time.Duration
}

func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{interval: time.Second / time.Duration(rps)}
}

func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if d := r.interval - now.Sub(r.last); d > 0 {
		time.Sleep(d)
	}
	r.last = time.Now()
}

type EtherscanClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
	rl      *RateLimiter
}

func NewEtherscan(apiKey, baseURL string, rps int) *EtherscanClient {
	return &EtherscanClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
		rl:      NewRateLimiter(rps),
	}
}

type etherscanResp struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Result  []json.RawMessage `json:"result"`
}

type etherscanTx struct {
	Hash            string `json:"hash"`
	From            string `json:"from"`
	To              string `json:"to"`
	Value           string `json:"value"`
	Gas             string `json:"gas"`
	GasPrice        string `json:"gasPrice"`
	BlockNumber     string `json:"blockNumber"`
	TimeStamp       string `json:"timeStamp"`
	Input           string `json:"input"`
	IsError         string `json:"isError"`
	ContractAddress string `json:"contractAddress"`
}

func (c *EtherscanClient) FetchByWallet(ctx context.Context, addr string, from, to uint64) ([]models.Tx, error) {
	url := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=%d&endblock=%d&sort=asc&apikey=%s",
		c.baseURL, addr, from, to, c.apiKey)
	return c.fetch(ctx, url, "ethereum")
}

func (c *EtherscanClient) FetchByContract(ctx context.Context, addr string, from, to uint64) ([]models.Tx, error) {
	url := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=%d&endblock=%d&sort=asc&apikey=%s",
		c.baseURL, addr, from, to, c.apiKey)
	return c.fetch(ctx, url, "ethereum")
}

func (c *EtherscanClient) fetch(ctx context.Context, url, chain string) ([]models.Tx, error) {
	var txs []models.Tx
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		c.rl.Wait()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("ethapi.fetch: %w", err)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return nil, fmt.Errorf("ethapi.fetch: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("ethapi.fetch: %w", err)
		}

		var r etherscanResp
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, fmt.Errorf("ethapi.fetch: %w", err)
		}

		if r.Status != "1" {
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return nil, fmt.Errorf("ethapi.fetch: api error: %s", r.Message)
		}

		for _, raw := range r.Result {
			var et etherscanTx
			if err := json.Unmarshal(raw, &et); err != nil {
				continue
			}
			tx, err := parseTx(et, chain)
			if err != nil {
				continue
			}
			txs = append(txs, tx)
		}
		return txs, nil
	}
	return txs, nil
}

func parseTx(et etherscanTx, chain string) (models.Tx, error) {
	val := parseFloat(et.Value)
	gas := parseUint(et.Gas)
	gasPrice := parseFloat(et.GasPrice)
	blk := parseUint(et.BlockNumber)
	ts := parseUnix(et.TimeStamp)

	return models.Tx{
		Hash:         et.Hash,
		From:         et.From,
		To:           et.To,
		Value:        val / 1e18,
		Gas:          gas,
		GasPrice:     gasPrice / 1e9,
		BlockNum:     blk,
		Timestamp:    ts,
		Input:        et.Input,
		IsError:      et.IsError == "1",
		ContractAddr: et.ContractAddress,
		Chain:        chain,
	}, nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseUint(s string) uint64 {
	var u uint64
	fmt.Sscanf(s, "%d", &u)
	return u
}

func parseUnix(s string) time.Time {
	var ts int64
	fmt.Sscanf(s, "%d", &ts)
	return time.Unix(ts, 0)
}
