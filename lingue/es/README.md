# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="../../lingue/fr/README.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="../../lingue/de/README.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub><b>Español</b></sub></td>
<td align="center"><a href="../../lingue/pt/README.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="../../lingue/ru/README.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="../../lingue/zh/README.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="../../lingue/ja/README.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="../../lingue/ko/README.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="../../lingue/ar/README.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="../../lingue/ms/README.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="../../lingue/hi/README.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/IU0VYT/IU0VYT/releases/latest"><img src="https://img.shields.io/github/v/release/IU0VYT/IU0VYT?label=release&color=2ea44f" alt="Latest release"></a>
<a href="https://github.com/IU0VYT/IU0VYT/releases"><img src="https://img.shields.io/github/downloads/IU0VYT/IU0VYT/total?color=1f6feb" alt="Downloads"></a>
<a href="../../LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0--or--later-blue" alt="License GPL-3.0-or-later"></a>
<img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d6" alt="Windows 10 | 11">
<img src="https://img.shields.io/badge/languages-13-orange" alt="13 languages">
</p>

**Versión 4.25.0** · Windows · interfaz en 13 idiomas (italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語, 한국어, العربية, Bahasa Melayu, हिन्दी) · un solo archivo ejecutable, sin instalación

Programa para radioaficionados y usuarios de transceptores de **52 marcas** (**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone, Yaesu, Icom, Kenwood): lee y escribe canales, frecuencias y ajustes por el cable de programación, gestiona un archivo de los repetidores italianos y actualiza el firmware de las radios compatibles.

🌐 Sitio del programa: **[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*Esta es la traducción al español de [README.it.md](../../README.it.md); el original italiano es el texto de referencia.*

## Funciones

- **Canales y frecuencias** de 544 modelos (incluidas radios PMR digitales y DMR, también por USB): lectura y escritura por cable, archivos `.img` compatibles con CHIRP, archivos originales del Baofeng K67 Pro (`.xlc`) y de la Bajeton 8300 (`.dat`).
- **Ajustes de las radios** (menús, pantalla, VOX, squelch, TOT, DTMF, teclas…) para 458 modelos, con la descripción de cada elemento y un aviso que confirmar para los delicados.
- **Identificación de la radio conectada**: el programa la reconoce solo y muestra marca, modelo, tipo, firmware y la demás información que comunica (para las radios DMR USB también ID DMR y nombre). Con la **identificación automática** (botón en la pestaña «Identificar», elección recordada) basta con conectar el cable, en cualquier pestaña: el programa escucha todos los puertos de los cables a la vez, reconoce también una radio encendida después de conectar el cable y varias radios conectadas a cables distintos, prueba primero las radios ya reconocidas y las más difundidas, no se deja engañar por las radios en modo de actualización del firmware y abre la pestaña de cada una con los canales leídos; el recuadro «Identificando la radio» solo aparece para el cable con una radio que responde; para otra radio está el botón **«🔎 Identificar radio»** arriba a la derecha.
- **Envío a otra radio**: los canales marcados en una radio se escriben directamente en otra radio conectada, elegida por marca y modelo, con control de compatibilidad, adaptación, copia de seguridad y verificación.
- **Convertidor universal**: los canales de cualquier radio o archivo (K67 `.xlc`, Bajeton `.dat`, imágenes `.img` también de CHIRP, CSV de CHIRP, Excel/CSV) en el formato de cualquier otra radio, adaptados a sus reglas.
- **Interfaz en 13 idiomas** con el selector de banderas junto a «ℹ Info»: traducción inmediata de todo el programa (árabe de derecha a izquierda), idioma recordado al reiniciar; archivos informativos traducidos en la carpeta `lingue`.
- **Nombres de las zonas** en lugar de «Zona 1», «Zona 2»… (Radtel RT-950 Pro, RT-950, RT-920, RT-910, Bajeton BJ-9000 y BJ-7700, Hiroyasu HI-UV850 Pro; zonas de las radios DMR RD-5R, DM-1801 y DM-32UV), leídos y escritos por cable junto con los canales.
- **Edición en bloque** de los canales marcados: potencia, ancho de banda, escaneo, subtonos, desplazamiento o transmisión a la vez, el resto de cada canal queda como está.
- **Galería de 598 modelos** de 52 marcas, divididos entre compatibles y no compatibles, por tipo (radioaficionado, DMR, GMRS/FRS, PMR446, para niños) y por marca, con búsqueda, filtros y enlaces a los programas y manuales oficiales.
- **Archivo de repetidores** basado en la lista actualizada de [IK2ANE](https://www.ik2ane.it), con filtros e inserción en las radios, con el control de compatibilidad entre repetidores y radio antes de escribir.
- **Actualización del firmware** de 57 modelos (también las Quansheng UV-K5, UV-K6 y UV-5R Plus, las Radtel RT-590 y RT-600 con los archivos .bin de Quansheng y de Radtel, las Radtel RT-470, RT-470L, RT-470X, RT-495 y RT-630 con los archivos .kdhx de Radtel, las TIDRADIO TD-H3, TD-H8 y Radtel RT-730 con los archivos .BIN del fabricante, las Retevis RA79 V3 y RA89 con los archivos .bin y la Retevis RT85 con los archivos .hex de Retevis, las Radtel RT-950 Pro y RT-950 y la Bajeton BJ-9000 con los archivos .BTF, las Radtel RT-920 y RT-910B con los archivos .anc, las Radtel RT-910 y RT-900 y la Bajeton BJ-7700 sin Bluetooth con los archivos .BTU, la Radtel RT-890 con los archivos .bin de Radtel y de M7OCM, la Baofeng DM-1701 por USB con los archivos .bin de Baofeng, las Baofeng RD-5R y DM-1801 por USB con los archivos .sgl), con control del archivo y copia de seguridad de la radio; actualización guiada con el programa oficial para BTECH DMR-6X2 y DMR-6X2 PRO.
- **Copia de seguridad** automática antes de cada escritura y verificación después.
- **Búsqueda de actualizaciones** del programa en cada arranque.

## Uso

1. Descarga [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip) (última versión) y extrae el archivo.
2. Inicia `IU0VYT Flash.exe` (Windows 10/11; usa el componente WebView2 de Windows).
3. En la pestaña **📻 Modelos** elige tu radio, conecta el cable y usa **Leer de la radio**.

La guía completa está en el archivo `LEEME.txt` (original italiano: `LEGGIMI.txt`).

## Compilación

Hace falta Go 1.22 o posterior. En la carpeta `sorgenti`:

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

Las bibliotecas están incluidas en la carpeta `vendor`.

## Advertencias

- Los protocolos de las radios se han reconstruido para la interoperabilidad y se han verificado con radios simuladas: **conserva siempre las copias de seguridad**. El programa se ofrece sin ninguna garantía.
- Programa independiente, **ni afiliado ni aprobado** por los fabricantes de las radios (Baofeng, Bajeton, Quansheng, Retevis, Radtel, Radioddity, TIDRADIO y los demás). Las marcas citadas pertenecen a sus respectivos titulares y solo sirven para indicar las radios compatibles.
- Respeta la normativa sobre el uso de las frecuencias de tu país.

## Agradecimientos

- [CHIRP](https://chirpmyradio.com) y sus autores: gran parte de los controladores de las radios deriva de su trabajo.
- [IK2ANE](https://www.ik2ane.it), que pone a disposición de todos, gratuitamente, la lista de los repetidores italianos.

Autores y licencias de los componentes: `AUTORES_Y_LICENCIAS.txt`.

## Licencia

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

Software libre distribuido con la **GNU General Public License versión 3 o posterior** (`GPL-3.0-or-later`): ver `LICENSE.txt` y `AVISO_DE_LICENCIA.txt`.
