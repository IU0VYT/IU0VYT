# IU0VYT Flash 4.5.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0 para Microsoft Windows 10/11

### Firmware Retevis RT85 (44 modelos con actualización del firmware)
- **Retevis RT85** con los archivos .hex de Retevis, por ejemplo «RT85-Firmware-V1.53.hex». Usa el protocolo del programa de Retevis «RT85 10W_5W_UpData», que se ejecutó con una radio simulada: los bloques y comandos enviados son idénticos byte a byte.
- Procedimiento de Retevis: radio apagada, encendida manteniendo pulsados **PTT y P1** (indicador verde). Después se elige la potencia de la radio (**5 W o 10 W**, como en el programa de Retevis) y se escribe RT85 para confirmar.

### Corregido: actualización de la Retevis RA89
- La 4.4 esperaba de la radio una respuesta equivocada a la presentación: con una RA89 real la actualización no empezaba (la radio no se tocaba). Ahora la respuesta esperada y las pausas son las del programa de Retevis.

### Puertos, pestañas Identificar y Convertir
- Los puertos de los cables están siempre en orden creciente, en cualquier sistema: COM1, COM2, COM3… COM10. También la identificación automática empieza por el primer puerto.
- Pestaña Identificar: «Las radios ya identificadas en el pasado se reconocen más rápido».
- Pestaña Convertir: «Elige un archivo Baofeng, Bajeton, un .img de todas las demás marcas de radio (también guardado con CHIRP), un CSV de CHIRP, tablas Excel o CSV».

### Sitio web
- Se ha quitado el texto de las notas de la versión debajo del botón de descarga.
- El título está más arriba: el número de versión bajo «PRESS START» se ve entero también en las pantallas de los portátiles.

Galería: 386 modelos de 27 marcas (348 con canales, 302 con ajustes, 44 con actualización del firmware).
