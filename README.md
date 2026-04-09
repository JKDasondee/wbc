# wbc

Wallet Behavior Classifier.

Go project for clustering on-chain wallets by behavioral features rather than by labels or token prices.

## Status

Alpha. Core CLI commands, feature extraction, clustering, and profiling are implemented. Documentation, CI, and larger-scale experiments still need work.

## What It Does

1. Ingests wallet or contract transactions from Etherscan into SQLite
2. Extracts behavioral features from transaction histories
3. Clusters wallets with from-scratch K-means or HDBSCAN
4. Produces profile labels such as `whale`, `retail_explorer`, `strategy_copier`, or `airdrop_farmer`

## Features

Current feature vector includes:

- transaction frequency
- protocol diversity
- average gas premium
- timing entropy
- value mean / std / skew
- repeated sequence pattern score
- first-interaction lag
- copy-score against known whale sequences

## CLI

```bash
wbc ingest --wallet 0x... --from-block 0 --to-block 99999999
wbc analyze --algorithm kmeans --k 5
wbc profile --format table
```

## Repo Layout

```text
wbc/
├── cmd/wbc/main.go
├── internal/
│   ├── classifier/
│   ├── config/
│   ├── ingester/
│   ├── profiler/
│   ├── store/
│   └── transformer/
├── pkg/
│   ├── ethapi/
│   └── models/
├── testdata/
└── PROGRESS.md
```

## Why This Repo Matters

- implements clustering logic in Go instead of relying on a Python notebook stack
- combines ingestion, storage, feature engineering, unsupervised learning, and profile labeling
- makes the feature definitions explicit so the behavioral assumptions are inspectable

## Caveats

- still an alpha repo
- labels are heuristic and exploratory, not ground-truth supervision
- Etherscan limits data quality and throughput
- clustering quality still needs larger validation runs and better reporting

## License

MIT