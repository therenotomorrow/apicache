#!/usr/bin/env bash

set -e

make docs code test/smoke

git add .

ln -sf ../../scripts/pre-commit.sh .git/hooks/pre-commit
