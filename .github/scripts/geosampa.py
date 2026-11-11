"""Baixa camadas do GeoSampa (WFS da Prefeitura de São Paulo).

    python geosampa.py listar [filtro...]     # lista camadas cujo nome/título contém o filtro
    python geosampa.py baixar CAMADA SAIDA    # baixa a camada inteira em GeoJSON (EPSG:4326), gzip

Só biblioteca padrão. O servidor corta respostas grandes, então a camada vem
em páginas pequenas; cada página é validada e repetida se vier incompleta.
No fim, o total baixado é comparado com o total que o servidor informa.
Erros também são escritos no resumo do GitHub Actions ($GITHUB_STEP_SUMMARY).
"""

import gzip
import json
import os
import re
import sys
import time
import traceback
import urllib.parse
import urllib.request
import xml.etree.ElementTree as ET

WFS = "http://wfs.geosampa.prefeitura.sp.gov.br/geoserver/geoportal/wfs"
UA = "sp-alagamentos/0.1 (+https://github.com/yssgor/sp-alagamentos)"
PAGINA = 1000
TENTATIVAS = 5


def log(msg: str) -> None:
    print(msg, file=sys.stderr, flush=True)


def get(params: dict, validar=None):
    """GET no WFS. `validar` converte a resposta e levanta exceção se ela vier
    incompleta; nesse caso a requisição é repetida."""
    url = WFS + "?" + urllib.parse.urlencode(params)
    for tentativa in range(1, TENTATIVAS + 1):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": UA})
            with urllib.request.urlopen(req, timeout=180) as r:
                corpo = r.read()
            return validar(corpo) if validar else corpo
        except Exception as e:  # noqa: BLE001
            log(f"  tentativa {tentativa}/{TENTATIVAS} falhou: {type(e).__name__}: {str(e)[:200]}")
            if tentativa < TENTATIVAS:
                time.sleep(10 * tentativa)
    raise RuntimeError(f"falhou {TENTATIVAS} vezes: {url}")


def json_valido(corpo: bytes) -> dict:
    texto = corpo.decode("utf-8")
    if texto.lstrip().startswith("<"):  # o GeoServer devolve erros em XML
        raise ValueError("servidor respondeu XML em vez de JSON: " + texto[:300])
    return json.loads(texto)  # JSON cortado → JSONDecodeError → nova tentativa


def total_no_servidor(camada: str) -> int | None:
    xml = get({"service": "WFS", "version": "2.0.0", "request": "GetFeature",
               "typeNames": camada, "resultType": "hits"}).decode("utf-8", "replace")
    m = re.search(r'numberMatched="(\d+)"', xml)
    return int(m.group(1)) if m else None


def listar(filtros: list[str]) -> None:
    xml = get({"service": "WFS", "version": "2.0.0", "request": "GetCapabilities"})
    raiz = ET.fromstring(xml)
    ns = {"wfs": "http://www.opengis.net/wfs/2.0"}
    filtros = [f.lower() for f in filtros] or [""]
    linhas = []
    for ft in raiz.iter("{http://www.opengis.net/wfs/2.0}FeatureType"):
        nome = ft.findtext("wfs:Name", "", ns)
        titulo = ft.findtext("wfs:Title", "", ns)
        texto = (nome + " " + titulo).lower()
        if any(f in texto for f in filtros):
            linhas.append(f"| `{nome}` | {titulo} |")
    print("| camada | título |\n|---|---|")
    print("\n".join(sorted(linhas)) or "| (nenhuma) | |")


def baixar(camada: str, saida: str) -> None:
    esperado = total_no_servidor(camada)
    log(f"{camada}: servidor informa {esperado} feições")

    feicoes, vistos, repetidas, inicio = [], set(), 0, 0
    while True:
        dados = get({
            "service": "WFS", "version": "2.0.0", "request": "GetFeature",
            "typeNames": camada, "outputFormat": "application/json",
            "srsName": "EPSG:4326", "count": PAGINA, "startIndex": inicio,
        }, validar=json_valido)
        lote = dados.get("features", [])
        for f in lote:
            fid = f.get("id") or json.dumps(f.get("properties"), sort_keys=True)
            if fid in vistos:
                repetidas += 1
                continue
            vistos.add(fid)
            feicoes.append(f)
        log(f"  {camada}: {len(feicoes)}/{esperado} feições")
        if len(lote) < PAGINA:
            break
        inicio += PAGINA
        time.sleep(1)

    if esperado is not None and len(feicoes) != esperado:
        raise RuntimeError(f"{camada}: servidor informa {esperado} feições, chegaram {len(feicoes)} "
                           f"({repetidas} repetidas descartadas)")

    with gzip.open(saida, "wt", encoding="utf-8") as f:
        json.dump({"type": "FeatureCollection", "features": feicoes}, f, ensure_ascii=False)
    mb = os.path.getsize(saida) / 1e6
    print(f"| `{camada}` | {len(feicoes)} de {esperado} | {repetidas} | `{saida}` ({mb:.1f} MB) |")


def resumo_erro(msg: str) -> None:
    caminho = os.environ.get("GITHUB_STEP_SUMMARY")
    if caminho:
        with open(caminho, "a", encoding="utf-8") as f:
            f.write(f"\n### ❌ Erro\n\n```\n{msg}\n```\n")


if __name__ == "__main__":
    cmd, *args = sys.argv[1:] or ["listar"]
    try:
        if cmd == "listar":
            listar(args)
        elif cmd == "baixar" and len(args) == 2:
            baixar(*args)
        else:
            raise SystemExit(__doc__)
    except Exception:  # noqa: BLE001
        msg = traceback.format_exc(limit=3)
        resumo_erro(msg)
        log(msg)
        sys.exit(1)
