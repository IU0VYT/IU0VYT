# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 para Microsoft Windows 10/11

Esta versión incluye también las novedades de la 4.3.0, 4.2.0, 4.1.0 y 4.0.0, que no tuvieron una versión estable propia.

## Novedades de la 4.4.0

### Corregido: recuadro «Identificación de la radio en curso» sin radio
- El recuadro podía aparecer también con un cable sin radio: las interferencias de la línea y el eco de los comandos enviados parecían una radio que responde.
- Ahora aparece **solo para el cable conectado a una radio que responde de verdad**. Con dos cables y una sola radio (COM6 sin radio, COM7 con la radio, o al revés) aparece solo para el puerto de la radio.
- Sin radio, la ronda de pruebas es además más rápida.

### Barra superior, Inicio, pestaña Identificar
- El botón se llama **«🔎 Identificar radio»** y está a la derecha, junto a «Actualizaciones».
- Con muchas pestañas abiertas, el botón se estrecha solo, hasta quedar en la lupa.
- Inicio: vuelven las descripciones de los recuadros de la 4.2. Abajo siguen visibles copyright, sitio web y código fuente, también a 1280×720.
- Pestaña Identificar: explicaciones breves y claras.

### Firmware Retevis (43 modelos con actualización del firmware)
- **Retevis RA79 V3** (y UV-K5 V3 / UV-K1, cargador 7.x) con los archivos .bin de Retevis, por ejemplo «RA79 V3.bin». El firmware solo se envía si el cargador de la radio es de la misma familia.
- **Retevis RA89** con los archivos .bin de Retevis, por ejemplo «Eng_RA89_Upadat_20250524_B91.38.bin». Usa el protocolo del programa de Retevis «RA89-update-software», que se ejecutó con una radio simulada. Los comandos enviados son idénticos byte a byte.

### Otros
- Las pruebas con radios simuladas ya no impiden compilar las pruebas en Windows.

Galería: 386 modelos de 27 marcas (348 con canales, 302 con ajustes, 43 con actualización del firmware).

## Novedades de la 4.3.0

### Identificación automática: más rápida, más radios, más puertos
- Una ronda completa de pruebas sin radio dura unos **50 segundos** (antes 135).
- El programa **escucha continuamente todos los puertos de los cables**, cada uno por su cuenta y todos a la vez:
  - reconoce también una radio conectada apagada y encendida después;
  - reconoce **varias radios en cables distintos** al mismo tiempo, muestra sus datos y abre todas sus pestañas.
- Las **radios ya reconocidas** (p. ej. K67 Pro, Bajeton 8300, BJ-9000) se prueban primero, luego todas las demás.
- Si la pestaña de la radio ya está abierta aparece el aviso «¡¡Radio ya identificada!! Pestaña de la radio ya abierta» (con varias radios «¡¡Radios ya identificadas!! Pestañas de las radios ya abiertas»).
- Pestaña Identificar: puertos en orden (COM6, COM7… COM10) y escritos como «COM6 — USB-SERIAL CH340»; «Identificar nueva radio automáticamente» en negrita.

### Actualización del firmware: 8 radios más (41 en total)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** con los archivos .kdhx de Radtel (cargador «JJCC Bootloader Pro»). RT-470: el firmware 1.x es solo para la placa antigua, el 2.x solo para la nueva. El programa lo indica y pide una confirmación escrita.
- **TIDRADIO TD-H3, TD-H8 y Radtel RT-730** con los archivos .BIN del fabricante: inicia la actualización y enciende la radio manteniendo pulsado el PTT.

### Otros
- **Bajeton BJ-9000**: como la BJ-8300, los 960 canales en una única lista sin zonas («Canales: todos (960)»).
- **Inicio** más compacto: abajo se ven siempre versión, copyright, sitio web y código fuente.
- **Galería**: una línea gruesa separa los modelos compatibles de los no compatibles.
- **LEGGIMI**: se ha quitado una parte larga repetida por error en la 4.2.0.

Galería: 386 modelos de 27 marcas (348 con canales, 302 con ajustes, 41 con actualización del firmware).

## Novedades de la 4.2.0

- Identificación automática corregida: el recuadro «Identificación de la radio en curso» aparece solo cuando responde una radio conectada (ya no al iniciar con el cable solo) y desaparece cuando la radio se reconoce o no se encuentra. Tras el reconocimiento se abren los datos y la pestaña de la radio y la búsqueda se detiene, sin desactivar la función.
- Nuevo botón «🔎 Identificar nueva radio automáticamente» en la barra superior, entre «Info» y «Actualizaciones»: busca una nueva radio con las mismas reglas y está en gris si la identificación automática está desactivada.
- 35 radios nuevas (36 versiones) y 8 marcas más: familia TYT TH-UV88 (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), familia Radtel RT-490 (con JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 radios de la familia Retevis RT21 (con Abbree AR-63) y Quansheng TK11.
- Actualización del firmware también para Radtel RT-590 / RT-590 Plus y RT-600 / RT-600 Plus con los archivos .bin de Radtel.
- Corregidos algunos defectos ocultos en la lectura de los nombres de canal con letras acentuadas o chinas.
- Galería de 386 modelos de 27 marcas: 348 con canales y repetidores, 302 con ajustes, 33 con actualización del firmware.

## Novedades de la 4.1.0

- Identificación automática corregida: funciona en cualquier pestaña y, cuando desconectas la radio, la apagas o la cambias, la búsqueda vuelve a empezar sola y reconoce la radio conectada después (aunque sea la misma o la enciendas después del cable). Antes, tras el primer reconocimiento, ya no buscaba otras radios.
- Durante la búsqueda automática aparece el recuadro «Identificación de la radio en curso. Espera, por favor.», con el botón «Interrumpir».
- 14 radios nuevas y 3 marcas más: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 y RT-730, Hiroyasu HI-8811, Talkpod A36plus y A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (2.ª y 3.ª generación) y TD-H3, en las versiones base, HAM y GMRS.
- Actualización del firmware de las Quansheng UV-K5, UV-K5(8), UV-K6 y UV-5R Plus con los archivos .bin de Quansheng, comprobando que el firmware sea para el tipo de radio conectada.
- Pantalla de inicio y sitio web: «GESTIÓN DE CANALES - FRECUENCIAS - REPETIDORES - FIRMWARE» y los nombres de todas las marcas en un marco al estilo de los videojuegos de los 80.
- Galería de 349 modelos de 19 marcas: 311 con canales y repetidores, 265 con ajustes, 31 con actualización del firmware.

## Novedades de la 4.0.0

- Programa multimarca: 335 modelos de 16 marcas en la galería, 297 con canales y frecuencias por cable y 251 con los ajustes de la radio.
- Radios nuevas: Quansheng UV-K5, UV-K6, UV-5R Plus y UV-K1; 31 modelos Retevis (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920 y T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S y GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton y TID.
- Identificación automática (pestaña «Identificar»): conectas el cable y la radio se reconoce sola, con su pestaña abierta y los canales leídos. La elección se recuerda y se desactiva con el mismo botón.
- Galería de modelos con el menú de marcas y un color para cada marca.
- «Leer de la radio» y «Escribir en la radio» en azul en las pestañas de las radios.
- Inicio, Info, guía, sitio web y documentación actualizados al programa multimarca.
