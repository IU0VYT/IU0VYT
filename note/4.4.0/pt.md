# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 para Microsoft Windows 10/11

Esta versão inclui também as novidades da 4.3.0, 4.2.0, 4.1.0 e 4.0.0, que não tiveram uma versão estável própria.

## Novidades da 4.4.0

### Corrigido: quadro «Identificação do rádio em curso» sem rádio
- O quadro podia aparecer também com um cabo sem rádio: o ruído da linha e o eco dos comandos enviados pareciam um rádio a responder.
- Agora aparece **só para o cabo ligado a um rádio que responde de facto**. Com dois cabos e um só rádio (COM6 sem rádio, COM7 com o rádio, ou o contrário) aparece só para a porta do rádio.
- Sem rádio, a ronda de tentativas também é mais rápida.

### Barra superior, Início, separador Identificar
- O botão chama-se **«🔎 Identificar rádio»** e está à direita, ao lado de «Atualizações».
- Com muitos separadores abertos, o botão encolhe sozinho, até ficar só a lupa.
- Início: voltam as descrições dos quadros da 4.2. Em baixo continuam visíveis copyright, site e código-fonte, também a 1280×720.
- Separador Identificar: explicações curtas e claras.

### Firmware Retevis (43 modelos com atualização do firmware)
- **Retevis RA79 V3** (e UV-K5 V3 / UV-K1, carregador 7.x) com os ficheiros .bin da Retevis, por exemplo «RA79 V3.bin». O firmware só é enviado se o carregador do rádio for da mesma família.
- **Retevis RA89** com os ficheiros .bin da Retevis, por exemplo «Eng_RA89_Upadat_20250524_B91.38.bin». Usa o protocolo do programa da Retevis «RA89-update-software», que foi executado com um rádio simulado. Os comandos enviados são idênticos byte a byte.

### Outros
- Os testes com rádios simulados já não impedem a compilação dos testes no Windows.

Galeria: 386 modelos de 27 marcas (348 com canais, 302 com definições, 43 com atualização do firmware).

## Novidades da 4.3.0

### Identificação automática: mais rápida, mais rádios, mais portas
- Uma ronda completa de tentativas sem rádio demora cerca de **50 segundos** (antes 135).
- O programa **escuta continuamente todas as portas dos cabos**, cada uma por si e todas em conjunto:
  - reconhece também um rádio ligado desligado e ligado mais tarde;
  - reconhece **vários rádios em cabos diferentes** ao mesmo tempo, mostra os seus dados e abre todos os seus separadores.
- Os **rádios já reconhecidos** (p. ex. K67 Pro, Bajeton 8300, BJ-9000) são tentados primeiro, depois todos os outros.
- Se o separador do rádio já estiver aberto aparece o aviso «Rádio já identificado!! Separador do rádio já aberto» (com vários rádios «Rádios já identificados!! Separadores dos rádios já abertos»).
- Separador Identificar: portas por ordem (COM6, COM7… COM10) e escritas como «COM6 — USB-SERIAL CH340»; «Identificar novo rádio automaticamente» a negrito.

### Atualização do firmware: mais 8 rádios (41 no total)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** com os ficheiros .kdhx da Radtel (carregador «JJCC Bootloader Pro»). RT-470: o firmware 1.x é só para a placa antiga, o 2.x só para a nova. O programa indica-o e pede confirmação escrita.
- **TIDRADIO TD-H3, TD-H8 e Radtel RT-730** com os ficheiros .BIN do fabricante: inicie a atualização e ligue o rádio mantendo premido o PTT.

### Outros
- **Bajeton BJ-9000**: como o BJ-8300, todos os 960 canais numa lista única sem zonas («Canais: todos (960)»).
- **Início** mais compacto: em baixo veem-se sempre versão, copyright, site e código-fonte.
- **Galeria**: uma linha grossa separa os modelos compatíveis dos não compatíveis.
- **LEGGIMI**: removida uma parte longa repetida por engano na 4.2.0.

Galeria: 386 modelos de 27 marcas (348 com canais, 302 com definições, 41 com atualização do firmware).

## Novidades da 4.2.0

- Identificação automática corrigida: o quadro «Identificação do rádio em curso» só aparece quando responde um rádio ligado (já não no arranque só com o cabo) e desaparece quando o rádio é reconhecido ou não é encontrado. Após o reconhecimento abrem-se os dados e o separador do rádio e a pesquisa para, sem desativar a função.
- Novo botão «🔎 Identificar novo rádio automaticamente» na barra superior, entre «Info» e «Atualizações»: procura um novo rádio com as mesmas regras e fica cinzento se a identificação automática estiver desativada.
- 35 rádios novos (36 versões) e mais 8 marcas: família TYT TH-UV88 (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), família Radtel RT-490 (com JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 rádios da família Retevis RT21 (com Abbree AR-63) e Quansheng TK11.
- Atualização do firmware também para Radtel RT-590 / RT-590 Plus e RT-600 / RT-600 Plus com os ficheiros .bin da Radtel.
- Corrigidos alguns defeitos ocultos na leitura dos nomes dos canais com letras acentuadas ou chinesas.
- Galeria de 386 modelos de 27 marcas: 348 com canais e repetidores, 302 com definições, 33 com atualização do firmware.

## Novidades da 4.1.0

- Identificação automática corrigida: funciona em qualquer separador e, quando desliga o cabo do rádio, o desliga ou o troca, a pesquisa recomeça sozinha e reconhece o rádio ligado a seguir (mesmo que seja o mesmo ou que o ligue depois do cabo). Antes, após o primeiro reconhecimento, já não procurava outros rádios.
- Durante a pesquisa automática aparece o quadro «Identificação do rádio em curso. Aguarde, por favor.», com o botão «Interromper».
- 14 rádios novos e mais 3 marcas: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 e RT-730, Hiroyasu HI-8811, Talkpod A36plus e A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (2.ª e 3.ª geração) e TD-H3, nas versões base, HAM e GMRS.
- Atualização do firmware dos Quansheng UV-K5, UV-K5(8), UV-K6 e UV-5R Plus com os ficheiros .bin da Quansheng, verificando que o firmware é para o tipo de rádio ligado.
- Ecrã inicial e site: «GESTÃO DE CANAIS - FREQUÊNCIAS - REPETIDORES - FIRMWARE» e os nomes de todas as marcas numa moldura ao estilo dos videojogos dos anos 80.
- Galeria de 349 modelos de 19 marcas: 311 com canais e repetidores, 265 com definições, 31 com atualização do firmware.

## Novidades da 4.0.0

- Programa multimarca: 335 modelos de 16 marcas na galeria, 297 com canais e frequências por cabo e 251 com as definições do rádio.
- Rádios novos: Quansheng UV-K5, UV-K6, UV-5R Plus e UV-K1; 31 modelos Retevis (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920 e T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S e GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton e TID.
- Identificação automática (separador «Identificar»): liga o cabo e o rádio é reconhecido sozinho, com o seu separador aberto e os canais lidos. A escolha fica memorizada e desativa-se com o mesmo botão.
- Galeria dos modelos com o menu das marcas e uma cor para cada marca.
- «Ler do rádio» e «Escrever no rádio» a azul nos separadores dos rádios.
- Início, Info, guia, site e documentação atualizados para o programa multimarca.
