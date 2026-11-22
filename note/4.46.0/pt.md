# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 para Microsoft Windows 10/11

### Correções
- **Radtel RT-880G**: já não é indicado como rádio GMRS / FRS. É um transceptor de radioamador VHF/UHF (com GPS e APRS), como o RT-880 e o RT-860; o aviso «Rádio vendido como GMRS» já não aparece antes da escrita.
- **Verificação após a escrita** no Radtel RT-880G: falhava porque o rádio, que no fim da escrita guarda os dados e reinicia, era contactado cedo demais e com demasiada frequência. Agora o programa espera 10 segundos sem o incomodar e tenta de novo a cada 15 segundos, até 90 segundos. Os mesmos tempos valem para o RT-880, o RT-860 e o RT-817A / RT-871A, que usam o mesmo controlador.
- Controlos mais rigorosos dos valores recebidos do AR8200, dos Kenwood e do TK-180 (avisos do CodeQL).

### Notas
- A verificação foi testada num RT-880G real; o RT-880, o RT-860 e o RT-817A / RT-871A não foram testados com rádios reais. Os outros rádios mantêm os tempos de antes.
- Nenhum rádio novo nesta versão.
- Programa independente, não afiliado aos fabricantes dos rádios.
- Bateria completa de testes superada: parte geral; 915 comparações de estrutura e 380 leituras / escritas idênticas ao CHIRP.
