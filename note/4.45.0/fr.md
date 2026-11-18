# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0 pour Microsoft Windows 10/11

### Nouveautés
- **Radios DMR de la famille AnyTone** : AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus et BTECH DMR-6X2 et DMR-6X2 PRO (ces deux-là étaient « non compatibles »). Canaux et zones avec le câble USB de la radio ; mémoire et protocole du programme libre dmrconfig, vérifiés avec son code d'origine. Seuls les canaux, les zones et les listes de balayage sont écrits.
- **TYT MD-UV380, MD-UV390, MD-2017, MD-9600 et Retevis RT3S et RT82** avec le firmware d'origine : même mémoire et même protocole que la Baofeng DM-1701.
- **Baofeng BF-666S** (pilote de la BF-888S) et **BF-T6** (expérimentale, pilote de la BF-T8).
- Pour les fiches qui restent « non compatibles » UV-8D, BF-UV8, DM-1702 et DM-1703, la raison est maintenant écrite.
- Galerie : 673 modèles de 54 marques (623 avec canaux, 556 avec réglages).

### Notes
- Aucune des nouvelles radios n'a été testée sur une vraie radio : faites d'abord une simple lecture.
- BTECH DMR-6X2 PRO et Baofeng BF-T6 sont expérimentales.
- Programme indépendant, non affilié aux constructeurs des radios.
- Suite complète de tests réussie : partie générale ; 918 comparaisons de structure et 381 lectures / écritures identiques à CHIRP ; famille AnyTone comparée avec le code d'origine de dmrconfig.

---

Cette version comprend aussi les nouveautés des versions 4.24.0 à 4.44.0, qui n'ont pas eu de version stable à elles. Les voici, de la plus récente.

## IU0VYT Flash 4.44.0

### Nouveautés
- **Réglages** pour huit autres radios qui n'avaient que les canaux : Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro et Bajeton DM-5000, d'après les programmes des constructeurs.
- **RT-480 et RT-68** : à l'écriture, le bloc des réglages généraux arrive maintenant aussi à la radio.
- **RT-4D, RT-6D, RT-1000 Pro** : le fichier de la radio contient aussi les réglages (les fichiers enregistrés avant restent valables).

### Notes
- Aucun des nouveaux réglages n'a été testé sur une vraie radio.
- RT-69 GPS : pour l'instant seulement les éléments vérifiés un par un sur le programme du constructeur.

## IU0VYT Flash 4.43.0

### Nouveautés
- **Erreur corrigée** : pour 33 radios (dont Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22 et HA1G, Kenwood TK-2140 / TK-3140) et pour les 8 « jumelles » de la 4.42, les réglages n'apparaissaient pas après la lecture par câble, mais seulement en ouvrant un fichier de CHIRP.
- **Réglages** pour cinq autres Radtel : RT-817A / RT-871A, RT-830, RT-850, RT-330 et RT-400, d'après les programmes du constructeur.

### Notes
- Aucun des nouveaux réglages n'a été testé sur une vraie radio.
- Radtel RT-830 et RT-850 : un élément par radio est lu par le programme du constructeur autrement qu'il n'est écrit ; ici on suit l'écriture et l'élément porte un avertissement.

## IU0VYT Flash 4.42.0

### Nouveautés
- **Onglet Accueil** : il tient maintenant en entier dans la fenêtre. Le cadre des marques est plus large et la page s'adapte toute seule à la taille de la fenêtre.
- **Réglages** pour 30 modèles qui n'avaient que les canaux : Baofeng série D (BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro et Radtel RT-13.

### Notes
- Aucun des nouveaux réglages n'a été testé sur une vraie radio.
- Baofeng série D : dans le programme du constructeur, « guide vocal » et « bip des touches » sont inversés entre lecture et écriture ; ici on suit l'écriture et les deux éléments portent un avertissement.

## IU0VYT Flash 4.41.0

### Nouveautés
- **Nouvelles radios** (les portatifs de CHIRP qui manquaient) : Yaesu VX-1R (canaux et réglages), Icom IC-91AD et IC-92AD (deux bandes, D-STAR aussi), BTECH UV-PRO (par Bluetooth, sans câble) et le récepteur AOR AR8200 (nouvelle marque).
- **Réglages** pour des radios qui n'avaient que les canaux : Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### Notes
- Aucune des nouvelles radios n'a été testée en vrai : faites d'abord une simple lecture.
- BTECH UV-PRO : l'écriture n'est permise qu'avec la version du firmware testée par l'auteur du pilote de CHIRP (147).

## IU0VYT Flash 4.40.0

### Nouveautés
- **Recherche automatique plus rapide avec l'onglet déjà ouvert** : le contrôle court (auparavant seulement pour la Baofeng K67 Pro) vaut maintenant aussi pour les Icom, les Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, les Quansheng UV-K5 et dérivées, les Bajeton BJ-9000, BJ-7800, BJ-7700 et BJ-8300, les Radtel RT-900 / RT-950 / RT-470 et les radios à la même mémoire. Pour les autres, la lecture complète reste.
- **Nouvelle marque : Alinco.** Portatifs DJ-596, DJ-175, DJ-G7EG et DJ-G7T ; mobiles DR-135T, DR-235T, DR-435T, DR-03T et DR-06T. Canaux seulement. Fonctionne aussi avec les câbles qui ne renvoient pas les données.

### Notes
- Aucune Alinco n'a été testée en vrai : faites d'abord une simple lecture.
- Le contrôle court est testé avec les radios simulées ; si après le contrôle une radio ne revient pas au fonctionnement normal, signalez-le.

## IU0VYT Flash 4.39.0

### Nouveautés
- **Quatre nouvelles radios Icom** : ID-80H (portatif D-STAR), ID-880H (mobile D-STAR), IC-W32A et IC-W32E (portatifs bibandes avec deux moitiés séparées : de 0 à 99 les canaux VHF, de 100 à 199 les canaux UHF). Canaux seulement.

### Notes
- Aucune des nouvelles radios n'a été testée en vrai : faites d'abord une simple lecture.

## IU0VYT Flash 4.38.0

### Corrections
- **Identification** : la Bajeton BJ-9000 se bloquait de temps en temps, surtout à la première identification. Le programme écoute de nouveau le port pendant une seconde avant le premier salut, comme dans la 4.35, aussi pour la radio déjà reconnue sur ce câble. Les autres améliorations de la 4.36 restent.

### Notes
- Aucune nouvelle radio dans cette version.

## IU0VYT Flash 4.37.0

### Nouveautés
- **Huit nouveaux portatifs Yaesu** : VX-5R et VX-7R (canaux) ; VX-2R, FT-50R, VX-150, VX-8R, VX-8DR et VX-8GE (canaux et réglages ; pour les VX-8 aussi l'APRS).
- La copie part de la radio (mode « CLONE ») : l'onglet se choisit à la main et les instructions sont dans la fenêtre du câble.

### Notes
- Aucune des nouvelles radios n'a été testée en vrai : faites d'abord une simple lecture.
- Non fait : Yaesu VX-1.

## IU0VYT Flash 4.36.0

### Corrections
- **« Identifier » et identification automatique** : le programme se souvient sur quel câble il a reconnu chaque radio et commence par celles-là ; la première recherche ne perd plus de secondes en écoute et en préparation ; la radio allumée après le câble est réessayée plus souvent.
- La fenêtre « Identification de la radio en cours » n'apparaît que pendant que la radio se laisse lire, avec le modèle et l'avancement. Si la radio est connectée mais éteinte, le programme demande de l'allumer et la reconnaît dès qu'elle est allumée.
- Baofeng K67 Pro avec l'onglet déjà ouvert : reconnue en quelques secondes, sans relire toute la mémoire.

### Nouveautés
- Onglet « Identifier » : bouton **« Enregistrer le diagnostic de l'identification »** (temps et premiers octets des dernières recherches dans un fichier texte, à envoyer à l'auteur si l'identification est lente ou se trompe ; il ne contient ni canaux ni données personnelles).

### Notes
- Aucune nouvelle radio dans cette version.

## IU0VYT Flash 4.35.0

### Corrections
- **« Identifier » et identification automatique** : retiré le « tour rapide » de la 4.31, qui avec de vraies radios pouvait laisser la radio bloquée en programmation et rendre la recherche lente. Les attentes sont de nouveau celles d'avant.
- La recherche essaie les mêmes groupes de radios sur chaque PC (avant, ils changeaient d'un démarrage à l'autre).
- L'écho d'un câble sans radio n'est plus pris pour une radio.

### Notes
- Aucune nouvelle radio dans cette version.

## IU0VYT Flash 4.34.0

### Nouveautés
- **Nouvelles radios** : Yaesu VX-170 et VX-177 (portatifs), FT-7800 / FT-7900 et FT-8900 (mobiles), canaux seulement, avec les pilotes de CHIRP : lecture, écriture et octets échangés avec la radio identiques à CHIRP.

### Notes
- La copie part de la radio (mode « CLONE ») : l'onglet doit être choisi à la main, « Identifier » n'essaie pas ces radios.

## IU0VYT Flash 4.33.0

### Nouveautés
- **Nouvelles radios** : Icom IC-T7H, IC-T8A, IC-Q7A et IC-P7 (canaux), avec les pilotes de CHIRP : lecture, écriture et octets échangés avec la radio identiques à CHIRP.

### Notes
- IC-Q7A et IC-P7 : le mode (FM, FM large, AM) suit la fréquence ; le saut « P » et le pas d'accord du canal sont conservés.

## IU0VYT Flash 4.32.0

### Nouveautés
- **Nouvelles radios** : Baofeng T22 en version FRS, TIDRADIO TD-M8S, TD-V2, TD-H4 et TD-H11 (canaux et réglages), d'après les programmes des constructeurs.
- **« Identifier »** propose aussi les nouvelles radios qui se présentent comme la BF-888S, sans allonger la recherche : la radio n'est lue avec leur pilote que quand on ouvre cet onglet.

### Corrections
- TIDRADIO de 2026 (TD-H3 Plus / SE, TD-H7 Plus, TD-H8 4e génération, TD-H9, TD-638UV) : la puissance moyenne n'est proposée que par les modèles qui l'ont vraiment, et jamais dans les versions GMRS.
- Baofeng T22 PMR446 : connexion comme le programme du constructeur.
- Notes des Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS et TIDRADIO TD-M2 Pro : « Identifier » les essaie, la phrase qui disait le contraire a été retirée.

## IU0VYT Flash 4.31.0

### Nouveautés
- **« Identifier » est beaucoup plus rapide** : il fait d'abord un tour rapide de toutes les radios et seulement ensuite, s'il n'a rien trouvé, le tour complet d'avant. Une radio en bas de la liste est trouvée en moins d'une demi-minute au lieu de presque deux.
- **« Identifier » essaie maintenant aussi les radios ajoutées dans les dernières versions** : Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G et les TIDRADIO sans écran.

### Corrections
- Retevis RB39 / RB639 : avec l'onglet du mauvais pays, la radio restait en programmation jusqu'à son extinction.

## IU0VYT Flash 4.30.0

### Nouveautés
- **Réglages généraux** pour les TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 de quatrième génération**, **TD-H9** et **TD-638UV** : touches programmables, squelch, VOX, durée maximale d'émission, affichage, langue, messages d'allumage et autres.
- **Nouvelles radios (canaux et réglages)** : TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR** et **TD-M2 Pro**.

### Notes
- Les nouvelles radios viennent des programmes du constructeur et n'ont pas encore été testées avec une vraie radio : faites d'abord une simple lecture. « Identifier » ne les distingue pas entre elles : choisissez vous-même l'onglet.

## IU0VYT Flash 4.29.0

### Nouveautés
- **Nouvelles radios (canaux seulement)** : TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 de quatrième génération**, **TD-H9** et **TD-638UV**, chacune dans les versions de base, HAM et GMRS.
- La TD-H3 Plus se présente comme la TD-H3 : utilisez les onglets « TD-H3 Plus », car elle a trois puissances.

### Notes
- Ces radios viennent du programme du constructeur de 2026 (« TIDRadioCPS ») et n'ont pas encore été testées avec une vraie radio : faites d'abord une simple lecture. Les réglages généraux ne sont pas encore gérés.

## IU0VYT Flash 4.28.0

### Nouveautés
- **Nouvelles radios** (mobiles Radioddity) : **UB15-G** (GMRS 15 W), **DB35-G** (GMRS 35 W), **DB35-UV** (VHF / UHF 35 W) et **DB40-G** (GMRS 40 W, 199 canaux), avec canaux et réglages généraux. La UB15-G a deux onglets, selon le firmware : avant la V1.10 (sans noms) et V1.10 (avec les noms).
- La **DB40-G** est aussi dans CHIRP : ici elle lit et écrit les mêmes octets et les mêmes fichiers `.img`.

### Corrections
- Les tests ne se compilaient plus sous Windows depuis la 4.26.0 (trois fichiers de test utilisaient les radios simulées, qui ne sont pas compilées sous Windows). Le programme n'était pas touché.

### Notes
- UB15-G, DB35-G et DB35-UV viennent du programme du constructeur et de son fichier de configuration et n'ont pas encore été testées avec une vraie radio : faites d'abord une simple lecture. « Identifier » ne reconnaît pas ces radios : l'onglet doit être choisi à la main.

## IU0VYT Flash 4.27.0

### Nouveautés
- **Nouvelles radios** : **Retevis C6** (FRS / GMRS, 30 canaux), **Retevis C66** (UHF, 16 canaux) et **Radioddity Echo-DE / Echo-IT** (PMR446, 16 canaux), avec canaux et réglages généraux.
- Avant d'écrire, le programme vérifie que la gamme, le modèle et la zone de la radio sont ceux de la mémoire à écrire ; mot de passe et étalonnages de la radio ne sont jamais écrits.

### Corrections
- **Retevis RB39 / RB39P** : elle est maintenant parmi les radios « GMRS / FRS (USA) » dans l'onglet Modèles.

### Notes
- Les nouvelles radios viennent du code des programmes des constructeurs et n'ont pas encore été testées avec une vraie radio : faites d'abord une simple lecture. « Identifier » ne les reconnaît pas : l'onglet doit être choisi à la main.
- Si un mot de passe de lecture est défini dans la radio, retirez-le d'abord avec le programme du constructeur.

## IU0VYT Flash 4.26.0

### Nouveautés
- **Nouvelles radios** : **Retevis RB39 / RB39P** (FRS, 22 canaux) et **RB639 / RB639P** (PMR446, 16 canaux), avec canaux et réglages généraux. La radio déclare pour quel pays elle est faite : avec le mauvais onglet, la lecture s'arrête et indique le bon.
- **Nouveaux réglages** pour **Baofeng BF-T17 / BF-T18** (PMR446 et FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3** et **PR-T3** : squelch, durée maximale d'émission, VOX, guide vocal, bip des touches, économie de batterie et les autres du programme du constructeur.

### Corrections
- Journaux des modifications dans les 12 langues : date de la 4.25.0 corrigée (12/11/2026).

### Notes
- Les nouvelles radios viennent du code des programmes des constructeurs et n'ont pas encore été testées avec une vraie radio : faites d'abord une simple lecture. « Identifier » ne les reconnaît pas : l'onglet doit être choisi à la main.

## IU0VYT Flash 4.25.0

### Nouveautés
- **Bajeton BJ-8300 avec le nouveau firmware** : nouvel onglet pour les radios à 990 canaux en 10 zones de 99 (programme du constructeur BT-11_CPS v1.2.6) : canaux, noms des zones et réglages. L'onglet « Bajeton 8300 » de toujours (960 canaux, 15 zones de 64) reste tel quel et maintenant **avertit** quand il trouve le nouveau firmware.
- **Nouvelles radios** : **Radioddity GS-10B Pro** (999 canaux et réglages), **Radioddity FS-T3** et **PR-T3** (canaux).
- **Bajeton BJ-9000** à 990 canaux : cinq réglages de plus (ceux du programme v1.3.3).

### Corrections
- **Baofeng BF-T17 / BF-T18, GT-18 et Radioddity PR-T5** : puissance haute et basse étaient inversées. Si vous avez écrit ces radios avec une version précédente, relisez-les et choisissez de nouveau la puissance.
- « Sortie automatique du menu » (BJ-8300, BJ-9000, Radtel RT-950) : le dernier choix est 60 secondes, pas 55.

### Notes
- Les nouvelles radios et le nouvel onglet viennent du code des programmes des constructeurs et n'ont pas encore été testés avec une vraie radio : faites d'abord une simple lecture.
- BJ-8300 : pour savoir quel onglet utiliser, regardez sur la radio combien il y a de zones (15 de 64 canaux, ou 10 de 99).

## IU0VYT Flash 4.24.0

### Nouveautés
- **Onglet Modèles plus lisible** : les étiquettes des types de radio sont au centre de la page et la ligne entre radios compatibles et non compatibles est rouge et plus épaisse.
- **26 radios de plus**, gérées avec les pilotes de radios identiques déjà présentes :
  - reconnues en comparant les programmes des constructeurs : **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28** ;
  - d'après le tableau des modèles équivalents de CHIRP : Radioddity GA-5S, Retevis RT5 (trois puissances), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+ et TS-T9+, Rugged RH5R et RH5X, Ansoko A-5R et A-8S, Arcshell AR-5, AR-6 et AR-8, Greaval GV-8S et GV-9S, Tenway TW-325, UV-82 Pro et UV-5R Pro.

### Notes
- Les six radios reconnues en comparant les programmes des constructeurs n'ont pas encore été testées avec une vraie radio : faites d'abord une simple lecture. Si le modèle n'est pas le bon, la radio ne répond pas et n'est pas modifiée.
