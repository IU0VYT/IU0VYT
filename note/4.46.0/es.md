# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="de.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub><b>Español</b></sub></td>
<td align="center"><a href="pt.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 para Microsoft Windows 10/11

### Correcciones
- **Radtel RT-880G**: ya no se indica como radio GMRS / FRS. Es un transceptor de radioaficionado VHF/UHF (con GPS y APRS), como la RT-880 y la RT-860; el aviso «Radio vendida como GMRS» ya no aparece antes de la escritura.
- **Verificación tras la escritura** en la Radtel RT-880G: fallaba porque a la radio, que al final de la escritura guarda los datos y se reinicia, se la buscaba demasiado pronto y demasiado a menudo. Ahora el programa espera 10 segundos sin molestarla y reintenta cada 15 segundos, hasta 90 segundos. Los mismos tiempos valen para la RT-880, la RT-860 y la RT-817A / RT-871A, que usan el mismo controlador.
- Controles más estrictos de los valores recibidos del AR8200, de las Kenwood y de la TK-180 (avisos de CodeQL).

### Notas
- La verificación se probó en una RT-880G real; la RT-880, la RT-860 y la RT-817A / RT-871A no se han probado con radios reales. Las demás radios mantienen los tiempos de antes.
- Ninguna radio nueva en esta versión.
- Programa independiente, no afiliado a los fabricantes de las radios.
- Batería completa de pruebas superada: parte general; 915 comparaciones de estructura y 380 lecturas / escrituras idénticas a CHIRP.
