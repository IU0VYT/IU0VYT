# IU0VYT Flash 4.5.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página da versão no GitHub (em inglês), com a transferência: [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0 para Microsoft Windows 10/11

### Firmware Retevis RT85 (44 modelos com atualização do firmware)
- **Retevis RT85** com os ficheiros .hex da Retevis, por exemplo «RT85-Firmware-V1.53.hex». Usa o protocolo do programa da Retevis «RT85 10W_5W_UpData», que foi executado com um rádio simulado: os blocos e comandos enviados são idênticos byte a byte.
- Procedimento da Retevis: rádio desligado, ligado mantendo premidos **PTT e P1** (indicador verde). Depois escolhe-se a potência do rádio (**5 W ou 10 W**, como no programa da Retevis) e escreve-se RT85 para confirmar.

### Corrigido: atualização do Retevis RA89
- A 4.4 esperava do rádio uma resposta errada à apresentação: com um RA89 real a atualização não arrancava (o rádio não era alterado). Agora a resposta esperada e as pausas são as do programa da Retevis.

### Portas, separadores Identificar e Converter
- As portas dos cabos estão sempre por ordem crescente, em qualquer sistema: COM1, COM2, COM3… COM10. Também a identificação automática começa pela primeira porta.
- Separador Identificar: «Os rádios já identificados no passado são reconhecidos mais depressa».
- Separador Converter: «Escolha um ficheiro Baofeng, Bajeton, um .img de todas as outras marcas de rádio (também guardado com o CHIRP), um CSV do CHIRP, tabelas Excel ou CSV».

### Site
- Removido o texto das notas da versão por baixo do botão de transferência.
- O título está mais acima: o número da versão por baixo de «PRESS START» vê-se inteiro também nos ecrãs dos portáteis.

Galeria: 386 modelos de 27 marcas (348 com canais, 302 com definições, 44 com atualização do firmware).
