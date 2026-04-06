package models

import "context"

// TxFetcher retrieves raw transactions from an on-chain data source.
type TxFetcher interface {
	FetchByWallet(ctx context.Context, addr string, from, to uint64) ([]Tx, error)
	FetchByContract(ctx context.Context, addr string, from, to uint64) ([]Tx, error)
}

// TxStore persists and retrieves normalized transactions locally.
type TxStore interface {
	SaveTxs(ctx context.Context, txs []Tx) error
	GetTxsByWallet(ctx context.Context, addr string) ([]Tx, error)
	GetTxsByContract(ctx context.Context, addr string) ([]Tx, error)
	GetAllWallets(ctx context.Context) ([]string, error)
	HasWallet(ctx context.Context, addr string) (bool, error)
	Close() error
}

// FeatureExtractor transforms raw transactions into feature vectors.
type FeatureExtractor interface {
	Extract(ctx context.Context, addr string, txs []Tx) (Feature, error)
	ExtractBatch(ctx context.Context, wallets map[string][]Tx) (map[string]Feature, error)
}

// Clusterer groups feature vectors into clusters.
type Clusterer interface {
	Fit(vectors [][]float64) ([]Cluster, error)
	Predict(vector []float64) (int, error)
}

// Profiler assigns behavioral labels to clusters.
type Profiler interface {
	Label(clusters []Cluster, features map[string]Feature) ([]Profile, error)
}
