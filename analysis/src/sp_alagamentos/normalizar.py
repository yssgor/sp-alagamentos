"""Normalização dos endereços do CGE.

O CGE escreve o mesmo lugar de vários jeitos: "AV. EUSEBIO MATOSO (BUTANTA)",
"AV EUSEBIO MATOSO", "Av Eusébio Matoso,". Aqui tudo vira uma forma canônica:
maiúsculas, sem acento, sem pontuação, sem o bairro entre parênteses e com
tipos e títulos de logradouro abreviados sempre do mesmo jeito.

A forma canônica serve para agrupar e geocodificar; o texto original continua
guardado na base.
"""

from __future__ import annotations

import re
import unicodedata

# primeiro termo: tipo de logradouro
TIPOS = {
    "AV": "AV", "AVENIDA": "AV",
    "R": "R", "RUA": "R",
    "PC": "PC", "PCA": "PC", "PRACA": "PC",
    "VD": "VD", "VIADUTO": "VD",
    "TN": "TUN", "TUN": "TUN", "TUNEL": "TUN",
    "PTE": "PTE", "PT": "PTE", "PONTE": "PTE",
    "ES": "EST", "EST": "EST", "ESTR": "EST", "ESTRADA": "EST",
    "AL": "AL", "ALAMEDA": "AL",
    "TV": "TV", "TRAV": "TV", "TRAVESSA": "TV",
    "LG": "LG", "LGO": "LG", "LARGO": "LG",
    "MARG": "MARGINAL", "MARGINAL": "MARGINAL",
    "PQ": "PQ", "PARQUE": "PQ",
    "VL": "VL", "VIELA": "VL",
    "CV": "CV", "CMPV": "CV",  # complexo viário
    "VIA": "VIA",
    "PSG": "PSG", "PSS": "PSG", "PASSAGEM": "PSG",
    "RAD": "RADIAL", "RADIAL": "RADIAL",
    "ROD": "ROD", "RODOVIA": "ROD",
}

# demais termos: títulos e palavras comuns
TERMOS = {
    "DOUTOR": "DR", "DR": "DR",
    "DOUTORA": "DRA", "DRA": "DRA",
    "ENGENHEIRO": "ENG", "ENG": "ENG",
    "PROFESSOR": "PROF", "PROF": "PROF",
    "PROFESSORA": "PROFA", "PROFA": "PROFA",
    "CARDEAL": "CARD", "CARD": "CARD",
    "GENERAL": "GEN", "GAL": "GEN", "GEN": "GEN",
    "PRESIDENTE": "PRES", "PRES": "PRES",
    "CAPITAO": "CAP", "CAP": "CAP", "CAPM": "CAP",
    "CORONEL": "CEL", "CEL": "CEL",
    "MARECHAL": "MAL", "MAL": "MAL",
    "BRIGADEIRO": "BRIG", "BRIG": "BRIG",
    "DEPUTADO": "DEP", "DEP": "DEP",
    "SENADOR": "SEN", "SEN": "SEN",
    "GOVERNADOR": "GOV", "GOV": "GOV",
    "DESEMBARGADOR": "DES", "DES": "DES",
    "COMENDADOR": "COM", "COM": "COM",
    "MINISTRO": "MIN", "MIN": "MIN",
    "PADRE": "PE", "PE": "PE",
    "DONA": "D", "D": "D",
    "SANTA": "STA", "STA": "STA",
    "SANTO": "STO", "STO": "STO",
    "SAO": "S", "S": "S",
    "NOSSA": "NSA", "NSA": "NSA",
    "SENHORA": "SRA", "SRA": "SRA",
    "NSRA": "NSA SRA",
    "JARDIM": "JD", "JD": "JD",
    "VILA": "VL", "VL": "VL",
    "PONTE": "PTE", "PTE": "PTE",
    "VIADUTO": "VD", "VD": "VD",
    "AVENIDA": "AV", "AV": "AV",
    "RUA": "R",
    "PRACA": "PC", "PCA": "PC",
}

# preposições que o CGE às vezes omite ("AV NOVE JULHO" x "AV NOVE DE JULHO")
PREPOSICOES = {"DE", "DA", "DO", "DAS", "DOS", "E"}

_RE_PARENTESES = re.compile(r"\s*\([^)]*\)\s*")
_RE_CRUZAMENTO = re.compile(r"\s+X\s+")
_RE_NAO_ALFANUM = re.compile(r"[^A-Z0-9 ]+")


def sem_acento(texto: str) -> str:
    nfkd = unicodedata.normalize("NFKD", texto)
    return "".join(c for c in nfkd if not unicodedata.combining(c))


def _limpar(texto: str | None) -> str:
    if not texto:
        return ""
    # "Nº" antes de tirar acentos: o "º" viraria "O" e sobraria um "NO"
    t = re.sub(r"N\s*[º°]", " ", texto, flags=re.IGNORECASE)
    t = sem_acento(t).upper()
    t = _RE_PARENTESES.sub(" ", t)
    t = _RE_NAO_ALFANUM.sub(" ", t)
    return " ".join(t.split())


def _canonico(tokens: list[str]) -> list[str]:
    if not tokens:
        return tokens
    out = [TIPOS.get(tokens[0], tokens[0])]
    for tok in tokens[1:]:
        out.extend(TERMOS.get(tok, tok).split())
    return out


def logradouro(texto: str | None) -> str:
    """Forma canônica de um logradouro (campo `local` do CGE)."""
    t = _limpar(texto)
    t = _RE_CRUZAMENTO.split(t)[0]  # "AV X X PCA Y" → só a primeira via
    return " ".join(_canonico(t.split()))


def chave(texto: str | None) -> str:
    """Chave de comparação: forma canônica sem preposições."""
    return " ".join(tok for tok in logradouro(texto).split() if tok not in PREPOSICOES)


# referências que não são outra via, e sim uma posição ao longo da via
_RE_ALTURA = re.compile(r"^(ALTURA|ALT)( DO| DA| N| NUM| NUMERO)*\s+(\d+)")
_RE_NAO_VIA = re.compile(
    r"^(MEIO|INICIO|FINAL|FIM|JUNTO|SOB|SOBRE|TODA|TODO|ACESSO|AC|PROXIMO|PROX|"
    r"EM FRENTE|FRENTE|ENTRE|APOS|ANTES|DEBAIXO|EMBAIXO|NO|NA)\b"
)


def referencia(texto: str | None) -> tuple[str, str]:
    """Classifica e normaliza a referência.

    Devolve (tipo, valor):
      ("cruzamento", "PC JORGE DE LIMA")  — outra via; geocodificar pelo cruzamento
      ("numero", "1500")                  — altura do número na própria via
      ("outro", "MEIO DE QUADRA")         — posição sem via; geocodificar só pela via
      ("vazio", "")
    """
    t = _limpar(texto)
    if not t:
        return "vazio", ""
    if m := _RE_ALTURA.match(t):
        return "numero", m.group(3)
    if _RE_NAO_VIA.match(t) or t.split()[0] not in TIPOS:
        return "outro", t
    return "cruzamento", logradouro(t)


SENTIDOS = {
    "C/B": "CENTRO/BAIRRO",
    "B/C": "BAIRRO/CENTRO",
    "UNICO": "UNICO",
    "AMBOS UNICA": "AMBOS",
    "NAO INFORMADO": "",
}


def sentido(texto: str | None) -> str:
    t = sem_acento(texto or "").upper().strip()
    t = " ".join(t.replace("-", " ").split())
    return SENTIDOS.get(t, t)
