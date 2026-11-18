# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 para Microsoft Windows 10/11

Esta versión incluye también las novedades de la 4.8.0, la 4.7.0 y la 4.6.0, publicadas solo como versiones de prueba.

## Novedades de la 4.9.0

### Corregido: Identificar con Lanchonlh HG-UV98 y Quansheng TG-UV2+
- Durante «Identificar» el puerto hablaba siempre en 8N1. Ahora usa paridad par (HG-UV98) y dos bits de parada (TG-UV2+), como la lectura desde la pestaña de la radio. Las dos radios ahora se reconocen.

### Corregido: paridad de la HG-UV98
- Con paridad par, el puerto de Windows ahora también controla los errores de paridad. Antes, el control se activaba y se desactivaba enseguida.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus con firmware libres
- Nuevas pestañas para los firmware egzumer, F4HWN y LOSEHU/LTS/UVE.
- F4HWN usa el controlador F4HWN de armel (fuera de CHIRP): 8 niveles de potencia, bloqueo de la transmisión, 3 listas de escaneo y los ajustes del firmware.
- Cada pestaña lee y escribe solo su firmware. Tras actualizar el firmware, el programa indica la pestaña que hay que usar.

### Radios nuevas
- Wouxun (20 modelos): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (EE. UU.)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (firmware de Radtel y firmware libre M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

Para las radios de los controladores de CHIRP, canales y ajustes se verifican con las pruebas automáticas frente a CHIRP. Las radios GMRS, FRS y MURS son para Estados Unidos: en Italia no se permite transmitir en esas frecuencias.

## Novedades de la 4.8.0

### Actualización del firmware .BTU (Radtel RT-910 y RT-900 sin Bluetooth, Bajeton BJ-7700 sin Bluetooth)
- Los firmware .BTU se actualizan con el protocolo del programa «Radtel Bootload Tools». El cargador es el de la Baofeng UV-5RM, con la presentación «PROGRAMBT80U».
- Modo automático, con la radio encendida normalmente. Modo manual: radio apagada, encendida manteniendo pulsadas las dos teclas laterales, hasta que la pantalla muestre «UPDATE».
- El cargador no comprueba el modelo. Por eso el programa reconoce la radio y la versión de la electrónica por el nombre del archivo, muestra los avisos y pide siempre escribir el modelo.
- En la galería, RT-910 (sin Bluetooth, .BTU) y RT-910B (con Bluetooth, .anc) tienen ahora dos fichas. Lo mismo vale para BJ-7700 y BJ-7700 sin Bluetooth.

### Corregido: Radtel RT-900 sin Bluetooth con el firmware de 512 canales
- Hasta la V1.09P la radio tiene 512 canales; desde la V1.14P tiene 999.
- Ahora los canales 513-999 solo se escriben si los modificas. Antes, el programa comprueba en la radio que esa zona esté vacía o contenga canales.

### Radios nuevas (controladores de CHIRP)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (la nueva versión de la H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

Para todas, canales y ajustes se verifican con las pruebas automáticas frente a CHIRP.

## Novedades de la 4.7.0

### Corregido: identificación automática bloqueada o lentísima
- Una radio encendida en modo de actualización del firmware transmite sola mensajes periódicos: antes parecía responder a todos los modelos y la búsqueda duraba minutos. Ahora el programa escucha el puerto antes de probar y no se deja engañar: la búsqueda termina en menos de un minuto y la pestaña Identificar explica qué ocurre.
- La interfaz ya no se bloquea mientras arrancan las búsquedas; las radios más comunes se prueban primero.
- Corregido el falso aviso «Pestaña de la radio ya abierta» para una pestaña cerrada antes.

### Radios con la misma gestión
- La ventana de selección y la pestaña Identificar muestran todas las radios posibles (p. ej. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), con el menú para elegir y una breve indicación de dónde ver la versión del firmware en la radio.

### Corregido: Radtel RT-950 Pro / Bajeton BJ-9000 con el firmware hasta 0.23
- El bit «transmisión habilitada» de los canales ahora se escribe siempre, con las dos versiones (como los programas de Radtel hasta la v1.2.x; la v1.3.3 lo ignora). Con el firmware antiguo, los canales modificados en la pestaña de 990 canales podían no transmitir. El bit se restablece también en los canales 1-960 que no lo tenían.
- Verificadas las versiones reales de la Bajeton BJ-9000: firmware V0.22, V0.27 y V0.29 (idéntico al de Radtel), programa v1.1.1 (960 canales) y v1.3.3 (990 canales, es el programa de Radtel). Canales y ajustes verificados también con el código de los programas de Bajeton.

### Actualización del firmware para otras radios
- Radtel RT-950 Pro, Radtel RT-950 y Hiroyasu HI-UV850 Pro (archivos .BTF, mismo cargador que la BJ-9000).
- Radtel RT-920 y RT-910B (archivos .anc, mismo protocolo que BJ-7800 y BJ-7700).

### Radios nuevas (controladores de CHIRP)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); en la galería también Hiroyasu HI-UV850 Pro.
- Ajustes también para SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 y Anysecu AC-580.

## Novedades de la 4.6.0

### Corregido: Radtel RT-950 Pro / Bajeton BJ-9000 con el firmware reciente
- Desde el firmware **0.24** (diciembre de 2025) la radio tiene **990 canales en 10 zonas de 99** (o una única lista de 990); hasta la **0.23** tenía **960 en 15 zonas de 64** (o una única lista de 960). El programa ahora gestiona las dos versiones.
- La versión se reconoce por los datos leídos de la radio (la radio no la comunica): la lectura abre los canales en la pestaña correcta. Si los datos no bastan, una nota explica cómo comprobar la versión en la radio.
- Los canales 961-990 solo se escriben si cambian, tras volver a leerlos de la radio; si la radio no los tiene (firmware hasta 0.23), no se escribe nada.
- Canales y ajustes verificados con el código original de los programas de Radtel v1.2.0 (960 canales) y v1.3.3 (990 canales).

### Nueva: Radtel RT-950 (sin «Pro»)
- 960 canales (15 zonas de 64) o una única lista, ajustes generales; verificada con el código original del programa «BT-9000_CPS» v1.3.0.

### Corregido: nombre exacto de la radio
- Radios distintas con la misma gestión (p. ej. **Radtel RT-950 Pro** y **Bajeton BJ-9000**, **Baofeng K67 Pro** y **5RH Pro**, Radtel RT-910 y Bajeton BJ-7700) ahora se llaman por su nombre en la pestaña, la lectura, la escritura, la identificación y los archivos de copia.
- En la primera identificación de una de estas radios, el programa pregunta cuál es y luego lo recuerda.

### Corregido: aviso de datos no guardados
- Tras una lectura, una escritura o un envío, los canales están en la radio: al cerrar el programa el aviso solo aparece si hay cambios reales.

### Radios nuevas (controladores de CHIRP)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: los 990 canales en 10 zonas de 99, como en la radio.

Galería: 489 modelos de 44 marcas (451 con canales, 386 con ajustes, 52 con actualización del firmware). Marcas nuevas: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
