// Package ingester orchestrates transaction fetching and storage.
package ingester

import (
	"context"
	"fmt"

	"github.com/jaydasondee/wbc/pkg/models"
)

type Ingester struct {
	fetcher models.TxFetcher
	store   models.TxStore
}

func New(f models.TxFetcher, s models.TxStore) *Ingester {
	return &Ingester{fetcher: f, store: s}
}

func (ig *Ingester) IngestWallet(ctx context.Context, addr string, from, to uint64) (int, error) {
	cached, err := ig.store.HasWallet(ctx, addr)
	if err != nil {
		return 0, fmt.Errorf("ingester.IngestWallet: %w", err)
	}
	if cached {
		txs, err := ig.store.GetTxsByWallet(ctx, addr)
		if err != nil {
			return 0, fmt.Errorf("ingester.IngestWallet: %w", err)
		}
		return len(txs), nil
	}

	txs, err := ig.fetcher.FetchByWallet(ctx, addr, from, to)
	if err != nil {
		return 0, fmt.Errorf("ingester.IngestWallet: %w", err)
	}

	if err := ig.store.SaveTxs(ctx, txs); err != nil {
		return 0, fmt.Errorf("ingester.IngestWallet: %w", err)
	}
	return len(txs), nil
}

func (ig *Ingester) IngestContract(ctx context.Context, addr string, from, to uint64) (int, error) {
	txs, err := ig.fetcher.FetchByContract(ctx, addr, from, to)
	if err != nil {
		return 0, fmt.Errorf("ingester.IngestContract: %w", err)
	}

	if err := ig.store.SaveTxs(ctx, txs); err != nil {
		return 0, fmt.Errorf("ingester.IngestContract: %w", err)
	}
	return len(txs), nil
}
