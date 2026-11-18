# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 para Microsoft Windows 10/11

### Novidades
- **Kenwood TK-481 e TK-981** (900 MHz, canais divididos em sistemas) e os móveis **TK-780 e TK-880**: canais e definições, com os controladores do CHIRP.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: agora também as definições do rádio.
- **Radtel RT-13**: canais (22, sem nome), obtidos do programa do fabricante.
- **Radtel RT-69 GPS**: canais (999, com nomes de 12 caracteres), obtidos do programa do fabricante; antes de escrever verifica-se que cada canal está nas bandas do rádio.
- Mais robustez com ficheiros danificados: corrigidos dois bloqueios na leitura dos tons dos Baofeng UV-B5 / UV-B6 e UV-3R.
- Conversões: um canal que o rádio de destino não consegue memorizar é saltado com um aviso em vez de parar a conversão (firmware IJV, nicFW, OpenGD77, UV-KX, R3b0rn, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Galeria: 569 modelos de 47 marcas (515 com canais, 433 com definições).

### Notas
- TK-481 / TK-981: só são usados os sistemas que o rádio já tem; os de recursos partilhados (trunking) não são alterados.
- RT-13 e RT-69 GPS: se o rádio tiver uma palavra-passe, retire-a antes com o programa da Radtel. Obtidos dos programas do fabricante e testados só com um rádio simulado: faça primeiro apenas uma leitura.
- Programa independente, não afiliado aos fabricantes dos rádios.

---

Esta versão inclui também as novidades das versões 4.13.0 a 4.22.0, publicadas só como versões de teste. Aqui estão, da mais recente.

## IU0VYT Flash 4.22.0

### Novidades
- **Radtel RT-480 e RT-68**: canais e nomes (256 canais, numerados de 0 a 255, com nomes de 12 caracteres), obtidos do programa do fabricante. Alteram-se frequências, nome, potência, largura de banda, varredura e tons; as outras opções do canal ficam como estão.
- **Radtel RT-830 e RT-850**: canais (128, sem nome), obtidos dos programas do fabricante. Alteram-se frequências, potência, largura de banda, varredura e tons.
- **Radtel RT-20M**: abre-se com o separador Ruyage UV58Plus (mesma memória do iRadio UV-5118plus).

### Notas
- RT-480 / RT-68: só se escrevem canais e nomes; definições, VFO e DTMF do rádio não são tocados. Frequências em passos de 1,25 kHz; cada canal tem de ter frequência de transmissão.
- RT-830 / RT-850: definições e VFO voltam como estavam; a palavra-passe do rádio nunca é escrita. Se o rádio tiver uma palavra-passe, retire-a antes com o programa da Radtel.
- Todos obtidos dos programas do fabricante e testados só com um rádio simulado: faça primeiro apenas uma leitura.

## IU0VYT Flash 4.21.0

### Novidades
- **Radtel RT-5D (DMR)**: canais (16 zonas de 64, com nomes de 12 caracteres), obtidos do programa do fabricante. Alteram-se frequências, nome, potência, largura de banda, varredura e tons; as outras opções do canal ficam como estão.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** e os móveis **TK-7180, TK-8180** (também E): canais, com os controladores do CHIRP. A lista tem 250 posições para cada uma das primeiras 32 zonas; só são usadas as zonas que o rádio já tem, 512 canais no total.

### Notas
- RT-5D: a leitura obtém toda a memória e a escrita devolve-a inteira tal como está no ficheiro, como faz o programa da Radtel. Antes de escrever faça uma leitura recente: um ficheiro antigo traria de volta os contactos e as definições dessa altura. Testado só com um rádio simulado.
- TK-x180: só canais, sem definições; as zonas não se criam nem se apagam. Se o rádio tiver zonas não convencionais (trunking), a leitura é recusada.

## IU0VYT Flash 4.20.0

### Novidades
- **Radtel RT-4D e RT-6D (DMR) e RT-1000 Pro**: canais (1024, com nomes de 16 caracteres) e zonas (250), obtidos dos programas do fabricante. Alteram-se frequências, nome, potência, largura de banda, tons e, no RT-4D e RT-6D, a varredura; as outras opções do canal ficam como estão. Só se escrevem canais e zonas: contactos, grupos, mensagens, rádio FM e definições não são tocados.

### Notas
- Testados só com um rádio simulado: faça primeiro apenas uma leitura e comunique qualquer problema.
- «Identificar» não distingue estes rádios dos outros Radtel e Bajeton com a mesma ligação: o separador tem de ser escolhido à mão.
- RT-1000 Pro: a exclusão da varredura não é gerida (fica como está no rádio).

## IU0VYT Flash 4.19.0

### Novidades
- **Radtel RT-817A / RT-871A**: canais (1024, com nomes de 16 caracteres), obtidos do programa do fabricante. Mesma ligação e mesmo formato de canal do RT-880G, sem as zonas. Só se leem e escrevem as áreas que o programa da Radtel usa (definições, canais, rádio FM); o logótipo não é tocado.

### Notas
- Testado só com um rádio simulado: faça primeiro apenas uma leitura e comunique qualquer problema.
- «Identificar» não distingue o RT-817A / RT-871A do RT-880G: o separador tem de ser escolhido à mão.

## IU0VYT Flash 4.18.0

### Novidades
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: portáteis profissionais com os canais divididos em zonas (grupos), canais e definições. A lista mostra as 250 posições da memória; um canal novo entra na zona do canal que o antecede.
- **Quansheng UV-K5 / UV-K6 com o firmware R3b0rn de fagci**: canais (298 com a memória original, até 1024 com uma memória maior). Definições, VFO, predefinições e correção SSB nunca são escritos.
- **Radtel RT-330 e RT-400**: canais, obtidos dos programas do fabricante.

### Notas
- Os controladores do CHIRP dos portáteis TK estão declarados experimentais. As zonas não se criam nem se apagam a partir do programa.
- O firmware R3b0rn, o RT-330 e o RT-400 estão testados só com um rádio simulado: faça primeiro apenas uma leitura e comunique qualquer problema.
- Com o firmware R3b0rn, um canal sem nome fica com a frequência como nome (para esse firmware não existe canal sem nome).
- O separador RT-400 recusa os rádios que se identificam como RT-470, RT-630, RT-495 ou HI-8811: use-o só com o RT-400.

## IU0VYT Flash 4.17.0

### Novidades
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: canais lidos e escritos um de cada vez, com o rádio ligado normalmente (9600 baud). Na escrita só se enviam os canais alterados.
- **Portáteis profissionais Kenwood TK**: canais e definições.
  - Série 60: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - Série 60G: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: usa-se o separador «Radtel RT-880G» (mesma ligação e mesma memória dos canais).

### Notas
- Os Kenwood TH-F6/F7, TH-G71, TH-D7 e TH-K2 usam os comandos dos controladores do CHIRP e estão testados só com um rádio simulado: faça primeiro apenas uma leitura e comunique qualquer problema.
- Os controladores do CHIRP dos portáteis TK estão declarados experimentais.
- O RT-860 resulta da comparação com o programa do fabricante e não foi testado num rádio real.

## IU0VYT Flash 4.16.0

### Novidades
- **OpenGD77**: canais dos rádios DMR com o firmware livre OpenGD77.
  - Radioddity GD-77 e GD-77S, Baofeng DM-1801 e RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 e MD-2017, Retevis RT3S e RT84.
  - O rádio aparece como porta série USB; só se escrevem os canais e as zonas alterados, nunca definições, contactos, grupos e calibração.
- **Quansheng UV-K5, outros firmware livres**:
  - **NUNU** de kamilsss655: canais e definições.
  - **UV-KX** de joaquim.org: canais e definições; a calibração nunca é escrita.
  - O separador «OSFW» vale também para os firmware originais modificados com bandas alargadas (uvmod-kitchen, coleção de Tunas1337).
- **Radtel RT-880** sem GPS com o firmware original: usa-se o separador «Radtel RT-880G».

### Notas
- O OpenGD77 foi escrito a partir do protocolo descrito no programa livre qdmr e testado só com um rádio simulado: faça primeiro apenas uma leitura, guarde a cópia de segurança e comunique qualquer problema.
- NUNU e UV-KX são comparados byte a byte com os controladores para o CHIRP dos seus autores.

## IU0VYT Flash 4.15.0

### Novidades
- **D-STAR e C4FM**: canais por cabo para os portáteis digitais de radioamador, com ficheiros .img compatíveis com o CHIRP. A coluna «Modo» mostra «D-STAR» ou «C4FM».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) e IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (a cópia começa no rádio: as instruções estão na janela do cabo).
  - **Kenwood** (marca nova, 47 no total) TH-D72, TH-D74, TH-D75, com o cabo USB.
  - Definições para IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D e TH-D72.
- **Firmware livres**:
  - **nicFW 2.x** do TIDRADIO TD-H3: canais e definições principais.
  - **nicFW880 / nicFW950** dos Radtel RT-880, RT-880G e RT-950 Pro: canais.
  - **F4HWN Fusion 5.x** dos Quansheng UV-K1 e UV-K5 V3: 1024 canais e definições.
  - Com o nicFW a calibração do rádio nunca é escrita.
- **Kenwood na galeria**: também os portáteis PMR446 e DMR/NXDN, entre os não compatíveis (só se programam com o programa KPG reservado aos revendedores).

### Correções
- Definições: um campo de texto nunca usado lê-se vazio em vez de «valor não previsto»; um valor não alterado não é reescrito.
- Puxing PX-888K: removidas seis opções das definições que podiam escrever um dado errado (códigos BOT/EOT e tempos DTMF).
- Yaesu VX-3: nos códigos DTMF o espaço deixou de ser aceite (o rádio não o tem).

### Notas
- Controladores escritos de raiz a partir dos do CHIRP (GNU GPL), da documentação publicada por nicsure (nicFW) e do controlador de F4HWN, e comparados byte a byte usando rádios simulados. Os rádios reais desta versão ainda não foram testados: guarde a cópia de segurança e comunique qualquer problema.
- Para Icom, Yaesu e Kenwood não há atualização do firmware (faz-se com os programas dos fabricantes).

## IU0VYT Flash 4.14.0

### Novidades
- **Yaesu e Icom**: duas marcas novas (46 no total) e mais 13 rádios, com canais por cabo e ficheiros .img compatíveis com o CHIRP.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (cabo SCU-35) e FT-60R/E (cópia iniciada no rádio, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (cabo OPC-478).
  - Definições para todos exceto o IC-T70. Para estes rádios não há atualização do firmware.
- **Mais duas línguas**: malaio (Bahasa Melayu) e hindi (हिन्दी), 13 no total: programa, site, guia e ficheiros informativos, README e política de segurança.
- **Conversor e «Enviar para outro rádio»** para os Yaesu e os Icom: os tons que o rádio não tem são removidos com um aviso e os canais que não consegue memorizar são saltados com um aviso, em vez de parar toda a operação.

### Correções
- Traduções: cinco palavras a negrito dos guias ficavam em italiano em todas as línguas; duas frases erradas em coreano e em árabe.
- Wouxun KG-UV8D: nomes internos dos campos alinhados com o CHIRP atual (sem efeito nos canais).

### Notas
- Controladores escritos de raiz a partir dos do CHIRP (GNU GPL) e comparados byte a byte com o CHIRP usando rádios simulados. Os rádios Yaesu e Icom reais ainda não foram testados: guarde a cópia de segurança e comunique qualquer problema.
- As traduções em malaio e em hindi são novas: comunique os textos pouco claros.

## IU0VYT Flash 4.13.1

### Novidades
- **Firmware IJV** para os Quansheng UV-K5, UV-K5 V3 e UV-K1 (firmware não original de IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): cinco separadores, um por versão do firmware.
  - **V3** de 200 ou 999 canais no UV-K5 clássico (hardware V1).
  - **V4**, **V5 («FW5»)** e **V6** no UV-K5 V3 e no UV-K1: os 500 canais do banco escolhido no rádio, com os nomes dos grupos e dos bancos.
  - Leem-se e escrevem-se canais e definições principais (menu, ecrã, VOX, teclas laterais). Antes de escrever, o programa verifica que no rádio continua escolhido o mesmo banco e escreve só as partes alteradas; as calibrações nunca são tocadas.
  - **V6 (novo na 4.13.1)**: ligação de dados da V6, escrita por setores de 4 KiB, cada um relido e verificado. Se o rádio disser que falta a ligação de dados, atualize-o com o FLASH V6 do RadioManager.
- **Atualização do firmware**: um ficheiro .bin com uma versão IJV é assinalado, com o separador a usar para os canais e um aviso para V5 e V6.
- Traduções nas 11 línguas.

### Notas
- Protocolo e memória obtidos dos módulos CHIRP publicados pela IJV e da análise do RadioManager 4.3.2; código escrito de raiz, testado com rádios simulados. Os rádios reais com IJV ainda não foram testados: comunique qualquer problema.
- Módulos, opções, licença e firmware IJV continuam a pertencer ao RadioManager da IJV.

## IU0VYT Flash 4.13.0

### Novidades
- **Firmware IJV** para os Quansheng UV-K5, UV-K5 V3 e UV-K1 (firmware não original de IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): quatro separadores novos, um por versão do firmware.
  - **V3** de 200 ou 999 canais no UV-K5 clássico (hardware V1).
  - **V4** e **V5 («FW5»)** no UV-K5 V3 e no UV-K1: os 500 canais do banco escolhido no rádio, com os nomes dos grupos e dos bancos.
  - Leem-se e escrevem-se canais e definições principais (menu, ecrã, VOX, teclas laterais). Antes de escrever, o programa verifica que no rádio continua escolhido o mesmo banco e escreve só os blocos alterados; as calibrações nunca são tocadas.
  - A **V6** é reconhecida: canais e definições da V6 gerem-se com o RadioManager da IJV.
- **Atualização do firmware**: um ficheiro .bin com uma versão IJV é assinalado, com o separador a usar para os canais e um aviso para V5 e V6.
- Traduções nas 11 línguas.

### Notas
- Protocolo e memória obtidos dos módulos CHIRP publicados pela IJV e da análise do RadioManager 4.3.2; código escrito de raiz, testado com rádios simulados. Os rádios reais com IJV ainda não foram testados: comunique qualquer problema.
