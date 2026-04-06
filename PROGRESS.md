## Phase 1: Foundation

- [x] @architect: Full system design — all interfaces, types, package layout (2026-04-07)
  - pkg/models/types.go — all shared types (Tx, Wallet, Feature, Profile, Cluster, Labels, Opts)
  - pkg/models/interfaces.go — TxFetcher, TxStore, FeatureExtractor, Clusterer, Profiler
  - pkg/ethapi/client.go — Etherscan API client with rate limiter + retry
  - internal/ingester/ingester.go — orchestrates fetch + store with caching
  - internal/store/sqlite.go — SQLite-backed TxStore implementation
  - internal/transformer/transformer.go — all 8 feature extractors
  - internal/classifier/kmeans.go — K-means from scratch + elbow SSE
  - internal/classifier/hdbscan.go — HDBSCAN with MST + union-find
  - internal/profiler/profiler.go — centroid-based label assignment
  - internal/config/config.go — viper-based config loading
  - cmd/wbc/main.go — cobra CLI with ingest, analyze, profile commands
- [ ] @infra: go mod tidy, verify build compiles — ready
- [ ] @pipeline: Wire ingest command end-to-end with real API — ready

## Phase 2: Core Engine

- [ ] @pipeline: Normalize raw tx data, handle edge cases — blocked by Phase 1 completion
- [ ] @analysis: Feature extraction pipeline refinement — blocked by Phase 1
- [ ] @analysis: K-means tuning + elbow method — blocked by Phase 1
- [ ] @analysis: HDBSCAN tuning — blocked by Phase 1

## Phase 3: Intelligence

- [ ] @analysis: Profile labeling engine refinement — blocked by Phase 2
- [ ] @infra: Wire analyze + profile commands, output formatting — blocked by Phase 2
- [ ] @qa: Full test suite, benchmarks, CI — blocked by Phase 2

## Phase 4: Ship

- [ ] @infra: Dockerfile, README with architecture diagram — blocked by Phase 3
- [ ] @orchestrator: Final review, tag v0.1.0 — blocked by Phase 3
