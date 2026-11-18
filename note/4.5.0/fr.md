# IU0VYT Flash 4.5.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.5.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.5.0)

IU0VYT Flash V.4.5.0 pour Microsoft Windows 10/11

### Firmware Retevis RT85 (44 modèles avec la mise à jour du firmware)
- **Retevis RT85** avec les fichiers .hex de Retevis, par exemple « RT85-Firmware-V1.53.hex ». Il utilise le protocole du programme de Retevis « RT85 10W_5W_UpData », qui a été exécuté avec une radio simulée : les blocs et les commandes envoyés sont identiques octet par octet.
- Procédure de Retevis : radio éteinte, allumée en maintenant **PTT et P1** (voyant vert). On choisit ensuite la puissance de la radio (**5 W ou 10 W**, comme dans le programme de Retevis) et on écrit RT85 pour confirmer.

### Corrigé : mise à jour de la Retevis RA89
- La 4.4 attendait de la radio une mauvaise réponse à la présentation : avec une vraie RA89, la mise à jour ne démarrait pas (la radio n'était pas touchée). La réponse attendue et les pauses sont maintenant celles du programme de Retevis.

### Ports, onglets Identifier et Convertir
- Les ports des câbles sont toujours dans l'ordre croissant, sur tous les systèmes : COM1, COM2, COM3… COM10. L'identification automatique part aussi du premier port.
- Onglet Identifier : « Les radios déjà identifiées par le passé sont reconnues plus vite ».
- Onglet Convertir : « Choisissez un fichier Baofeng, Bajeton, un .img de toutes les autres marques de radio (y compris enregistré avec CHIRP), un CSV de CHIRP, des tableaux Excel ou CSV ».

### Site web
- Retiré le texte des notes de version sous le bouton de téléchargement.
- Le titre est plus haut : le numéro de version sous « PRESS START » se voit en entier même sur les écrans des portables.

Galerie : 386 modèles de 27 marques (348 avec canaux, 302 avec réglages, 44 avec la mise à jour du firmware).
