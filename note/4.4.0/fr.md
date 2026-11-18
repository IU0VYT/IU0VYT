# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 pour Microsoft Windows 10/11

Cette version comprend aussi les nouveautés des 4.3.0, 4.2.0, 4.1.0 et 4.0.0, qui n'ont pas eu de version stable à elles.

## Nouveautés de la 4.4.0

### Corrigé : fenêtre « Identification de la radio en cours » sans radio
- La fenêtre pouvait apparaître même avec un câble sans radio : les parasites de la ligne et l'écho des commandes envoyées ressemblaient à une radio qui répond.
- Elle n'apparaît maintenant **que pour le câble relié à une radio qui répond vraiment**. Avec deux câbles et une seule radio (COM6 sans radio, COM7 avec la radio, ou l'inverse), elle n'apparaît que pour le port de la radio.
- Sans radio, le tour d'essais est aussi plus rapide.

### Barre du haut, accueil, onglet Identifier
- Le bouton s'appelle **« 🔎 Identifier la radio »** et se trouve à droite, à côté de « Mises à jour ».
- Avec beaucoup d'onglets ouverts, le bouton se rétrécit tout seul, jusqu'à la seule loupe.
- Accueil : les descriptions des cadres de la 4.2 sont de retour. En bas, le copyright, le site et les sources restent visibles, même en 1280×720.
- Onglet Identifier : explications courtes et claires.

### Firmware Retevis (43 modèles avec la mise à jour du firmware)
- **Retevis RA79 V3** (et UV-K5 V3 / UV-K1, chargeur 7.x) avec les fichiers .bin de Retevis, par exemple « RA79 V3.bin ». Le firmware n'est envoyé que si le chargeur de la radio est de la même famille.
- **Retevis RA89** avec les fichiers .bin de Retevis, par exemple « Eng_RA89_Upadat_20250524_B91.38.bin ». Il utilise le protocole du programme de Retevis « RA89-update-software », qui a été exécuté avec une radio simulée. Les commandes envoyées sont identiques octet par octet.

### Autres
- Les tests avec les radios simulées n'empêchent plus la compilation des tests sous Windows.

Galerie : 386 modèles de 27 marques (348 avec canaux, 302 avec réglages, 43 avec la mise à jour du firmware).

## Nouveautés de la 4.3.0

### Identification automatique : plus rapide, plus de radios, plus de ports
- Un tour complet d'essais sans radio dure environ **50 secondes** (135 auparavant).
- Le programme **écoute en continu tous les ports des câbles**, chacun pour soi et tous ensemble :
  - il reconnaît aussi une radio connectée éteinte puis allumée plus tard ;
  - il reconnaît **plusieurs radios sur des câbles différents** en même temps, en affiche les données et en ouvre tous les onglets.
- Les **radios déjà reconnues** (p. ex. K67 Pro, Bajeton 8300, BJ-9000) sont essayées en premier, puis toutes les autres.
- Si l'onglet de la radio est déjà ouvert, l'avis « Radio déjà identifiée !! Onglet de la radio déjà ouvert » apparaît (avec plusieurs radios « Radios déjà identifiées !! Onglets des radios déjà ouverts »).
- Onglet Identifier : ports dans l'ordre (COM6, COM7… COM10) et écrits sous la forme « COM6 — USB-SERIAL CH340 » ; « Identifier automatiquement une nouvelle radio » en gras.

### Mise à jour du firmware : 8 radios de plus (41 en tout)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** avec les fichiers .kdhx de Radtel (chargeur « JJCC Bootloader Pro »). RT-470 : le firmware 1.x est seulement pour l'ancienne carte, le 2.x seulement pour la nouvelle. Le programme l'indique et demande une confirmation écrite.
- **TIDRADIO TD-H3, TD-H8 et Radtel RT-730** avec les fichiers .BIN du constructeur : lancez la mise à jour et allumez la radio en maintenant le PTT.

### Autres
- **Bajeton BJ-9000** : comme la BJ-8300, les 960 canaux dans une liste unique sans zones (« Canaux : tous (960) »).
- **Accueil** plus compact : en bas, version, copyright, site et sources sont toujours visibles.
- **Galerie** : une ligne épaisse sépare les modèles compatibles des non compatibles.
- **LEGGIMI** : retiré une longue partie répétée par erreur dans la 4.2.0.

Galerie : 386 modèles de 27 marques (348 avec canaux, 302 avec réglages, 41 avec la mise à jour du firmware).

## Nouveautés de la 4.2.0

- Identification automatique corrigée : la fenêtre « Identification de la radio en cours » n'apparaît que lorsqu'une radio connectée répond (plus au démarrage avec le seul câble) et disparaît quand la radio est reconnue ou n'est pas trouvée. Après la reconnaissance, les données et l'onglet de la radio s'ouvrent et la recherche s'arrête, sans désactiver la fonction.
- Nouveau bouton « 🔎 Identifier automatiquement une nouvelle radio » dans la barre du haut, entre « Info » et « Mises à jour » : il cherche une nouvelle radio avec les mêmes règles et est grisé si l'identification automatique est désactivée.
- 35 nouvelles radios (36 versions) et 8 marques de plus : famille TYT TH-UV88 (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), famille Radtel RT-490 (avec JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 radios de la famille Retevis RT21 (avec Abbree AR-63) et Quansheng TK11.
- Mise à jour du firmware aussi pour Radtel RT-590 / RT-590 Plus et RT-600 / RT-600 Plus avec les fichiers .bin de Radtel.
- Corrigé quelques défauts cachés dans la lecture des noms de canaux avec des lettres accentuées ou chinoises.
- Galerie de 386 modèles de 27 marques : 348 avec canaux et relais, 302 avec les réglages, 33 avec la mise à jour du firmware.

## Nouveautés de la 4.1.0

- Identification automatique corrigée : elle fonctionne dans n'importe quel onglet et, quand vous débranchez la radio, l'éteignez ou la changez, la recherche repart toute seule et reconnaît la radio connectée ensuite (même si c'est la même ou si vous l'allumez après le câble). Avant, après la première reconnaissance, elle ne cherchait plus d'autres radios.
- Pendant la recherche automatique apparaît la fenêtre « Identification de la radio en cours. Veuillez patienter. », avec le bouton « Interrompre ».
- 14 nouvelles radios et 3 marques de plus : Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 et RT-730, Hiroyasu HI-8811, Talkpod A36plus et A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (2e et 3e génération) et TD-H3, dans les versions de base, HAM et GMRS.
- Mise à jour du firmware des Quansheng UV-K5, UV-K5(8), UV-K6 et UV-5R Plus avec les fichiers .bin de Quansheng, avec le contrôle que le firmware correspond au type de radio connectée.
- Écran d'accueil et site : « GESTION DES CANAUX - FRÉQUENCES - RELAIS - FIRMWARE » et les noms de toutes les marques dans un cadre façon jeu vidéo des années 80.
- Galerie de 349 modèles de 19 marques : 311 avec canaux et relais, 265 avec les réglages, 31 avec la mise à jour du firmware.

## Nouveautés de la 4.0.0

- Programme multimarque : 335 modèles de 16 marques dans la galerie, 297 avec canaux et fréquences par câble et 251 avec les réglages de la radio.
- Nouvelles radios : Quansheng UV-K5, UV-K6, UV-5R Plus et UV-K1 ; 31 modèles Retevis (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…) ; Radtel RT-900, RT-900 BT, RT-920 et T18 ; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S et GA-510 ; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton et TID.
- Identification automatique (onglet « Identifier ») : vous branchez le câble et la radio est reconnue toute seule, avec son onglet ouvert et les canaux lus. Le choix est mémorisé et se désactive avec le même bouton.
- Galerie des modèles avec le menu des marques et une couleur par marque.
- « Lire la radio » et « Écrire dans la radio » en bleu dans les onglets des radios.
- Accueil, Info, guide, site et documentation mis à jour pour le programme multimarque.
