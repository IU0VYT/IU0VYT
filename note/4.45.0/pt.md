# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0 para Microsoft Windows 10/11

### Novidades
- **Rádios DMR da família AnyTone**: AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus e BTECH DMR-6X2 e DMR-6X2 PRO (estes dois eram «não compatíveis»). Canais e zonas com o cabo USB do rádio; memória e protocolo do programa livre dmrconfig, verificados com o seu código original. Só se escrevem canais, zonas e listas de varredura.
- **TYT MD-UV380, MD-UV390, MD-2017, MD-9600 e Retevis RT3S e RT82** com o firmware original: mesma memória e mesmo protocolo do Baofeng DM-1701.
- **Baofeng BF-666S** (controlador do BF-888S) e **BF-T6** (experimental, controlador do BF-T8).
- Para as fichas que continuam «não compatíveis» UV-8D, BF-UV8, DM-1702 e DM-1703 agora está escrito o motivo.
- Galeria: 673 modelos de 54 marcas (623 com canais, 556 com definições).

### Notas
- Nenhum dos rádios novos foi testado num rádio real: faça primeiro apenas uma leitura.
- BTECH DMR-6X2 PRO e Baofeng BF-T6 são experimentais.
- Programa independente, não afiliado aos fabricantes dos rádios.
- Bateria completa de testes superada: parte geral; 918 comparações de estrutura e 381 leituras / escritas idênticas ao CHIRP; família AnyTone comparada com o código original do dmrconfig.

---

Esta versão inclui também as novidades das versões 4.24.0 a 4.44.0, que não tiveram uma versão estável própria. Aqui estão, da mais recente.

## IU0VYT Flash 4.44.0

### Novidades
- **Definições** para mais oito rádios que só tinham os canais: Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro e Bajeton DM-5000, a partir dos programas dos fabricantes.
- **RT-480 e RT-68**: na escrita, agora chega ao rádio também o bloco das definições gerais.
- **RT-4D, RT-6D, RT-1000 Pro**: o ficheiro do rádio contém também as definições (os ficheiros guardados antes continuam válidos).

### Notas
- Nenhuma das definições novas foi testada num rádio real.
- RT-69 GPS: por agora só as opções verificadas uma a uma no programa do fabricante.

## IU0VYT Flash 4.43.0

### Novidades
- **Erro corrigido**: em 33 rádios (entre os quais Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22 e HA1G, Kenwood TK-2140 / TK-3140) e nos 8 «gémeos» da 4.42, as definições não apareciam depois da leitura por cabo, mas só ao abrir um ficheiro do CHIRP.
- **Definições** para mais cinco Radtel: RT-817A / RT-871A, RT-830, RT-850, RT-330 e RT-400, a partir dos programas do fabricante.

### Notas
- Nenhuma das definições novas foi testada num rádio real.
- Radtel RT-830 e RT-850: uma opção por rádio é lida pelo programa do fabricante de forma diferente de como é escrita; aqui segue-se a escrita e a opção tem um aviso.

## IU0VYT Flash 4.42.0

### Novidades
- **Separador Início**: agora cabe inteiro na janela. O quadro das marcas é mais largo e a página adapta-se sozinha ao tamanho da janela.
- **Definições** para 30 modelos que só tinham os canais: Baofeng série D (BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro e Radtel RT-13.

### Notas
- Nenhuma das definições novas foi testada num rádio real.
- Baofeng série D: no programa do fabricante «guia de voz» e «som das teclas» estão trocados entre leitura e escrita; aqui segue-se a escrita e as duas opções têm um aviso.

## IU0VYT Flash 4.41.0

### Novidades
- **Rádios novos** (os portáteis do CHIRP que faltavam): Yaesu VX-1R (canais e definições), Icom IC-91AD e IC-92AD (duas bandas, também D-STAR), BTECH UV-PRO (por Bluetooth, sem cabo) e o recetor AOR AR8200 (marca nova).
- **Definições** para rádios que só tinham os canais: Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### Notas
- Nenhum dos rádios novos foi testado na prática: faça primeiro apenas uma leitura.
- BTECH UV-PRO: a escrita só é permitida com a versão do firmware testada pelo autor do controlador do CHIRP (147).

## IU0VYT Flash 4.40.0

### Novidades
- **Pesquisa automática mais rápida com o separador já aberto**: a verificação curta (antes só para o Baofeng K67 Pro) vale agora também para os Icom, os Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, os Quansheng UV-K5 e derivados, os Bajeton BJ-9000, BJ-7800, BJ-7700 e BJ-8300, os Radtel RT-900 / RT-950 / RT-470 e os rádios com a mesma memória. Para os outros mantém-se a leitura completa.
- **Marca nova: Alinco.** Portáteis DJ-596, DJ-175, DJ-G7EG e DJ-G7T; móveis DR-135T, DR-235T, DR-435T, DR-03T e DR-06T. Só canais. Funciona também com os cabos que não devolvem os dados.

### Notas
- Nenhum Alinco foi testado na prática: faça primeiro apenas uma leitura.
- A verificação curta foi testada com os rádios simulados; se depois da verificação um rádio não voltar ao funcionamento normal, comunique-o.

## IU0VYT Flash 4.39.0

### Novidades
- **Quatro rádios Icom novos**: ID-80H (portátil D-STAR), ID-880H (móvel D-STAR), IC-W32A e IC-W32E (portáteis bibanda com duas metades separadas: de 0 a 99 os canais VHF, de 100 a 199 os canais UHF). Só canais.

### Notas
- Nenhum dos rádios novos foi testado na prática: faça primeiro apenas uma leitura.

## IU0VYT Flash 4.38.0

### Correções
- **Identificação**: o Bajeton BJ-9000 às vezes bloqueava, sobretudo na primeira identificação. O programa voltou a escutar a porta durante um segundo antes da primeira saudação, como na 4.35, também para o rádio já reconhecido nesse cabo. As outras melhorias da 4.36 mantêm-se.

### Notas
- Nenhum rádio novo nesta versão.

## IU0VYT Flash 4.37.0

### Novidades
- **Oito portáteis Yaesu novos**: VX-5R e VX-7R (canais); VX-2R, FT-50R, VX-150, VX-8R, VX-8DR e VX-8GE (canais e definições; nos VX-8 também APRS).
- A cópia começa no rádio (modo «CLONE»): o separador escolhe-se à mão e as instruções estão na janela do cabo.

### Notas
- Nenhum dos rádios novos foi testado na prática: faça primeiro apenas uma leitura.
- Não feito: Yaesu VX-1.

## IU0VYT Flash 4.36.0

### Correções
- **«Identificar» e identificação automática**: o programa lembra-se em que cabo reconheceu cada rádio e começa por esses; a primeira pesquisa já não perde segundos a escutar e a preparar; o rádio ligado depois do cabo é tentado de novo com mais frequência.
- O quadro «Identificação do rádio em curso» só aparece enquanto o rádio se deixa ler, com o modelo e o progresso. Se o rádio estiver ligado ao cabo mas desligado, o programa avisa para o ligar e reconhece-o logo que está ligado.
- Baofeng K67 Pro com o separador já aberto: reconhecido em poucos segundos, sem voltar a ler toda a memória.

### Novidades
- Separador «Identificar»: botão **«Guardar o diagnóstico da identificação»** (tempos e primeiros bytes das últimas pesquisas num ficheiro de texto, para enviar ao autor se a identificação for lenta ou se enganar; não contém canais nem dados pessoais).

### Notas
- Nenhum rádio novo nesta versão.

## IU0VYT Flash 4.35.0

### Correções
- **«Identificar» e identificação automática**: removida a «ronda rápida» da 4.31, que com rádios reais podia deixar o rádio bloqueado em programação e tornar a pesquisa lenta. As esperas voltaram a ser as de antes.
- A pesquisa tenta os mesmos grupos de rádios em qualquer PC (antes mudavam de um arranque para o outro).
- O eco de um cabo sem rádio já não é confundido com um rádio.

### Notas
- Nenhum rádio novo nesta versão.

## IU0VYT Flash 4.34.0

### Novidades
- **Rádios novos**: Yaesu VX-170 e VX-177 (portáteis), FT-7800 / FT-7900 e FT-8900 (móveis), só canais, com os controladores do CHIRP: leitura, escrita e bytes trocados com o rádio idênticos ao CHIRP.

### Notas
- A cópia começa no rádio (modo «CLONE»): o separador tem de ser escolhido à mão, «Identificar» não tenta estes rádios.

## IU0VYT Flash 4.33.0

### Novidades
- **Rádios novos**: Icom IC-T7H, IC-T8A, IC-Q7A e IC-P7 (canais), com os controladores do CHIRP: leitura, escrita e bytes trocados com o rádio idênticos ao CHIRP.

### Notas
- IC-Q7A e IC-P7: o modo (FM, FM larga, AM) segue a frequência; o salto «P» e o passo de sintonia do canal são conservados.

## IU0VYT Flash 4.32.0

### Novidades
- **Rádios novos**: Baofeng T22 na versão FRS, TIDRADIO TD-M8S, TD-V2, TD-H4 e TD-H11 (canais e definições), a partir dos programas dos fabricantes.
- **«Identificar»** propõe também os rádios novos que se apresentam como o BF-888S, sem alongar a pesquisa: o rádio só é lido com o controlador deles quando se abre esse separador.

### Correções
- TIDRADIO de 2026 (TD-H3 Plus / SE, TD-H7 Plus, TD-H8 4.ª geração, TD-H9, TD-638UV): a potência média só é oferecida pelos modelos que a têm de facto, e nunca nas versões GMRS.
- Baofeng T22 PMR446: ligação como no programa do fabricante.
- Notas dos Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS e TIDRADIO TD-M2 Pro: «Identificar» tenta-os; a frase que dizia o contrário foi removida.

## IU0VYT Flash 4.31.0

### Novidades
- **«Identificar» é muito mais rápido**: primeiro faz uma ronda rápida de todos os rádios e só depois, se não tiver encontrado nada, a completa de antes. Um rádio no fim da lista é encontrado em menos de meio minuto em vez de quase dois.
- **«Identificar» tenta agora também os rádios acrescentados nas últimas versões**: Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G e os TIDRADIO sem ecrã.

### Correções
- Retevis RB39 / RB639: com o separador do país errado o rádio ficava em programação até ser desligado.

## IU0VYT Flash 4.30.0

### Novidades
- **Definições gerais** para os TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 de quarta geração**, **TD-H9** e **TD-638UV**: teclas programáveis, squelch, VOX, tempo máximo de transmissão, ecrã, língua, mensagens de arranque e mais.
- **Rádios novos (canais e definições)**: TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR** e **TD-M2 Pro**.

### Notas
- Os rádios novos vêm dos programas do fabricante e ainda não foram testados com um rádio real: faça primeiro apenas uma leitura. «Identificar» não os distingue entre si: escolha o separador você mesmo.

## IU0VYT Flash 4.29.0

### Novidades
- **Rádios novos (só canais)**: TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 de quarta geração**, **TD-H9** e **TD-638UV**, cada um nas versões base, HAM e GMRS.
- O TD-H3 Plus apresenta-se como o TD-H3: use os separadores «TD-H3 Plus», porque tem três potências.

### Notas
- Estes rádios vêm do programa do fabricante de 2026 («TIDRadioCPS») e ainda não foram testados com um rádio real: faça primeiro apenas uma leitura. As definições gerais ainda não são geridas.

## IU0VYT Flash 4.28.0

### Novidades
- **Rádios novos** (móveis Radioddity): **UB15-G** (GMRS 15 W), **DB35-G** (GMRS 35 W), **DB35-UV** (VHF / UHF 35 W) e **DB40-G** (GMRS 40 W, 199 canais), com canais e definições gerais. O UB15-G tem dois separadores, consoante o firmware: antes da V1.10 (sem nomes) e V1.10 (com os nomes).
- O **DB40-G** também está no CHIRP: aqui lê e escreve os mesmos bytes e os mesmos ficheiros `.img`.

### Correções
- Os testes não compilavam no Windows desde a 4.26.0 (três ficheiros de teste usavam os rádios simulados, que no Windows não são compilados). O programa não era afetado.

### Notas
- UB15-G, DB35-G e DB35-UV vêm do programa do fabricante e do seu ficheiro de configuração e ainda não foram testados com um rádio real: faça primeiro apenas uma leitura. «Identificar» não reconhece estes rádios: o separador tem de ser escolhido à mão.

## IU0VYT Flash 4.27.0

### Novidades
- **Rádios novos**: **Retevis C6** (FRS / GMRS, 30 canais), **Retevis C66** (UHF, 16 canais) e **Radioddity Echo-DE / Echo-IT** (PMR446, 16 canais), com canais e definições gerais.
- Antes de escrever, o programa verifica que a banda, o modelo e a região do rádio são os da memória a escrever; palavra-passe e calibrações do rádio nunca são escritas.

### Correções
- **Retevis RB39 / RB39P**: agora está entre os rádios «GMRS / FRS (EUA)» no separador Modelos.

### Notas
- Os rádios novos vêm do código dos programas dos fabricantes e ainda não foram testados com um rádio real: faça primeiro apenas uma leitura. «Identificar» não os reconhece: o separador tem de ser escolhido à mão.
- Se o rádio tiver uma palavra-passe de leitura, retire-a antes com o programa do fabricante.

## IU0VYT Flash 4.26.0

### Novidades
- **Rádios novos**: **Retevis RB39 / RB39P** (FRS, 22 canais) e **RB639 / RB639P** (PMR446, 16 canais), com canais e definições gerais. O rádio declara para que país é feito: com o separador errado a leitura para e indica o certo.
- **Definições novas** para **Baofeng BF-T17 / BF-T18** (PMR446 e FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3** e **PR-T3**: squelch, tempo máximo de transmissão, VOX, guia de voz, som das teclas, poupança de bateria e as outras do programa do fabricante.

### Correções
- Registos de alterações nas 12 línguas: data da 4.25.0 corrigida (12/11/2026).

### Notas
- Os rádios novos vêm do código dos programas dos fabricantes e ainda não foram testados com um rádio real: faça primeiro apenas uma leitura. «Identificar» não os reconhece: o separador tem de ser escolhido à mão.

## IU0VYT Flash 4.25.0

### Novidades
- **Bajeton BJ-8300 com o firmware novo**: separador novo para os rádios com 990 canais em 10 zonas de 99 (programa do fabricante BT-11_CPS v1.2.6): canais, nomes das zonas e definições. O separador «Bajeton 8300» de sempre (960 canais, 15 zonas de 64) fica como está e agora **avisa** quando encontra o firmware novo.
- **Rádios novos**: **Radioddity GS-10B Pro** (999 canais e definições), **Radioddity FS-T3** e **PR-T3** (canais).
- **Bajeton BJ-9000** com 990 canais: mais cinco definições (as do programa v1.3.3).

### Correções
- **Baofeng BF-T17 / BF-T18, GT-18 e Radioddity PR-T5**: a potência alta e a baixa estavam trocadas. Quem escreveu estes rádios com uma versão anterior deve relê-los e escolher de novo a potência.
- «Saída automática do menu» (BJ-8300, BJ-9000, Radtel RT-950): a última opção é 60 segundos, não 55.

### Notas
- Os rádios e o separador novos vêm do código dos programas dos fabricantes e ainda não foram testados com um rádio real: faça primeiro apenas uma leitura.
- BJ-8300: para saber que separador usar, veja no rádio quantas zonas há (15 de 64 canais, ou 10 de 99).

## IU0VYT Flash 4.24.0

### Novidades
- **Separador Modelos mais legível**: as etiquetas dos tipos de rádio estão no centro da página e a linha entre rádios compatíveis e não compatíveis é vermelha e mais grossa.
- **Mais 26 rádios**, geridos com os controladores de rádios iguais já presentes:
  - reconhecidos comparando os programas dos fabricantes: **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28**;
  - da tabela de modelos equivalentes do CHIRP: Radioddity GA-5S, Retevis RT5 (três potências), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+ e TS-T9+, Rugged RH5R e RH5X, Ansoko A-5R e A-8S, Arcshell AR-5, AR-6 e AR-8, Greaval GV-8S e GV-9S, Tenway TW-325, UV-82 Pro e UV-5R Pro.

### Notas
- Os seis rádios reconhecidos comparando os programas dos fabricantes ainda não foram testados com um rádio real: faça primeiro apenas uma leitura. Se o modelo não for o certo, o rádio não responde e não é alterado.
