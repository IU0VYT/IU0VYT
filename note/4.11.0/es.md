# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 para Microsoft Windows 10/11

### Nombres de las zonas
- En lugar de «Zona 1», «Zona 2»… cada zona puede tener un nombre, como en los programas del fabricante: botón «✎ Nombres de las zonas…» o doble clic en la zona. La barra de las zonas y la línea inferior muestran «número · nombre».
- Radios: Radtel RT-950 Pro, Bajeton BJ-9000 y Hiroyasu HI-UV850 Pro (firmware hasta 0.23 y desde 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 sin Bluetooth; zonas de las radios DMR Baofeng RD-5R, DM-1801 y DM-32UV.
- Los nombres se leen y se escriben por cable junto con los canales y se guardan en el archivo .img. Longitud y caracteres como en los programas del fabricante (12 caracteres, también chinos, en la RT-950 Pro; 10 en la RT-920; 16 en las DMR).
- Seguridad: el área de los nombres solo se escribe si los nombres cambian, tras volver a leerla de la radio; si la radio tiene allí datos que no son nombres, no se escribe nada.

### Edición en bloque
- «✎ Edición en bloque…» (también con el botón derecho) cambia a la vez potencia, ancho de banda, escaneo, tono TX, tono RX, desplazamiento y transmisión de los canales marcados (o seleccionados). Las opciones dejadas en «no cambiar» quedan como están; Ctrl+Z lo deshace todo.

### Dos idiomas más (11 en total)
- Coreano, con la bandera de Corea del Sur.
- Árabe, con la bandera de los Emiratos Árabes Unidos y la página de derecha a izquierda (frecuencias y números siguen de izquierda a derecha).
- Todos los archivos informativos de la carpeta `lingue` están traducidos también al coreano y al árabe y actualizados a la 4.11 en todos los idiomas.

### Otros
- Sitio web: se indican los idiomas disponibles; se ha quitado la lista de archivos de firmware.
- `go test` pasa en Windows (también i18n_test.go); una nueva prueba comprueba que LEGGIMI.txt, AUTORI_E_LICENZE.txt y la licencia son idénticos en la carpeta principal y en «sorgenti».

Los nuevos formatos se obtienen del código de los programas del fabricante y del controlador de CHIRP y se han probado con radios simuladas; no se han probado con radios reales: conserva siempre las copias de seguridad.

Galería: 489 modelos de 44 marcas (451 con canales, 396 con ajustes, 57 con actualización del firmware).
