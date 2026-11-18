# IU0VYT Flash 4.10.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.10.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.10.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.10.0)

IU0VYT Flash V.4.10.0 pour Microsoft Windows 10/11

### Programme en 9 langues
- Nouveau sélecteur de langue avec les drapeaux, juste à droite de « ℹ Info » : italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語.
- La traduction est immédiate, sans redémarrer : onglets, boutons, menus, fenêtres, guide, messages et avertissements (y compris ceux qui viennent des radios), fiches de la galerie, réglages et les fenêtres Windows « Ouvrir » et « Enregistrer ».
- La dernière langue choisie est mémorisée au redémarrage.
- Les fichiers d'information sont traduits dans le dossier `lingue` (un sous-dossier par langue) : guide complet, README, CHANGELOG avec les nouveautés de la 4.10, auteurs et licences, version et une note sur la licence. La licence qui fait foi reste le texte anglais de la GNU GPL (LICENSE.txt).

### Barre du haut
- Retiré le bouton « Mises à jour » : la vérification reste automatique à chaque démarrage et avec « 🔄 Vérifier les mises à jour » dans l'onglet Info.
- « 🔎 Identifier la radio » est maintenant à côté de « ? Guide ».

### Mise à jour du firmware : 5 radios de plus (57 en tout)
- Radtel RT-890 : fichiers .bin de Radtel et firmwares libres (M7OCM, OEFW) ; après le firmware M7OCM, les canaux se gèrent avec son onglet.
- Baofeng DM-1701 / Retevis RT84 par USB, avec les fichiers .bin de Baofeng : seule la mémoire du programme est écrite, le chargeur de la radio n'est pas touché.
- Baofeng RD-5R et DM-1801 par USB (aucun pilote), avec les fichiers .sgl du constructeur : la radio contrôle le modèle avant d'effacer la mémoire.

### Réglages : 10 radios de plus (396 en tout)
- Feidaxin FD-288A, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH et Retevis RT76P (depuis CHIRP).
- Baofeng RD-5R et DM-1801 (31 éléments), Baofeng DM-1701 (nom, ID DMR, textes d'allumage).

Les protocoles sont réécrits de zéro et testés avec des radios simulées et avec les vrais fichiers des constructeurs ; ils n'ont pas été testés avec de vraies radios : conservez toujours les copies de sauvegarde.

Galerie : 489 modèles de 44 marques (451 avec canaux, 396 avec réglages, 57 avec la mise à jour du firmware).
