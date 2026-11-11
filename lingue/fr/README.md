# IU0VYT Flash

<div align="center">
<table>
<tr>
<td align="center"><a href="../../README.md"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="../../README.it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub><b>Français</b></sub></td>
<td align="center"><a href="../../lingue/de/README.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="../../lingue/es/README.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="../../lingue/pt/README.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="../../lingue/ru/README.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="../../lingue/zh/README.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="../../lingue/ja/README.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="../../lingue/ko/README.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="../../lingue/ar/README.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><a href="../../lingue/ms/README.md"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub>Bahasa Melayu</sub></a></td>
<td align="center"><a href="../../lingue/hi/README.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

<p align="center">
<a href="https://github.com/IU0VYT/IU0VYT/releases/latest"><img src="https://img.shields.io/github/v/release/IU0VYT/IU0VYT?label=release&color=2ea44f" alt="Latest release"></a>
<a href="https://github.com/IU0VYT/IU0VYT/releases"><img src="https://img.shields.io/github/downloads/IU0VYT/IU0VYT/total?color=1f6feb" alt="Downloads"></a>
<a href="../../LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0--or--later-blue" alt="License GPL-3.0-or-later"></a>
<img src="https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078d6" alt="Windows 10 | 11">
<img src="https://img.shields.io/badge/languages-13-orange" alt="13 languages">
</p>

**Version 4.16.0** · Windows · interface en 13 langues (italiano, English, Français, Deutsch, Español, Português, Русский, 中文, 日本語, 한국어, العربية, Bahasa Melayu, हिन्दी) · un seul fichier exécutable, aucune installation

Programme pour radioamateurs et utilisateurs d'émetteurs-récepteurs de **47 marques** (**Baofeng**, **Bajeton**, **Quansheng**, **Retevis**, **Radtel**, **Radioddity**, **TIDRADIO**, **TYT**, **Hiroyasu**, Talkpod, Anysecu, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton, TID, Explorer, MMLradio, JJCC, Socotran, Jianpai, Boristone, HamGeek, SenhaiX, Cignus, Puxing, Feidaxin, KYD, KSUN, Lanchonlh, Boblov, Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone, Yaesu, Icom, Kenwood) : il lit et écrit canaux, fréquences et réglages par le câble de programmation, gère une archive des relais italiens et met à jour le firmware des radios prises en charge.

🌐 Site du programme : **[iu0vyt.github.io/IU0VYT](https://iu0vyt.github.io/IU0VYT/)**

*Ceci est la traduction française de [README.it.md](../../README.it.md) ; l'original italien est le texte de référence.*

## Fonctions

- **Canaux et fréquences** de 485 modèles (y compris radios PMR numériques et DMR, aussi par USB) : lecture et écriture par câble, fichiers `.img` compatibles avec CHIRP, fichiers d'origine du Baofeng K67 Pro (`.xlc`) et de la Bajeton 8300 (`.dat`).
- **Réglages des radios** (menus, écran, VOX, squelch, TOT, DTMF, touches…) pour 419 modèles, avec la description de chaque élément et un avertissement à confirmer pour les éléments sensibles.
- **Identification de la radio branchée** : le programme la reconnaît tout seul et affiche marque, modèle, type, firmware et les autres informations qu'elle communique (pour les radios DMR USB aussi ID DMR et nom). Avec l'**identification automatique** (bouton dans l'onglet « Identifier », choix mémorisé) il suffit de brancher le câble, quel que soit l'onglet : le programme écoute tous les ports des câbles ensemble, reconnaît aussi une radio allumée après le branchement du câble et plusieurs radios branchées à des câbles différents, essaie d'abord les radios déjà reconnues et les plus répandues, ne se laisse pas tromper par les radios en mode de mise à jour du firmware et ouvre l'onglet de chacune avec les canaux lus ; le cadre « Identification de la radio en cours » n'apparaît que pour le câble avec une radio qui répond ; pour une autre radio il y a le bouton **« 🔎 Identifier la radio »** en haut à droite.
- **Transfert vers une autre radio** : les canaux cochés dans une radio sont écrits directement dans une autre radio branchée, choisie par marque et modèle, avec contrôle de compatibilité, adaptation, copie de sécurité et vérification.
- **Convertisseur universel** : les canaux de n'importe quelle radio ou fichier (K67 `.xlc`, Bajeton `.dat`, images `.img` même de CHIRP, CSV de CHIRP, Excel/CSV) au format de n'importe quelle autre radio, adaptés à ses règles.
- **Interface en 13 langues** avec le sélecteur de drapeaux à côté de « ℹ Info » : traduction immédiate de tout le programme (arabe de droite à gauche), langue mémorisée au redémarrage ; fichiers d'information traduits dans le dossier `lingue`.
- **Noms des zones** au lieu de « Zone 1 », « Zone 2 »… (Radtel RT-950 Pro, RT-950, RT-920, RT-910, Bajeton BJ-9000 et BJ-7700, Hiroyasu HI-UV850 Pro ; zones des radios DMR RD-5R, DM-1801 et DM-32UV), lus et écrits par câble avec les canaux.
- **Modification groupée** des canaux cochés : puissance, largeur de bande, balayage, sous-tons, décalage ou émission ensemble, le reste de chaque canal reste tel quel.
- **Galerie de 539 modèles** de 47 marques, classés en compatibles et non compatibles, par type (radioamateur, DMR, GMRS/FRS, PMR446, pour enfants) et par marque, avec recherche, filtres et liens vers les programmes et manuels officiels.
- **Archive des relais** basée sur la liste à jour d'[IK2ANE](https://www.ik2ane.it), avec filtres et insertion dans les radios, et contrôle de compatibilité entre relais et radio avant l'écriture.
- **Mise à jour du firmware** de 57 modèles (y compris les Quansheng UV-K5, UV-K6 et UV-5R Plus, les Radtel RT-590 et RT-600 avec les fichiers .bin de Quansheng et de Radtel, les Radtel RT-470, RT-470L, RT-470X, RT-495 et RT-630 avec les fichiers .kdhx de Radtel, les TIDRADIO TD-H3, TD-H8 et Radtel RT-730 avec les fichiers .BIN du fabricant, les Retevis RA79 V3 et RA89 avec les fichiers .bin et la Retevis RT85 avec les fichiers .hex de Retevis, les Radtel RT-950 Pro et RT-950 et la Bajeton BJ-9000 avec les fichiers .BTF, les Radtel RT-920 et RT-910B avec les fichiers .anc, les Radtel RT-910 et RT-900 et la Bajeton BJ-7700 sans Bluetooth avec les fichiers .BTU, la Radtel RT-890 avec les fichiers .bin de Radtel et de M7OCM, la Baofeng DM-1701 par USB avec les fichiers .bin de Baofeng, les Baofeng RD-5R et DM-1801 par USB avec les fichiers .sgl), avec contrôle du fichier et copie de sécurité de la radio ; mise à jour guidée avec le programme officiel pour BTECH DMR-6X2 et DMR-6X2 PRO.
- **Copie de sécurité** automatique avant chaque écriture et vérification après.
- **Vérification des mises à jour** du programme à chaque démarrage.

## Utilisation

1. Téléchargez [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip) (dernière version) et extrayez l'archive.
2. Lancez `IU0VYT Flash.exe` (Windows 10/11 ; il utilise le composant WebView2 de Windows).
3. Dans l'onglet **📻 Modèles** choisissez votre radio, branchez le câble et utilisez **Lire la radio**.

Le guide complet est dans le fichier `LISEZMOI.txt` (original italien : `LEGGIMI.txt`).

## Compilation

Il faut Go 1.22 ou ultérieur. Dans le dossier `sorgenti` :

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

Les bibliothèques sont incluses dans le dossier `vendor`.

## Avertissements

- Les protocoles des radios ont été reconstitués pour l'interopérabilité et vérifiés avec des radios simulées : **conservez toujours les copies de sécurité**. Le programme est fourni sans aucune garantie.
- Programme indépendant, **ni affilié ni approuvé** par les fabricants des radios (Baofeng, Bajeton, Quansheng, Retevis, Radtel, Radioddity, TIDRADIO et les autres). Les marques citées appartiennent à leurs titulaires respectifs et servent seulement à indiquer les radios compatibles.
- Respectez la réglementation sur l'usage des fréquences de votre pays.

## Remerciements

- [CHIRP](https://chirpmyradio.com) et ses auteurs : une grande partie des pilotes des radios dérive de leur travail.
- [IK2ANE](https://www.ik2ane.it), qui met gratuitement à la disposition de tous la liste des relais italiens.

Auteurs et licences des composants : `AUTEURS_ET_LICENCES.txt`.

## Licence

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

Logiciel libre distribué sous la **GNU General Public License version 3 ou ultérieure** (`GPL-3.0-or-later`) : voir `LICENSE.txt` et `AVIS_DE_LICENCE.txt`.
