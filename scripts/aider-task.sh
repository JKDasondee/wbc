#!/bin/bash
# Usage: ./scripts/aider-task.sh "your prompt here"
# Runs aider with local Qwen3:14b via Ollama on the wbc project
cd "$(dirname "$0")/.."
aider \
  --model ollama/qwen3:14b \
  --no-auto-commits \
  --yes \
  --no-git \
  --message "$1" \
  $2 $3 $4 $5 $6 $7 $8 $9
