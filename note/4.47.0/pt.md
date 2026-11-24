# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 para Microsoft Windows 10/11

### Novidades
- **Retevis, 24 modelos novos obtidos a partir dos programas do fabricante**: B3B, B63B, RT22P, RT622P, RB666, RB666P (mesma memória de canais da nova versão do H777); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (família do T18); RB65, RB66, RB66P (uma frequência e um tom por canal). A ligação, a leitura e a escrita foram reescritas de raiz e verificadas com os ficheiros de dados de fábrica dos programas. Os RB66 e RB66P, e os RB669, RT68H e RT668H, respondem com o mesmo código: a ficha certa tem de ser escolhida à mão.
- **Definições dos 24 modelos Retevis**: 241 itens ao todo (squelch, tempo máximo de transmissão, VOX, pesquisa, guia de voz, poupança de bateria, bip, teclas laterais…). Só são oferecidos os itens que o programa do modelo mostra; os valores dos 24 ficheiros de fábrica são todos opções admitidas.
- **Palavra-passe de programação**: nos 24 modelos Retevis e na nova versão do Retevis H777 não impede a leitura nem a escrita, como no CHIRP (o programa do fabricante, pelo contrário, para se a palavra-passe não for a escrita pelo utilizador). É um item das definições, em claro: pode ser vista, alterada e retirada. O comprimento e os caracteres são os do programa do fabricante: 6 caracteres (3 nos RB45, RB645, RB47 e RB66; 4 no RB66P), algarismos e letras (só algarismos nos RB65 e RB66). Nos B3H, B63H, RB45, RB645, RB66 e RB666, cujo programa não mostra a caixa da palavra-passe, um aviso lembra-o.
- **Definições para rádios que só tinham os canais**: família AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 e DMR-6X2 PRO: 27 itens, a partir das descrições da memória do qdmr e do dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 e MD-9600, Retevis RT3S e RT82 (os itens do Baofeng DM-1701); Radtel RT-880 e RT-950 Pro com o firmware nicFW880 / nicFW950 (17 itens, a partir da documentação do autor); TID TD-UV68 (os itens do TD-H8); Radtel RT-890 com o firmware M7OCM (53 itens).
- **Versões de teste**: uma cópia com um número mais alto do que a última versão estável publicada mostra uma barra de aviso («Versão de teste (beta testing)…») e, só ela, propõe a versão de teste seguinte, sempre descarregada do GitHub. Quem usa a versão estável não vê as versões de teste.
- Galeria: 697 modelos de 54 marcas (647 com canais, 596 com definições).

### Notas
- Nenhum dos 24 rádios Retevis foi testado em equipamento real, nem os canais nem as definições: primeiro só a leitura, depois uma única definição inofensiva. Se o rádio responder de forma inesperada, a leitura para com uma mensagem e nada é escrito no rádio.
- Com uma palavra-passe de programação já presente o programa tenta ler e escrever na mesma, como o CHIRP: não foi testado num rádio real.
- Os RB66 e RB66P, e os RB669, RT68H e RT668H, respondem com o mesmo código: a ficha certa tem de ser escolhida à mão.
- Programa independente, não afiliado aos fabricantes dos rádios.
- Bateria completa de testes superada: parte geral; 915 comparações de estrutura e 380 leituras / escritas idênticas ao CHIRP.
