# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Página de la versión en GitHub (en inglés), con la descarga: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 para Microsoft Windows 10/11

### Novedades
- **Retevis, 24 modelos nuevos obtenidos de los programas del fabricante**: B3B, B63B, RT22P, RT622P, RB666, RB666P (misma memoria de canales que la nueva versión de la H777); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (familia de la T18); RB65, RB66, RB66P (una frecuencia y un tono por canal). La conexión, la lectura y la escritura están reescritas desde cero y comprobadas con los archivos de datos de fábrica de los programas. Las RB66 y RB66P, y las RB669, RT68H y RT668H, responden con el mismo código: la ficha correcta hay que elegirla a mano.
- **Ajustes de los 24 modelos Retevis**: 241 elementos en total (squelch, tiempo máximo de transmisión, VOX, exploración, guía de voz, ahorro de batería, pitido, teclas laterales…). Solo se ofrecen los elementos que el programa del modelo muestra; los valores de los 24 archivos de fábrica son todos opciones admitidas.
- **Contraseña de programación**: en los 24 modelos Retevis y en la nueva versión de la Retevis H777 no impide la lectura ni la escritura, como en CHIRP (el programa del fabricante, en cambio, se detiene si la contraseña no es la escrita por el usuario). Es un elemento de los ajustes, en claro: se puede ver, cambiar y quitar. La longitud y los caracteres son los del programa del fabricante: 6 caracteres (3 en RB45, RB645, RB47 y RB66; 4 en RB66P), cifras y letras (solo cifras en RB65 y RB66). En B3H, B63H, RB45, RB645, RB66 y RB666, cuyo programa no muestra la casilla de la contraseña, un aviso lo recuerda.
- **Ajustes para radios que solo tenían los canales**: familia AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 y DMR-6X2 PRO: 27 elementos, a partir de las descripciones de la memoria de qdmr y dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 y MD-9600, Retevis RT3S y RT82 (los elementos de la Baofeng DM-1701); Radtel RT-880 y RT-950 Pro con el firmware nicFW880 / nicFW950 (17 elementos, a partir de la documentación del autor); TID TD-UV68 (los elementos de la TD-H8); Radtel RT-890 con el firmware M7OCM (53 elementos).
- **Versiones de prueba**: una copia con un número más alto que la última versión estable publicada muestra una barra de aviso («Versión de prueba (beta testing)…») y, solo ella, propone la versión de prueba siguiente, descargada siempre de GitHub. Quien usa la versión estable no ve las versiones de prueba.
- Galería: 697 modelos de 54 marcas (647 con canales, 596 con ajustes).

### Notas
- Ninguna de las 24 radios Retevis se ha probado en un equipo real, ni los canales ni los ajustes: primero solo lectura, después un único ajuste inofensivo. Si la radio responde de forma inesperada, la lectura se detiene con un mensaje y no se escribe nada en la radio.
- Con una contraseña de programación ya presente el programa intenta leer y escribir igualmente, como CHIRP: no se ha probado en una radio real.
- Las RB66 y RB66P, y las RB669, RT68H y RT668H, responden con el mismo código: la ficha correcta hay que elegirla a mano.
- Programa independiente, no afiliado a los fabricantes de las radios.
- Batería completa de pruebas superada: parte general; 915 comparaciones de estructura y 380 lecturas / escrituras idénticas a CHIRP.
