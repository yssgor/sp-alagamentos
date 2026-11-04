#!/usr/bin/env bash
# Funções compartilhadas pelos workflows que gravam na branch `dados`.
#   bash .github/scripts/dados.sh preparar        # cria o worktree ./dados
#   bash .github/scripts/dados.sh commit "msg"    # commit + push (com retry)
set -euo pipefail

preparar() {
  git config --global user.name "github-actions[bot]"
  git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
  if git ls-remote --exit-code --heads origin dados > /dev/null; then
    git fetch origin dados:dados
    git worktree add dados dados
  else
    git worktree add --orphan -b dados dados
  fi
}

commit() {
  cd dados
  git add -A
  if git diff --cached --quiet; then
    echo "Sem mudanças."
    return 0
  fi
  local n
  n=$(git diff --cached --name-only | wc -l)
  git commit -m "$1 ($n arquivos)"
  # outro workflow pode ter gravado enquanto este rodava: rebase e tenta de novo
  for tentativa in 1 2 3 4 5; do
    if git ls-remote --exit-code --heads origin dados > /dev/null; then
      git pull --rebase origin dados
    fi
    if git push origin dados; then
      return 0
    fi
    echo "Push falhou (tentativa $tentativa), tentando de novo..."
    sleep $((tentativa * 5))
  done
  return 1
}

"$@"
