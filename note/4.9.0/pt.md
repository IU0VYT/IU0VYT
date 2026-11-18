# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 para Microsoft Windows 10/11

Esta versão inclui também as novidades da 4.8.0, da 4.7.0 e da 4.6.0, publicadas só como versões de teste.

## Novidades da 4.9.0

### Corrigido: Identificar com Lanchonlh HG-UV98 e Quansheng TG-UV2+
- Durante «Identificar» a porta falava sempre em 8N1. Agora usa paridade par (HG-UV98) e dois bits de paragem (TG-UV2+), como a leitura a partir do separador do rádio. Os dois rádios passam a ser reconhecidos.

### Corrigido: paridade do HG-UV98
- Com paridade par, a porta do Windows agora também verifica os erros de paridade. Antes, a verificação era ligada e logo desligada.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus com firmware livres
- Novos separadores para os firmware egzumer, F4HWN e LOSEHU/LTS/UVE.
- O F4HWN usa o controlador F4HWN de armel (fora do CHIRP): 8 níveis de potência, bloqueio da transmissão, 3 listas de varredura e as definições do firmware.
- Cada separador lê e escreve só o seu firmware. Depois da atualização do firmware, o programa indica o separador a usar.

### Rádios novos
- Wouxun (20 modelos): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (EUA)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (firmware Radtel e firmware livre M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

Para os rádios dos controladores do CHIRP, canais e definições são verificados com os testes automáticos contra o CHIRP. Os rádios GMRS, FRS e MURS são para os Estados Unidos: em Itália não é permitido transmitir nessas frequências.

## Novidades da 4.8.0

### Atualização do firmware .BTU (Radtel RT-910 e RT-900 sem Bluetooth, Bajeton BJ-7700 sem Bluetooth)
- Os firmware .BTU atualizam-se com o protocolo do programa «Radtel Bootload Tools». O carregador é o do Baofeng UV-5RM, com a apresentação «PROGRAMBT80U».
- Modo automático, com o rádio ligado normalmente. Modo manual: rádio desligado, ligado mantendo premidas as duas teclas laterais, até o ecrã mostrar «UPDATE».
- O carregador não verifica o modelo. Por isso o programa reconhece o rádio e a versão da eletrónica pelo nome do ficheiro, mostra os avisos e pede sempre para escrever o modelo.
- Na galeria, o RT-910 (sem Bluetooth, .BTU) e o RT-910B (com Bluetooth, .anc) têm agora duas fichas. O mesmo vale para o BJ-7700 e o BJ-7700 sem Bluetooth.

### Corrigido: Radtel RT-900 sem Bluetooth com o firmware de 512 canais
- Até à V1.09P o rádio tem 512 canais; a partir da V1.14P tem 999.
- Agora os canais 513-999 só são escritos se os modificar. Antes, o programa verifica no rádio que essa área está vazia ou contém canais.

### Rádios novos (controladores do CHIRP)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (a nova versão do H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

Para todos, canais e definições são verificados com os testes automáticos contra o CHIRP.

## Novidades da 4.7.0

### Corrigido: identificação automática bloqueada ou muito lenta
- Um rádio ligado em modo de atualização do firmware transmite sozinho mensagens periódicas: antes parecia responder a todos os modelos e a pesquisa durava minutos. Agora o programa escuta a porta antes de tentar e não se deixa enganar: a pesquisa termina em menos de um minuto e o separador Identificar explica o que se passa.
- A interface já não bloqueia enquanto as pesquisas arrancam; os rádios mais comuns são tentados primeiro.
- Corrigido o falso aviso «Separador do rádio já aberto» para um separador fechado antes.

### Rádios com a mesma gestão
- A janela de escolha e o separador Identificar listam todos os rádios possíveis (p. ex. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), com o menu para escolher e uma breve indicação de onde ver a versão do firmware no rádio.

### Corrigido: Radtel RT-950 Pro / Bajeton BJ-9000 com o firmware até 0.23
- O bit «transmissão ativada» dos canais agora é sempre escrito, com as duas versões (como os programas da Radtel até à v1.2.x; a v1.3.3 ignora-o). Com o firmware antigo, os canais modificados no separador de 990 canais podiam não transmitir. O bit é reposto também nos canais 1-960 que não o tinham.
- Verificadas as versões reais do Bajeton BJ-9000: firmware V0.22, V0.27 e V0.29 (idêntico ao da Radtel), programa v1.1.1 (960 canais) e v1.3.3 (990 canais, é o programa da Radtel). Canais e definições verificados também com o código dos programas da Bajeton.

### Atualização do firmware para outros rádios
- Radtel RT-950 Pro, Radtel RT-950 e Hiroyasu HI-UV850 Pro (ficheiros .BTF, mesmo carregador do BJ-9000).
- Radtel RT-920 e RT-910B (ficheiros .anc, mesmo protocolo do BJ-7800 e do BJ-7700).

### Rádios novos (controladores do CHIRP)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); na galeria também o Hiroyasu HI-UV850 Pro.
- Definições também para SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 e Anysecu AC-580.

## Novidades da 4.6.0

### Corrigido: Radtel RT-950 Pro / Bajeton BJ-9000 com o firmware recente
- A partir do firmware **0.24** (dezembro de 2025) o rádio tem **990 canais em 10 zonas de 99** (ou uma lista única de 990); até à **0.23** tinha **960 em 15 zonas de 64** (ou uma lista única de 960). O programa agora gere as duas versões.
- A versão é reconhecida pelos dados lidos do rádio (o rádio não a comunica): a leitura abre os canais no separador certo. Se os dados não chegarem, uma nota explica como verificar a versão no rádio.
- Os canais 961-990 só são escritos se mudarem, depois de os reler do rádio; se o rádio não os tiver (firmware até 0.23), nada é escrito.
- Canais e definições verificados com o código original dos programas da Radtel v1.2.0 (960 canais) e v1.3.3 (990 canais).

### Novo: Radtel RT-950 (sem «Pro»)
- 960 canais (15 zonas de 64) ou uma lista única, definições gerais; verificado com o código original do programa «BT-9000_CPS» v1.3.0.

### Corrigido: nome exato do rádio
- Rádios diferentes com a mesma gestão (p. ex. **Radtel RT-950 Pro** e **Bajeton BJ-9000**, **Baofeng K67 Pro** e **5RH Pro**, Radtel RT-910 e Bajeton BJ-7700) são agora chamados pelo seu nome no separador, na leitura, na escrita, na identificação e nos ficheiros de cópia.
- Na primeira identificação de um destes rádios, o programa pergunta qual é e depois memoriza-o.

### Corrigido: aviso de dados não guardados
- Depois de uma leitura, de uma escrita ou de um envio, os canais estão no rádio: ao fechar o programa, o aviso só aparece se houver alterações reais.

### Rádios novos (controladores do CHIRP)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: os 990 canais em 10 zonas de 99, como no rádio.

Galeria: 489 modelos de 44 marcas (451 com canais, 386 com definições, 52 com atualização do firmware). Marcas novas: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
