## Phase 1: Foundation

- [x] @architect: Full system design — all interfaces, types, package layout (2026-04-07)
- [x] @infra: go mod tidy, verify build compiles (2026-04-07)
- [x] @infra: Wire CLI end-to-end — mapMembers(), profiler.Label(), table output (2026-04-07)
- [x] @analysis: copy_score via LCS ratio against whale tx sequences (2026-04-07)
- [x] @qa: Test fixtures — testdata/wallet_txs.json, 3 wallet archetypes (2026-04-07)

## Phase 2: Core Engine

- [ ] @pipeline: Normalize raw tx data, handle edge cases — ready
- [ ] @analysis: Feature extraction pipeline refinement — ready
- [ ] @analysis: K-means tuning + elbow method — ready
- [ ] @analysis: HDBSCAN tuning — ready

## Phase 3: Intelligence

- [ ] @analysis: Profile labeling engine refinement — blocked by Phase 2
- [ ] @infra: CSV output format for profile command — blocked by Phase 2
- [ ] @qa: Full test suite, benchmarks, CI — blocked by Phase 2

## Phase 4: Ship

- [ ] @infra: Dockerfile, README with architecture diagram — blocked by Phase 3
- [ ] @orchestrator: Final review, tag v0.1.0 — blocked by Phase 3
