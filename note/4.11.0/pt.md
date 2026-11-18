# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="de.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="es.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub><b>Português</b></sub></td>
<td align="center"><a href="ru.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="ms.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

Página da versão no GitHub (em inglês), com a transferência: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 para Microsoft Windows 10/11

### Nomes das zonas
- Em vez de «Zona 1», «Zona 2»… cada zona pode ter um nome, como nos programas do fabricante: botão «✎ Nomes das zonas…» ou duplo clique na zona. A barra das zonas e a linha inferior mostram «número · nome».
- Rádios: Radtel RT-950 Pro, Bajeton BJ-9000 e Hiroyasu HI-UV850 Pro (firmware até 0.23 e a partir da 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 sem Bluetooth; zonas dos rádios DMR Baofeng RD-5R, DM-1801 e DM-32UV.
- Os nomes leem-se e escrevem-se por cabo juntamente com os canais e guardam-se no ficheiro .img. Comprimento e caracteres como nos programas do fabricante (12 caracteres, também chineses, no RT-950 Pro; 10 no RT-920; 16 nos DMR).
- Segurança: a área dos nomes só é escrita se os nomes mudarem, depois de a reler do rádio; se o rádio tiver ali dados que não são nomes, nada é escrito.

### Edição em bloco
- «✎ Edição em bloco…» (também com o botão direito) altera em conjunto potência, largura de banda, varredura, tom TX, tom RX, desvio e transmissão dos canais marcados (ou selecionados). As opções deixadas em «não alterar» ficam como estão; Ctrl+Z anula tudo.

### Mais duas línguas (11 no total)
- Coreano, com a bandeira da Coreia do Sul.
- Árabe, com a bandeira dos Emirados Árabes Unidos e a página da direita para a esquerda (frequências e números continuam da esquerda para a direita).
- Todos os ficheiros informativos da pasta `lingue` estão traduzidos também em coreano e árabe e atualizados para a 4.11 em todas as línguas.

### Outros
- Site: indicadas as línguas disponíveis; removida a lista dos ficheiros de firmware.
- `go test` passa no Windows (também i18n_test.go); um novo teste verifica que LEGGIMI.txt, AUTORI_E_LICENZE.txt e a licença são idênticos na pasta principal e em «sorgenti».

Os novos formatos foram obtidos do código dos programas do fabricante e do controlador do CHIRP e testados com rádios simulados; não foram testados com rádios reais: guarde sempre as cópias de segurança.

Galeria: 489 modelos de 44 marcas (451 com canais, 396 com definições, 57 com atualização do firmware).
