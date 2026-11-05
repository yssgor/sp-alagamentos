"""Baixa camadas do GeoSampa (WFS da Prefeitura de São Paulo).

    python geosampa.py listar [filtro...]     # lista camadas cujo nome/título contém o filtro
    python geosampa.py baixar CAMADA SAIDA    # baixa a camada inteira em GeoJSON (EPSG:4326), gzip

Só biblioteca padrão. Pagina com startIndex/count porque o servidor limita
o número de feições por requisição.
"""

import gzip
import json
import sys
import time
import urllib.parse
import urllib.request
import xml.etree.ElementTree as ET

WFS = "http://wfs.geosampa.prefeitura.sp.gov.br/geoserver/geoportal/wfs"
UA = "sp-alagamentos/0.1 (+https://github.com/yssgor/sp-alagamentos)"
PAGINA = 5000


def get(params: dict) -> bytes:
    url = WFS + "?" + urllib.parse.urlencode(params)
    for tentativa in range(5):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": UA})
            with urllib.request.urlopen(req, timeout=180) as r:
                return r.read()
        except Exception as e:  # noqa: BLE001
            print(f"  erro ({e}), tentativa {tentativa + 1}", file=sys.stderr)
            time.sleep(10 * (tentativa + 1))
    raise SystemExit(f"falhou: {url}")


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
    feicoes, inicio = [], 0
    while True:
        dados = json.loads(get({
            "service": "WFS", "version": "2.0.0", "request": "GetFeature",
            "typeNames": camada, "outputFormat": "application/json",
            "srsName": "EPSG:4326", "count": PAGINA, "startIndex": inicio,
        }))
        lote = dados.get("features", [])
        feicoes.extend(lote)
        print(f"  {camada}: {len(feicoes)} feições", file=sys.stderr)
        if len(lote) < PAGINA:
            break
        inicio += PAGINA
        time.sleep(2)
    with gzip.open(saida, "wt", encoding="utf-8") as f:
        json.dump({"type": "FeatureCollection", "features": feicoes}, f, ensure_ascii=False)
    print(f"| `{camada}` | {len(feicoes)} feições | `{saida}` |")


if __name__ == "__main__":
    cmd, *args = sys.argv[1:] or ["listar"]
    if cmd == "listar":
        listar(args)
    elif cmd == "baixar" and len(args) == 2:
        baixar(*args)
    else:
        raise SystemExit(__doc__)
