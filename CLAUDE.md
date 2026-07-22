# CLAUDE.md

## Project Identity

This repository is positioned as **fuse**: a Go data fusion engine for heterogeneous event sources.

Primary value proposition:
- Ingest from mixed sources (CSV, JSONL, SQLite, HTTP APIs)
- Normalize and unify schema/types
- Query via SQL-like interface
- Detect anomalies in streaming and batch windows

Do not frame this project as a crypto classifier. Treat prior WBC components as architecture scaffolding for fuse.

## Target Outcomes

- Thesis-ready: reproducible experiments, benchmarkable correctness, documented methodology
- Hiring-ready: production software engineering signals (tests, CI gates, reliability, observability, security)

Resume framing:
- "Built a data fusion engine in Go that ingests heterogeneous sources, normalizes schemas, supports SQL queries, and performs real-time anomaly detection."

## Canonical Architecture Mapping

Legacy-to-fuse mapping used during migration:
- `pkg/ethapi/client.go` -> `internal/source/` adapters (CSV, JSONL, SQLite, HTTP)
- `internal/ingester/` -> `internal/ingester/` multi-source ingest orchestration
- `internal/transformer/` -> `internal/schema/` mapping and coercion
- `internal/classifier/` -> `internal/query/` SQL-subset parser + executor
- `internal/profiler/` -> `internal/anomaly/` rolling/z-score anomaly detection
- `internal/store/sqlite.go` -> `internal/store/` unified event store + indexing

## Milestone Order (Must Keep)

Execute in strict order:

1) **Milestone A: Quality Baseline**
- Unit + integration + golden tests
- CI gates (`go test -race`, `go vet`, `staticcheck`, build)
- Error-path hardening and command timeouts

2) **Milestone B: Correctness + Reproducibility**
- Query correctness proofs/tests
- Benchmark datasets and deterministic runs
- Persist run metadata (config hash, seed, input window, commit SHA)

3) **Milestone C: Operability**
- Structured logs and stage metrics
- Security checks and secret hygiene
- Architecture, testing, and operations docs

4) **Milestone D: Expansion**
- Federated sources
- Streaming watch mode
- HTTP API service mode

Never skip A to chase feature work.

## Definition of Done

A feature is done only when:
- Tests exist at the right level (unit/integration/CLI as relevant)
- CI passes with required checks
- Failure modes are explicit (no silent data drops)
- Logs/metrics are sufficient for troubleshooting
- Docs updated for behavior, flags, and constraints

## Engineering Guardrails

- Prefer deterministic behavior (seed control, stable fixtures, reproducible outputs)
- Use bounded contexts (`context.WithTimeout`) in command and I/O boundaries
- Validate all source adapter outputs before schema normalization
- Treat partial ingest/query failures as first-class results with counts and reasons
- Keep interfaces small and mockable; isolate external dependencies behind adapters

## Performance and Quality Targets

- Throughput target: progressive path toward 50K events/sec (document environment and method)
- Test coverage target: at least 85% in core packages
- Query correctness: golden suite for parser/executor behavior and edge cases
- Operability: command summaries include ingest counts, rejects, retries, and latency buckets

## Commands (Current and Expected)

Current baseline:
- `make build`
- `make test`
- `make lint`

As fuse matures, maintain copy-paste runnable commands in README and keep this file in sync.

## Anti-Patterns to Avoid

- Domain drift back to wallet/crypto narrative
- Feature additions without tests and CI enforcement
- Silent `continue` on extraction/query errors without accounting
- Non-deterministic benchmark claims

## Documentation Expectations

Maintain these docs as code evolves:
- `README.md` (problem, quickstart, architecture)
- `docs/ARCHITECTURE.md`
- `docs/TESTING.md`
- `docs/OPERATIONS.md`
- `docs/SECURITY.md`
- `docs/BENCHMARKS.md`

## Collaboration Notes

- Keep PRs milestone-scoped (A/B/C/D), not mixed.
- Prefer small, reviewable increments with measurable outcomes.
- Every benchmark claim must include dataset shape, machine profile, and command used.
