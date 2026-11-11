# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="../../lingue/fr/README.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="../../lingue/de/README.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="../../lingue/es/README.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub><b>Português</b></sub></td>
<td align="center"><a href="../../lingue/ru/README.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="../../lingue/zh/README.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="../../lingue/ja/README.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="../../lingue/ko/README.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="../../lingue/ar/README.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="../../lingue/ms/README.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="../../lingue/hi/README.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/IU0VYT/IU0VYT/releases/latest"><img src="https://img.shields.io/github/v/release/IU0VYT/IU0VYT?label=release&color=2ea44f" alt="Latest release"></a>
<a href="https://github.com/IU0VYT/IU0VYT/releases"><img src="https://img.shields.io/github/downloads/IU0VYT/IU0VYT/total?color=1f6feb" alt="Downloads"></a>
<a href="../../LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0--or--later-blue" alt="License GPL-3.0-or-later"></a>
<img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d6" alt="Windows 10 | 11">
<img src="https://img.shields.io/badge/languages-13-orange" alt="13 languages">
</p>

**Versão 4.22.0** · Windows · interface em 13 línguas (italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語, 한국어, العربية, Bahasa Melayu, हिन्दी) · um só ficheiro executável, sem instalação

Programa para radioamadores e utilizadores de transceptores de **47 marcas** (**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone, Yaesu, Icom, Kenwood): lê e escreve canais, frequências e definições pelo cabo de programação, gere um arquivo dos repetidores italianos e atualiza o firmware dos rádios suportados.

🌐 Site do programa: **[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*Esta é a tradução portuguesa de [README.it.md](../../README.it.md); o original italiano é o texto de referência.*

## Funções

- **Canais e frequências** de 510 modelos (incluindo rádios PMR digitais e DMR, também por USB): leitura e escrita por cabo, ficheiros `.img` compatíveis com o CHIRP, ficheiros originais do Baofeng K67 Pro (`.xlc`) e do Bajeton 8300 (`.dat`).
- **Definições dos rádios** (menus, visor, VOX, squelch, TOT, DTMF, teclas…) para 428 modelos, com a descrição de cada item e um aviso a confirmar para os delicados.
- **Identificação do rádio ligado**: o programa reconhece-o sozinho e mostra marca, modelo, tipo, firmware e as outras informações que comunica (para os rádios DMR USB também ID DMR e nome). Com a **identificação automática** (botão no separador «Identificar», escolha lembrada) basta ligar o cabo, em qualquer separador: o programa escuta todas as portas dos cabos ao mesmo tempo, reconhece também um rádio ligado depois de ligar o cabo e vários rádios ligados a cabos diferentes, experimenta primeiro os rádios já reconhecidos e os mais difundidos, não se deixa enganar pelos rádios em modo de atualização do firmware e abre o separador de cada um com os canais lidos; o quadro «A identificar o rádio» só aparece para o cabo com um rádio que responde; para outro rádio há o botão **«🔎 Identificar rádio»** em cima à direita.
- **Envio para outro rádio**: os canais marcados num rádio são escritos diretamente noutro rádio ligado, escolhido por marca e modelo, com verificação de compatibilidade, adaptação, cópia de segurança e verificação.
- **Conversor universal**: os canais de qualquer rádio ou ficheiro (K67 `.xlc`, Bajeton `.dat`, imagens `.img` também do CHIRP, CSV do CHIRP, Excel/CSV) no formato de qualquer outro rádio, adaptados às suas regras.
- **Interface em 13 línguas** com o seletor de bandeiras ao lado de «ℹ Info»: tradução imediata de todo o programa (árabe da direita para a esquerda), língua lembrada ao reiniciar; ficheiros informativos traduzidos na pasta `lingue`.
- **Nomes das zonas** em vez de «Zona 1», «Zona 2»… (Radtel RT-950 Pro, RT-950, RT-920, RT-910, Bajeton BJ-9000 e BJ-7700, Hiroyasu HI-UV850 Pro; zonas dos rádios DMR RD-5R, DM-1801 e DM-32UV), lidos e escritos por cabo juntamente com os canais.
- **Edição em bloco** dos canais marcados: potência, largura de banda, varredura, subtons, desvio ou transmissão em conjunto, o resto de cada canal fica como está.
- **Galeria de 564 modelos** de 47 marcas, divididos entre compatíveis e não compatíveis, por tipo (radioamador, DMR, GMRS/FRS, PMR446, para crianças) e por marca, com pesquisa, filtros e ligações aos programas e manuais oficiais.
- **Arquivo de repetidores** baseado na lista atualizada de [IK2ANE](https://www.ik2ane.it), com filtros e inserção nos rádios, com a verificação de compatibilidade entre repetidores e rádio antes da escrita.
- **Atualização do firmware** de 57 modelos (também os Quansheng UV-K5, UV-K6 e UV-5R Plus, os Radtel RT-590 e RT-600 com os ficheiros .bin da Quansheng e da Radtel, os Radtel RT-470, RT-470L, RT-470X, RT-495 e RT-630 com os ficheiros .kdhx da Radtel, os TIDRADIO TD-H3, TD-H8 e Radtel RT-730 com os ficheiros .BIN do fabricante, os Retevis RA79 V3 e RA89 com os ficheiros .bin e o Retevis RT85 com os ficheiros .hex da Retevis, os Radtel RT-950 Pro e RT-950 e o Bajeton BJ-9000 com os ficheiros .BTF, os Radtel RT-920 e RT-910B com os ficheiros .anc, os Radtel RT-910 e RT-900 e o Bajeton BJ-7700 sem Bluetooth com os ficheiros .BTU, o Radtel RT-890 com os ficheiros .bin da Radtel e da M7OCM, o Baofeng DM-1701 por USB com os ficheiros .bin da Baofeng, os Baofeng RD-5R e DM-1801 por USB com os ficheiros .sgl), com verificação do ficheiro e cópia de segurança do rádio; atualização guiada com o programa oficial para BTECH DMR-6X2 e DMR-6X2 PRO.
- **Cópia de segurança** automática antes de cada escrita e verificação depois.
- **Verificação de atualizações** do programa a cada arranque.

## Utilização

1. Transfira [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip) (última versão) e extraia o arquivo.
2. Inicie `IU0VYT Flash.exe` (Windows 10/11; usa o componente WebView2 do Windows).
3. No separador **📻 Modelos** escolha o seu rádio, ligue o cabo e use **Ler do rádio**.

O guia completo está no ficheiro `LEIAME.txt` (original italiano: `LEGGIMI.txt`).

## Compilação

É preciso o Go 1.22 ou posterior. Na pasta `sorgenti`:

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

As bibliotecas estão incluídas na pasta `vendor`.

## Avisos

- Os protocolos dos rádios foram reconstruídos para a interoperabilidade e verificados com rádios simulados: **guarde sempre as cópias de segurança**. O programa é fornecido sem qualquer garantia.
- Programa independente, **não afiliado nem aprovado** pelos fabricantes dos rádios (Baofeng, Bajeton, Quansheng, Retevis, Radtel, Radioddity, TIDRADIO e os outros). As marcas citadas pertencem aos respetivos titulares e servem só para indicar os rádios compatíveis.
- Respeite as normas sobre o uso das frequências do seu país.

## Agradecimentos

- [CHIRP](https://chirpmyradio.com) e os seus autores: grande parte dos controladores dos rádios deriva do seu trabalho.
- [IK2ANE](https://www.ik2ane.it), que põe gratuitamente à disposição de todos a lista dos repetidores italianos.

Autores e licenças dos componentes: `AUTORES_E_LICENCAS.txt`.

## Licença

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

Software livre distribuído com a **GNU General Public License versão 3 ou posterior** (`GPL-3.0-or-later`): ver `LICENSE.txt` e `AVISO_DE_LICENCA.txt`.
