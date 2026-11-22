# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub><b>Français</b></sub></td>
<td align="center"><a href="de.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="es.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 pour Microsoft Windows 10/11

### Corrections
- **Radtel RT-880G** : elle n'est plus indiquée comme radio GMRS / FRS. C'est un émetteur-récepteur radioamateur VHF/UHF (avec GPS et APRS), comme la RT-880 et la RT-860 ; l'avertissement « Radio vendue comme GMRS » n'apparaît plus avant l'écriture.
- **Vérification après l'écriture** sur la Radtel RT-880G : elle échouait parce que la radio, qui à la fin de l'écriture enregistre les données et redémarre, était contactée trop tôt et trop souvent. Désormais le programme attend 10 secondes sans la déranger et réessaie toutes les 15 secondes, jusqu'à 90 secondes. Les mêmes délais valent pour les RT-880, RT-860 et RT-817A / RT-871A, qui utilisent le même pilote.
- Contrôles plus stricts sur les valeurs reçues de l'AR8200, des Kenwood et de la TK-180 (signalements CodeQL).

### Notes
- La vérification a été testée sur une vraie RT-880G ; les RT-880, RT-860 et RT-817A / RT-871A n'ont pas été testées en réel. Les autres radios gardent les délais d'avant.
- Aucune nouvelle radio dans cette version.
- Programme indépendant, non affilié aux constructeurs des radios.
- Suite complète de tests réussie : partie générale ; 915 comparaisons de structure et 380 lectures / écritures identiques à CHIRP.
