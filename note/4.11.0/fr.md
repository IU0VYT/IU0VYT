# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 pour Microsoft Windows 10/11

### Noms des zones
- À la place de « Zone 1 », « Zone 2 »… chaque zone peut avoir un nom, comme dans les programmes du constructeur : bouton « ✎ Noms des zones… » ou double-clic sur la zone. La barre des zones et la ligne du bas affichent « numéro · nom ».
- Radios : Radtel RT-950 Pro, Bajeton BJ-9000 et Hiroyasu HI-UV850 Pro (firmware jusqu'à 0.23 et à partir de 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 sans Bluetooth ; zones des radios DMR Baofeng RD-5R, DM-1801 et DM-32UV.
- Les noms se lisent et s'écrivent par câble avec les canaux et s'enregistrent dans le fichier .img. Longueur et caractères comme dans les programmes du constructeur (12 caractères, chinois compris, sur la RT-950 Pro ; 10 sur la RT-920 ; 16 sur les DMR).
- Sécurité : la zone des noms n'est écrite que si les noms changent, après l'avoir relue dans la radio ; si la radio contient à cet endroit des données qui ne sont pas des noms, rien n'est écrit.

### Modification groupée
- « ✎ Modification groupée… » (aussi par le clic droit) change en une fois la puissance, la largeur de bande, le balayage, la tonalité TX, la tonalité RX, le décalage et l'émission des canaux cochés (ou sélectionnés). Les éléments laissés sur « ne pas changer » restent tels quels ; Ctrl+Z annule tout.

### Deux langues de plus (11 en tout)
- Coréen, avec le drapeau de la Corée du Sud.
- Arabe, avec le drapeau des Émirats arabes unis et la page de droite à gauche (les fréquences et les nombres restent de gauche à droite).
- Tous les fichiers d'information du dossier `lingue` sont traduits aussi en coréen et en arabe et mis à jour pour la 4.11 dans toutes les langues.

### Autres
- Site : les langues disponibles sont indiquées ; la liste des fichiers de firmware a été retirée.
- `go test` passe sous Windows (i18n_test.go compris) ; un nouveau test vérifie que LEGGIMI.txt, AUTORI_E_LICENZE.txt et la licence sont identiques dans le dossier principal et dans « sorgenti ».

Les nouveaux formats sont tirés du code des programmes du constructeur et du pilote de CHIRP et testés avec des radios simulées ; ils n'ont pas été testés avec de vraies radios : conservez toujours les copies de sauvegarde.

Galerie : 489 modèles de 44 marques (451 avec canaux, 396 avec réglages, 57 avec la mise à jour du firmware).
