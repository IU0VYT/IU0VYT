# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 para Microsoft Windows 10/11

### Novedades
- **Kenwood TK-481 y TK-981** (900 MHz, canales divididos en sistemas) y los móviles **TK-780 y TK-880**: canales y ajustes, con los controladores de CHIRP.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: ahora también los ajustes de la radio.
- **Radtel RT-13**: canales (22, sin nombre), obtenidos del programa del fabricante.
- **Radtel RT-69 GPS**: canales (999, con nombres de 12 caracteres), obtenidos del programa del fabricante; antes de escribir se comprueba que cada canal esté en las bandas de la radio.
- Más robusto con archivos dañados: corregidos dos cierres inesperados al leer los tonos de Baofeng UV-B5 / UV-B6 y UV-3R.
- Conversiones: un canal que la radio de destino no puede memorizar se salta con un aviso en lugar de detener la conversión (firmware IJV, nicFW, OpenGD77, UV-KX, R3b0rn, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Galería: 569 modelos de 47 marcas (515 con canales, 433 con ajustes).

### Notas
- TK-481 / TK-981: solo se usan los sistemas que la radio ya tiene; los de recursos compartidos (trunking) no se modifican.
- RT-13 y RT-69 GPS: si la radio tiene una contraseña, quítala antes con el programa de Radtel. Obtenidas de los programas del fabricante y probadas solo con una radio simulada: haz primero solo una lectura.
- Programa independiente, no afiliado a los fabricantes de las radios.

---

Esta versión incluye también las novedades de las versiones 4.13.0 a 4.22.0, publicadas solo como versiones de prueba. Aquí están, de la más reciente.

## IU0VYT Flash 4.22.0

### Novedades
- **Radtel RT-480 y RT-68**: canales y nombres (256 canales, numerados de 0 a 255, con nombres de 12 caracteres), obtenidos del programa del fabricante. Se modifican frecuencias, nombre, potencia, ancho de banda, escaneo y tonos; las demás opciones del canal quedan como están.
- **Radtel RT-830 y RT-850**: canales (128, sin nombre), obtenidos de los programas del fabricante. Se modifican frecuencias, potencia, ancho de banda, escaneo y tonos.
- **Radtel RT-20M**: se abre con la pestaña Ruyage UV58Plus (misma memoria que la iRadio UV-5118plus).

### Notas
- RT-480 / RT-68: solo se escriben canales y nombres; los ajustes, el VFO y el DTMF de la radio no se tocan. Frecuencias en pasos de 1,25 kHz; cada canal debe tener frecuencia de transmisión.
- RT-830 / RT-850: ajustes y VFO vuelven como estaban; la contraseña de la radio nunca se escribe. Si la radio tiene una contraseña, quítala antes con el programa de Radtel.
- Todas obtenidas de los programas del fabricante y probadas solo con una radio simulada: haz primero solo una lectura.

## IU0VYT Flash 4.21.0

### Novedades
- **Radtel RT-5D (DMR)**: canales (16 zonas de 64, con nombres de 12 caracteres), obtenidos del programa del fabricante. Se modifican frecuencias, nombre, potencia, ancho de banda, escaneo y tonos; las demás opciones del canal quedan como están.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** y los móviles **TK-7180, TK-8180** (también E): canales, con los controladores de CHIRP. La lista tiene 250 posiciones para cada una de las primeras 32 zonas; solo se usan las zonas que la radio ya tiene, 512 canales en total.

### Notas
- RT-5D: la lectura toma toda la memoria y la escritura la devuelve entera tal como está en el archivo, como hace el programa de Radtel. Antes de escribir haz una lectura reciente: un archivo antiguo devolvería los contactos y los ajustes de entonces. Probada solo con una radio simulada.
- TK-x180: solo canales, sin ajustes; las zonas no se crean ni se borran. Si la radio tiene zonas no convencionales (trunking), la lectura se rechaza.

## IU0VYT Flash 4.20.0

### Novedades
- **Radtel RT-4D y RT-6D (DMR) y RT-1000 Pro**: canales (1024, con nombres de 16 caracteres) y zonas (250), obtenidos de los programas del fabricante. Se modifican frecuencias, nombre, potencia, ancho de banda, tonos y, en RT-4D y RT-6D, el escaneo; las demás opciones del canal quedan como están. Solo se escriben canales y zonas: contactos, grupos, mensajes, radio FM y ajustes no se tocan.

### Notas
- Probadas solo con una radio simulada: haz primero solo una lectura e informa de cualquier problema.
- «Identificar» no distingue estas radios de las otras Radtel y Bajeton con la misma conexión: la pestaña hay que elegirla a mano.
- RT-1000 Pro: la exclusión del escaneo no se gestiona (queda como está en la radio).

## IU0VYT Flash 4.19.0

### Novedades
- **Radtel RT-817A / RT-871A**: canales (1024, con nombres de 16 caracteres), obtenidos del programa del fabricante. Misma conexión y mismo formato de canal que la RT-880G, sin zonas. Solo se leen y escriben las áreas que usa el programa de Radtel (ajustes, canales, radio FM); el logotipo no se toca.

### Notas
- Probada solo con una radio simulada: haz primero solo una lectura e informa de cualquier problema.
- «Identificar» no distingue la RT-817A / RT-871A de la RT-880G: la pestaña hay que elegirla a mano.

## IU0VYT Flash 4.18.0

### Novedades
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: portátiles profesionales con los canales divididos en zonas (grupos), canales y ajustes. La lista muestra las 250 posiciones de la memoria; un canal nuevo entra en la zona del canal que lo precede.
- **Quansheng UV-K5 / UV-K6 con el firmware R3b0rn de fagci**: canales (298 con la memoria original, hasta 1024 con una memoria mayor). Ajustes, VFO, presintonías y corrección SSB nunca se escriben.
- **Radtel RT-330 y RT-400**: canales, obtenidos de los programas del fabricante.

### Notas
- Los controladores de CHIRP de los portátiles TK están declarados experimentales. Las zonas no se crean ni se borran desde el programa.
- El firmware R3b0rn, la RT-330 y la RT-400 están probados solo con una radio simulada: haz primero solo una lectura e informa de cualquier problema.
- Con el firmware R3b0rn, un canal sin nombre toma como nombre la frecuencia (para ese firmware un canal sin nombre no existe).
- La pestaña RT-400 rechaza las radios que se identifican como RT-470, RT-630, RT-495 o HI-8811: úsala solo con la RT-400.

## IU0VYT Flash 4.17.0

### Novedades
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: canales leídos y escritos de uno en uno, con la radio encendida normalmente (9600 baudios). Al escribir solo se envían los canales cambiados.
- **Portátiles profesionales Kenwood TK**: canales y ajustes.
  - Serie 60: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - Serie 60G: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: se usa la pestaña «Radtel RT-880G» (misma conexión y misma memoria de canales).

### Notas
- Las Kenwood TH-F6/F7, TH-G71, TH-D7 y TH-K2 usan los comandos de los controladores de CHIRP y están probadas solo con una radio simulada: haz primero solo una lectura e informa de cualquier problema.
- Los controladores de CHIRP de los portátiles TK están declarados experimentales.
- La RT-860 sale de la comparación con el programa del fabricante y no se ha probado en una radio real.

## IU0VYT Flash 4.16.0

### Novedades
- **OpenGD77**: canales de las radios DMR con el firmware libre OpenGD77.
  - Radioddity GD-77 y GD-77S, Baofeng DM-1801 y RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 y MD-2017, Retevis RT3S y RT84.
  - La radio aparece como puerto serie USB; solo se escriben los canales y las zonas cambiados, nunca ajustes, contactos, grupos ni calibración.
- **Quansheng UV-K5, otros firmware libres**:
  - **NUNU** de kamilsss655: canales y ajustes.
  - **UV-KX** de joaquim.org: canales y ajustes; la calibración nunca se escribe.
  - La pestaña «OSFW» vale también para los firmware originales modificados con bandas ampliadas (uvmod-kitchen, colección de Tunas1337).
- **Radtel RT-880** sin GPS con el firmware original: se usa la pestaña «Radtel RT-880G».

### Notas
- OpenGD77 está escrito a partir del protocolo descrito en el programa libre qdmr y probado solo con una radio simulada: haz primero solo una lectura, conserva la copia de seguridad e informa de cualquier problema.
- NUNU y UV-KX se comparan byte a byte con los controladores para CHIRP de sus autores.

## IU0VYT Flash 4.15.0

### Novedades
- **D-STAR y C4FM**: canales por cable para los portátiles digitales de radioaficionado, con archivos .img compatibles con CHIRP. La columna «Modo» muestra «D-STAR» o «C4FM».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) e IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (la copia empieza desde la radio: las instrucciones están en la ventana del cable).
  - **Kenwood** (marca nueva, 47 en total) TH-D72, TH-D74, TH-D75, con el cable USB.
  - Ajustes para IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D y TH-D72.
- **Firmware libres**:
  - **nicFW 2.x** de la TIDRADIO TD-H3: canales y ajustes principales.
  - **nicFW880 / nicFW950** de Radtel RT-880, RT-880G y RT-950 Pro: canales.
  - **F4HWN Fusion 5.x** de Quansheng UV-K1 y UV-K5 V3: 1024 canales y ajustes.
  - Con nicFW la calibración de la radio nunca se escribe.
- **Kenwood en la galería**: también los portátiles PMR446 y DMR/NXDN, entre los no compatibles (solo se programan con el programa KPG reservado a los distribuidores).

### Correcciones
- Ajustes: un campo de texto nunca usado se lee vacío en lugar de «valor no previsto»; un valor no cambiado no se reescribe.
- Puxing PX-888K: quitadas seis opciones de los ajustes que podían escribir un dato erróneo (códigos BOT/EOT y tiempos DTMF).
- Yaesu VX-3: en los códigos DTMF ya no se admite el espacio (la radio no lo tiene).

### Notas
- Controladores escritos desde cero a partir de los de CHIRP (GNU GPL), de la documentación publicada por nicsure (nicFW) y del controlador de F4HWN, y comparados byte a byte usando radios simuladas. Las radios reales de esta versión aún no se han probado: conserva la copia de seguridad e informa de cualquier problema.
- Para Icom, Yaesu y Kenwood no hay actualización del firmware (se hace con los programas de los fabricantes).

## IU0VYT Flash 4.14.0

### Novedades
- **Yaesu e Icom**: dos marcas nuevas (46 en total) y 13 radios más, con canales por cable y archivos .img compatibles con CHIRP.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (cable SCU-35) y FT-60R/E (copia iniciada desde la radio, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (cable OPC-478).
  - Ajustes para todas salvo la IC-T70. Para estas radios no hay actualización del firmware.
- **Dos idiomas más**: malayo (Bahasa Melayu) e hindi (हिन्दी), 13 en total: programa, sitio web, guía y archivos informativos, README y política de seguridad.
- **Convertidor y «Enviar a otra radio»** hacia las Yaesu y las Icom: los tonos que la radio no tiene se quitan con un aviso y los canales que no puede memorizar se saltan con un aviso, en lugar de detener toda la operación.

### Correcciones
- Traducciones: cinco palabras en negrita de las guías seguían en italiano en todos los idiomas; dos frases equivocadas en coreano y en árabe.
- Wouxun KG-UV8D: nombres internos de los campos alineados con el CHIRP actual (sin efecto en los canales).

### Notas
- Controladores escritos desde cero a partir de los de CHIRP (GNU GPL) y comparados byte a byte con CHIRP usando radios simuladas. Las radios Yaesu e Icom reales aún no se han probado: conserva la copia de seguridad e informa de cualquier problema.
- Las traducciones al malayo y al hindi son nuevas: informa de los textos poco claros.

## IU0VYT Flash 4.13.1

### Novedades
- **Firmware IJV** para las Quansheng UV-K5, UV-K5 V3 y UV-K1 (firmware no original de IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): cinco pestañas, una por versión del firmware.
  - **V3** de 200 o 999 canales en la UV-K5 clásica (hardware V1).
  - **V4**, **V5 («FW5»)** y **V6** en la UV-K5 V3 y en la UV-K1: los 500 canales del banco elegido en la radio, con los nombres de los grupos y de los bancos.
  - Se leen y escriben canales y ajustes principales (menú, pantalla, VOX, teclas laterales). Antes de escribir, el programa comprueba que en la radio siga elegido el mismo banco y escribe solo las partes cambiadas; las calibraciones nunca se tocan.
  - **V6 (nuevo en la 4.13.1)**: enlace de datos de la V6, escritura por sectores de 4 KiB, cada uno releído y comprobado. Si la radio dice que falta el enlace de datos, actualízala con FLASH V6 de RadioManager.
- **Actualización del firmware**: un archivo .bin con una versión IJV se señala, con la pestaña que hay que usar para los canales y un aviso para V5 y V6.
- Traducciones en los 11 idiomas.

### Notas
- Protocolo y memoria obtenidos de los módulos CHIRP publicados por IJV y del análisis de RadioManager 4.3.2; código escrito desde cero, probado con radios simuladas. Las radios reales con IJV aún no se han probado: informa de cualquier problema.
- Módulos, opciones, licencia y firmware IJV siguen siendo de RadioManager de IJV.

## IU0VYT Flash 4.13.0

### Novedades
- **Firmware IJV** para las Quansheng UV-K5, UV-K5 V3 y UV-K1 (firmware no original de IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): cuatro pestañas nuevas, una por versión del firmware.
  - **V3** de 200 o 999 canales en la UV-K5 clásica (hardware V1).
  - **V4** y **V5 («FW5»)** en la UV-K5 V3 y en la UV-K1: los 500 canales del banco elegido en la radio, con los nombres de los grupos y de los bancos.
  - Se leen y escriben canales y ajustes principales (menú, pantalla, VOX, teclas laterales). Antes de escribir, el programa comprueba que en la radio siga elegido el mismo banco y escribe solo los bloques cambiados; las calibraciones nunca se tocan.
  - La **V6** se reconoce: canales y ajustes de la V6 se gestionan con RadioManager de IJV.
- **Actualización del firmware**: un archivo .bin con una versión IJV se señala, con la pestaña que hay que usar para los canales y un aviso para V5 y V6.
- Traducciones en los 11 idiomas.

### Notas
- Protocolo y memoria obtenidos de los módulos CHIRP publicados por IJV y del análisis de RadioManager 4.3.2; código escrito desde cero, probado con radios simuladas. Las radios reales con IJV aún no se han probado: informa de cualquier problema.
