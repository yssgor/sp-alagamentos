from sp_alagamentos import normalizar as n


def test_mesmo_logradouro_grafias_diferentes():
    grafias = [
        "AV EUSEBIO MATOSO",
        "AV. EUSEBIO MATOSO",
        "AV EUSEBIO MATOSO,",
        "AV. EUSEBIO MATOSO (BUTANTA)",
        "Av Eusébio Matoso",
        "AVENIDA EUSEBIO MATOSO",
        "AV EUSEBIO MATOSO X PCA EUGENE BOUDIN",
    ]
    assert {n.logradouro(g) for g in grafias} == {"AV EUSEBIO MATOSO"}


def test_tipos_e_titulos():
    assert n.logradouro("PÇA. JORGE DE LIMA") == "PC JORGE DE LIMA"
    assert n.logradouro("PCA JORGE DE LIMA") == "PC JORGE DE LIMA"
    assert n.logradouro("R. CAPM GONCALO MONTEIRO") == "R CAP GONCALO MONTEIRO"
    assert n.logradouro("R. CARDEAL ARCOVERDE") == n.logradouro("R CARD ARCOVERDE")
    assert n.logradouro("TÚNEL ANHANGABAÚ") == n.logradouro("TN ANHANGABAU")
    assert n.logradouro("*AV. DO ESTADO") == "AV DO ESTADO"


def test_chave_ignora_preposicao():
    assert n.chave("AV NOVE DE JULHO") == n.chave("AV. NOVE JULHO")


def test_referencia():
    assert n.referencia("PC. JORGE DE LIMA") == ("cruzamento", "PC JORGE DE LIMA")
    assert n.referencia("ALTURA DO Nº 1500") == ("numero", "1500")
    assert n.referencia("ALT. 320") == ("numero", "320")
    assert n.referencia("MEIO DE QUADRA")[0] == "outro"
    assert n.referencia(None) == ("vazio", "")
    assert n.referencia("PT DA VILA GUILHERME") == ("cruzamento", "PTE DA VL GUILHERME")
    assert n.referencia("CMPV JOAO JORGE SAAD")[0] == "cruzamento"


def test_sentido():
    assert n.sentido("C/B") == "CENTRO/BAIRRO"
    assert n.sentido("ÚNICO") == n.sentido("UNICO")
    assert n.sentido("AMBOS - ÚNICA") == "AMBOS"
