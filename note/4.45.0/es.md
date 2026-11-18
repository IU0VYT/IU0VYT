# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0 para Microsoft Windows 10/11

### Novedades
- **Radios DMR de la familia AnyTone**: AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus y BTECH DMR-6X2 y DMR-6X2 PRO (estas dos eran «no compatibles»). Canales y zonas con el cable USB de la radio; memoria y protocolo del programa libre dmrconfig, verificados con su código original. Solo se escriben canales, zonas y listas de escaneo.
- **TYT MD-UV380, MD-UV390, MD-2017, MD-9600 y Retevis RT3S y RT82** con el firmware original: misma memoria y mismo protocolo que la Baofeng DM-1701.
- **Baofeng BF-666S** (controlador de la BF-888S) y **BF-T6** (experimental, controlador de la BF-T8).
- Para las fichas que siguen «no compatibles» UV-8D, BF-UV8, DM-1702 y DM-1703 ahora se indica el motivo.
- Galería: 673 modelos de 54 marcas (623 con canales, 556 con ajustes).

### Notas
- Ninguna de las radios nuevas se ha probado en una radio real: haz primero solo una lectura.
- BTECH DMR-6X2 PRO y Baofeng BF-T6 son experimentales.
- Programa independiente, no afiliado a los fabricantes de las radios.
- Batería completa de pruebas superada: parte general; 918 comparaciones de estructura y 381 lecturas / escrituras idénticas a CHIRP; familia AnyTone comparada con el código original de dmrconfig.

---

Esta versión incluye también las novedades de las versiones 4.24.0 a 4.44.0, que no tuvieron una versión estable propia. Aquí están, de la más reciente.

## IU0VYT Flash 4.44.0

### Novedades
- **Ajustes** para otras ocho radios que solo tenían los canales: Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro y Bajeton DM-5000, a partir de los programas de los fabricantes.
- **RT-480 y RT-68**: al escribir, ahora también llega a la radio el bloque de los ajustes generales.
- **RT-4D, RT-6D, RT-1000 Pro**: el archivo de la radio contiene también los ajustes (los archivos guardados antes siguen siendo válidos).

### Notas
- Ninguno de los ajustes nuevos se ha probado en una radio real.
- RT-69 GPS: por ahora solo las opciones verificadas una a una en el programa del fabricante.

## IU0VYT Flash 4.43.0

### Novedades
- **Error corregido**: en 33 radios (entre ellas Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22 y HA1G, Kenwood TK-2140 / TK-3140) y en las 8 «gemelas» de la 4.42, los ajustes no aparecían tras la lectura por cable, sino solo al abrir un archivo de CHIRP.
- **Ajustes** para otras cinco Radtel: RT-817A / RT-871A, RT-830, RT-850, RT-330 y RT-400, a partir de los programas del fabricante.

### Notas
- Ninguno de los ajustes nuevos se ha probado en una radio real.
- Radtel RT-830 y RT-850: una opción por radio la lee el programa del fabricante de forma distinta a como la escribe; aquí se sigue la escritura y la opción lleva un aviso.

## IU0VYT Flash 4.42.0

### Novedades
- **Pestaña Inicio**: ahora cabe entera en la ventana. El recuadro de las marcas es más ancho y la página se adapta sola al tamaño de la ventana.
- **Ajustes** para 30 modelos que solo tenían los canales: Baofeng serie D (BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro y Radtel RT-13.

### Notas
- Ninguno de los ajustes nuevos se ha probado en una radio real.
- Baofeng serie D: en el programa del fabricante «guía de voz» y «sonido de las teclas» están intercambiados entre lectura y escritura; aquí se sigue la escritura y las dos opciones llevan un aviso.

## IU0VYT Flash 4.41.0

### Novedades
- **Radios nuevas** (los portátiles de CHIRP que faltaban): Yaesu VX-1R (canales y ajustes), Icom IC-91AD e IC-92AD (dos bandas, también D-STAR), BTECH UV-PRO (por Bluetooth, sin cable) y el receptor AOR AR8200 (marca nueva).
- **Ajustes** para radios que solo tenían los canales: Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### Notas
- Ninguna de las radios nuevas se ha probado de verdad: haz primero solo una lectura.
- BTECH UV-PRO: la escritura solo se permite con la versión del firmware probada por el autor del controlador de CHIRP (147).

## IU0VYT Flash 4.40.0

### Novedades
- **Búsqueda automática más rápida con la pestaña ya abierta**: la comprobación breve (antes solo para la Baofeng K67 Pro) vale ahora también para las Icom, las Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, las Quansheng UV-K5 y derivadas, las Bajeton BJ-9000, BJ-7800, BJ-7700 y BJ-8300, las Radtel RT-900 / RT-950 / RT-470 y las radios con la misma memoria. Para las demás sigue la lectura completa.
- **Marca nueva: Alinco.** Portátiles DJ-596, DJ-175, DJ-G7EG y DJ-G7T; móviles DR-135T, DR-235T, DR-435T, DR-03T y DR-06T. Solo canales. Funciona también con los cables que no devuelven los datos.

### Notas
- Ninguna Alinco se ha probado de verdad: haz primero solo una lectura.
- La comprobación breve está probada con las radios simuladas; si tras la comprobación una radio no vuelve al funcionamiento normal, infórmalo.

## IU0VYT Flash 4.39.0

### Novedades
- **Cuatro radios Icom nuevas**: ID-80H (portátil D-STAR), ID-880H (móvil D-STAR), IC-W32A e IC-W32E (portátiles bibanda con dos mitades separadas: de 0 a 99 los canales VHF, de 100 a 199 los canales UHF). Solo canales.

### Notas
- Ninguna de las radios nuevas se ha probado de verdad: haz primero solo una lectura.

## IU0VYT Flash 4.38.0

### Correcciones
- **Identificación**: la Bajeton BJ-9000 a veces se bloqueaba, sobre todo en la primera identificación. El programa vuelve a escuchar el puerto durante un segundo antes del primer saludo, como en la 4.35, también para la radio ya reconocida en ese cable. Las demás mejoras de la 4.36 se mantienen.

### Notas
- Ninguna radio nueva en esta versión.

## IU0VYT Flash 4.37.0

### Novedades
- **Ocho portátiles Yaesu nuevos**: VX-5R y VX-7R (canales); VX-2R, FT-50R, VX-150, VX-8R, VX-8DR y VX-8GE (canales y ajustes; en las VX-8 también APRS).
- La copia empieza desde la radio (modo «CLONE»): la pestaña se elige a mano y las instrucciones están en la ventana del cable.

### Notas
- Ninguna de las radios nuevas se ha probado de verdad: haz primero solo una lectura.
- No hecho: Yaesu VX-1.

## IU0VYT Flash 4.36.0

### Correcciones
- **«Identificar» e identificación automática**: el programa recuerda en qué cable reconoció cada radio y empieza por esas; la primera búsqueda ya no pierde segundos escuchando y preparando; la radio encendida después del cable se vuelve a probar más a menudo.
- El recuadro «Identificación de la radio en curso» aparece solo mientras la radio se deja leer, con el modelo y el avance. Si la radio está conectada pero apagada, el programa avisa de que hay que encenderla y la reconoce en cuanto se enciende.
- Baofeng K67 Pro con la pestaña ya abierta: reconocida en un par de segundos, sin volver a leer toda la memoria.

### Novedades
- Pestaña «Identificar»: botón **«Guardar el diagnóstico de la identificación»** (tiempos y primeros bytes de las últimas búsquedas en un archivo de texto, para enviarlo al autor si la identificación es lenta o se equivoca; no contiene canales ni datos personales).

### Notas
- Ninguna radio nueva en esta versión.

## IU0VYT Flash 4.35.0

### Correcciones
- **«Identificar» e identificación automática**: se ha quitado la «ronda rápida» de la 4.31, que con radios reales podía dejar la radio bloqueada en programación y hacer lenta la búsqueda. Las esperas vuelven a ser las de antes.
- La búsqueda prueba los mismos grupos de radios en cualquier PC (antes cambiaban de un inicio a otro).
- El eco de un cable sin radio ya no se confunde con una radio.

### Notas
- Ninguna radio nueva en esta versión.

## IU0VYT Flash 4.34.0

### Novedades
- **Radios nuevas**: Yaesu VX-170 y VX-177 (portátiles), FT-7800 / FT-7900 y FT-8900 (móviles), solo canales, con los controladores de CHIRP: lectura, escritura y bytes intercambiados con la radio idénticos a CHIRP.

### Notas
- La copia empieza desde la radio (modo «CLONE»): la pestaña hay que elegirla a mano, «Identificar» no prueba estas radios.

## IU0VYT Flash 4.33.0

### Novedades
- **Radios nuevas**: Icom IC-T7H, IC-T8A, IC-Q7A e IC-P7 (canales), con los controladores de CHIRP: lectura, escritura y bytes intercambiados con la radio idénticos a CHIRP.

### Notas
- IC-Q7A e IC-P7: el modo (FM, FM ancha, AM) sigue a la frecuencia; el salto «P» y el paso de sintonía del canal se conservan.

## IU0VYT Flash 4.32.0

### Novedades
- **Radios nuevas**: Baofeng T22 en versión FRS, TIDRADIO TD-M8S, TD-V2, TD-H4 y TD-H11 (canales y ajustes), a partir de los programas de los fabricantes.
- **«Identificar»** propone también las radios nuevas que se presentan como la BF-888S, sin alargar la búsqueda: la radio se lee con su controlador solo cuando se abre esa pestaña.

### Correcciones
- TIDRADIO de 2026 (TD-H3 Plus / SE, TD-H7 Plus, TD-H8 de 4.ª generación, TD-H9, TD-638UV): la potencia media solo la ofrecen los modelos que de verdad la tienen, y nunca las versiones GMRS.
- Baofeng T22 PMR446: conexión como en el programa del fabricante.
- Notas de Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS y TIDRADIO TD-M2 Pro: «Identificar» sí las prueba; se ha quitado la frase que decía lo contrario.

## IU0VYT Flash 4.31.0

### Novedades
- **«Identificar» es mucho más rápido**: primero hace una ronda rápida de todas las radios y solo después, si no ha encontrado nada, la completa de antes. Una radio al final de la lista se encuentra en menos de medio minuto en lugar de casi dos.
- **«Identificar» prueba ahora también las radios añadidas en las últimas versiones**: Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G y las TIDRADIO sin pantalla.

### Correcciones
- Retevis RB39 / RB639: con la pestaña del país equivocado la radio se quedaba en programación hasta apagarla.

## IU0VYT Flash 4.30.0

### Novedades
- **Ajustes generales** para TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 de cuarta generación**, **TD-H9** y **TD-638UV**: teclas programables, silenciador, VOX, tiempo máximo de transmisión, pantalla, idioma, mensajes de encendido y más.
- **Radios nuevas (canales y ajustes)**: TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR** y **TD-M2 Pro**.

### Notas
- Las radios nuevas proceden de los programas del fabricante y aún no se han probado con una radio real: haz primero solo una lectura. «Identificar» no las distingue entre sí: elige tú la pestaña.

## IU0VYT Flash 4.29.0

### Novedades
- **Radios nuevas (solo canales)**: TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 de cuarta generación**, **TD-H9** y **TD-638UV**, cada una en las versiones base, HAM y GMRS.
- La TD-H3 Plus se presenta como la TD-H3: usa las pestañas «TD-H3 Plus», porque tiene tres potencias.

### Notas
- Estas radios proceden del programa del fabricante de 2026 («TIDRadioCPS») y aún no se han probado con una radio real: haz primero solo una lectura. Los ajustes generales aún no se gestionan.

## IU0VYT Flash 4.28.0

### Novedades
- **Radios nuevas** (móviles Radioddity): **UB15-G** (GMRS 15 W), **DB35-G** (GMRS 35 W), **DB35-UV** (VHF / UHF 35 W) y **DB40-G** (GMRS 40 W, 199 canales), con canales y ajustes generales. La UB15-G tiene dos pestañas, según el firmware: anterior a la V1.10 (sin nombres) y V1.10 (con nombres).
- La **DB40-G** está también en CHIRP: aquí lee y escribe los mismos bytes y los mismos archivos `.img`.

### Correcciones
- Las pruebas no compilaban en Windows desde la 4.26.0 (tres archivos de prueba usaban las radios simuladas, que en Windows no se compilan). El programa no se veía afectado.

### Notas
- UB15-G, DB35-G y DB35-UV proceden del programa del fabricante y de su archivo de configuración y aún no se han probado con una radio real: haz primero solo una lectura. «Identificar» no reconoce estas radios: la pestaña hay que elegirla a mano.

## IU0VYT Flash 4.27.0

### Novedades
- **Radios nuevas**: **Retevis C6** (FRS / GMRS, 30 canales), **Retevis C66** (UHF, 16 canales) y **Radioddity Echo-DE / Echo-IT** (PMR446, 16 canales), con canales y ajustes generales.
- Antes de escribir, el programa comprueba que la banda, el modelo y la región de la radio sean los de la memoria que se va a escribir; la contraseña y las calibraciones de la radio nunca se escriben.

### Correcciones
- **Retevis RB39 / RB39P**: ahora está entre las radios «GMRS / FRS (EE. UU.)» en la pestaña Modelos.

### Notas
- Las radios nuevas proceden del código de los programas de los fabricantes y aún no se han probado con una radio real: haz primero solo una lectura. «Identificar» no las reconoce: la pestaña hay que elegirla a mano.
- Si la radio tiene una contraseña de lectura, quítala antes con el programa del fabricante.

## IU0VYT Flash 4.26.0

### Novedades
- **Radios nuevas**: **Retevis RB39 / RB39P** (FRS, 22 canales) y **RB639 / RB639P** (PMR446, 16 canales), con canales y ajustes generales. La radio declara para qué país está hecha: con la pestaña equivocada la lectura se detiene e indica la correcta.
- **Ajustes nuevos** para **Baofeng BF-T17 / BF-T18** (PMR446 y FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3** y **PR-T3**: silenciador, tiempo máximo de transmisión, VOX, guía de voz, sonido de las teclas, ahorro de batería y los demás del programa del fabricante.

### Correcciones
- Registros de cambios en los 12 idiomas: fecha de la 4.25.0 corregida (12/11/2026).

### Notas
- Las radios nuevas proceden del código de los programas de los fabricantes y aún no se han probado con una radio real: haz primero solo una lectura. «Identificar» no las reconoce: la pestaña hay que elegirla a mano.

## IU0VYT Flash 4.25.0

### Novedades
- **Bajeton BJ-8300 con el firmware nuevo**: pestaña nueva para las radios con 990 canales en 10 zonas de 99 (programa del fabricante BT-11_CPS v1.2.6): canales, nombres de las zonas y ajustes. La pestaña «Bajeton 8300» de siempre (960 canales, 15 zonas de 64) queda como está y ahora **avisa** cuando encuentra el firmware nuevo.
- **Radios nuevas**: **Radioddity GS-10B Pro** (999 canales y ajustes), **Radioddity FS-T3** y **PR-T3** (canales).
- **Bajeton BJ-9000** con 990 canales: cinco ajustes más (los del programa v1.3.3).

### Correcciones
- **Baofeng BF-T17 / BF-T18, GT-18 y Radioddity PR-T5**: la potencia alta y la baja estaban intercambiadas. Quien haya escrito estas radios con una versión anterior, que las vuelva a leer y elija de nuevo la potencia.
- «Salida automática del menú» (BJ-8300, BJ-9000, Radtel RT-950): la última opción es 60 segundos, no 55.

### Notas
- Las radios y la pestaña nuevas proceden del código de los programas de los fabricantes y aún no se han probado con una radio real: haz primero solo una lectura.
- BJ-8300: para saber qué pestaña usar, mira en la radio cuántas zonas hay (15 de 64 canales, o 10 de 99).

## IU0VYT Flash 4.24.0

### Novedades
- **Pestaña Modelos más legible**: las etiquetas de los tipos de radio están en el centro de la página y la línea entre radios compatibles y no compatibles es roja y más gruesa.
- **26 radios más**, gestionadas con los controladores de radios iguales ya presentes:
  - reconocidas comparando los programas de los fabricantes: **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28**;
  - de la tabla de modelos equivalentes de CHIRP: Radioddity GA-5S, Retevis RT5 (tres potencias), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+ y TS-T9+, Rugged RH5R y RH5X, Ansoko A-5R y A-8S, Arcshell AR-5, AR-6 y AR-8, Greaval GV-8S y GV-9S, Tenway TW-325, UV-82 Pro y UV-5R Pro.

### Notas
- Las seis radios reconocidas comparando los programas de los fabricantes aún no se han probado con una radio real: haz primero solo una lectura. Si el modelo no es el correcto, la radio no responde y no se modifica.
