"""Acrescenta os endereços padronizados à tabela de ocorrências.

    uv run python -m sp_alagamentos.padronizar

Lê  data/processed/ocorrencias.csv              (gerado por `coletor processar`)
Grava data/processed/ocorrencias_padronizadas.csv, com as colunas originais e mais:

    local_pad   logradouro padronizado          ("AV EUSEBIO MATOSO")
    ref_tipo    cruzamento | numero | outro | vazio
    ref_pad     referência padronizada          ("PC JORGE DE LIMA", "1500"...)
    sentido_pad sentido padronizado             ("CENTRO/BAIRRO")
    ponto_id    identificador do lugar: mesmo local + mesma referência = mesmo ponto,
                não importa como o CGE escreveu
"""

from __future__ import annotations

import csv
import hashlib
import sys
from pathlib import Path

from sp_alagamentos import normalizar as n
from sp_alagamentos.caminhos import PROCESSED

COLUNAS_NOVAS = ["local_pad", "ref_tipo", "ref_pad", "sentido_pad", "ponto_id"]


def ponto_id(local: str, ref_tipo: str, ref: str) -> str:
    ref_chave = n.chave(ref) if ref_tipo == "cruzamento" else ref
    texto = f"{n.chave(local)}|{ref_tipo}:{ref_chave}"
    return hashlib.sha256(texto.encode()).hexdigest()[:10]


def padronizar(entrada: Path, saida: Path) -> tuple[int, int, int]:
    linhas = 0
    combinacoes_originais, pontos = set(), set()
    with entrada.open(encoding="utf-8", newline="") as fin, \
            saida.open("w", encoding="utf-8", newline="") as fout:
        leitor = csv.DictReader(fin)
        escritor = csv.DictWriter(fout, fieldnames=[*leitor.fieldnames, *COLUNAS_NOVAS])
        escritor.writeheader()
        for row in leitor:
            local = n.logradouro(row["local"])
            ref_tipo, ref = n.referencia(row["referencia"])
            pid = ponto_id(row["local"], ref_tipo, ref)
            row.update(local_pad=local, ref_tipo=ref_tipo, ref_pad=ref,
                       sentido_pad=n.sentido(row["sentido"]), ponto_id=pid)
            escritor.writerow(row)
            linhas += 1
            combinacoes_originais.add((row["local"], row["referencia"]))
            pontos.add(pid)
    return linhas, len(combinacoes_originais), len(pontos)


if __name__ == "__main__":
    entrada = Path(sys.argv[1]) if len(sys.argv) > 1 else PROCESSED / "ocorrencias.csv"
    saida = Path(sys.argv[2]) if len(sys.argv) > 2 else PROCESSED / "ocorrencias_padronizadas.csv"
    linhas, antes, depois = padronizar(entrada, saida)
    print(f"{linhas} ocorrências: {antes} grafias de local+referência → {depois} pontos → {saida}")
