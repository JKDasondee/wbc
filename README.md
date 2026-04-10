# wbc

**On-chain wallet behavior classifier. EVM transaction feature extraction, K-means and HDBSCAN clustering, centroid-based label assignment. Written in Go.**

## Overview

Fetches raw EVM transaction history via Etherscan, extracts 8 behavioral features per wallet, clusters wallets into behavioral archetypes, and assigns human-readable labels (e.g. bot, whale, retail, copy-trader) based on cluster centroids. All three stages are available as independent CLI subcommands with JSON output.

## Pipeline

```
wbc ingest --wallet <addr>
  → ethapi.Etherscan        (rate-limited, cached to SQLite)
  → ingester                (dedup, block-range filtering)
  → store.SQLite

wbc analyze --algorithm kmeans --k 5
  → transformer.Extract()   (8 features per wallet)
        ├── TxFreq           txs per day
        ├── ProtocolDiv      unique contract addresses touched
        ├── AvgGasPremium    mean gas price paid
        ├── TimingEntropy    Shannon entropy of hour-of-day distribution
        ├── ValueMean/Std/Skew  ETH value transfer distribution
        ├── SeqPattern       top recurring contract-pair transition ratio
        └── CopyScore        LCS similarity to known whale sequences
  → classifier.KMeans       (from-scratch, elbow SSE)
  → classifier.HDBSCAN      (MST + union-find)
  → JSON cluster output

wbc profile
  → profiler.Label()        (centroid-based archetype assignment)
  → JSON per-wallet label + confidence
```

## Usage

```bash
# build
go build ./cmd/wbc

# ingest a wallet's transaction history
./wbc ingest --wallet 0xabc...123 --from-block 0 --to-block 99999999

# cluster all ingested wallets
./wbc analyze --algorithm kmeans --k 5
./wbc analyze --algorithm hdbscan --min-cluster 10

# generate behavioral profiles
./wbc profile --output table
./wbc profile --output json
```

Config via `wbc.yaml` (Viper): Etherscan API key, RPS limit, SQLite path, default cluster count.

## Stack

```
Go 1.23
cobra · viper · modernc/sqlite
Etherscan public API v2
```

MIT License
