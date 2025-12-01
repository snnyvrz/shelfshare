#!/usr/bin/env bash
set -euo pipefail

for chart in charts/*; do
    if [ -d "$chart" ]; then
        helm lint "$chart"
    fi
done
