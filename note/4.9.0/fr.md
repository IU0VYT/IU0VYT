# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 pour Microsoft Windows 10/11

Cette version comprend aussi les nouveautés des 4.8.0, 4.7.0 et 4.6.0, publiées seulement comme versions d'essai.

## Nouveautés de la 4.9.0

### Corrigé : Identifier avec Lanchonlh HG-UV98 et Quansheng TG-UV2+
- Pendant « Identifier », le port parlait toujours en 8N1. Il utilise maintenant la parité paire (HG-UV98) et deux bits de stop (TG-UV2+), comme la lecture depuis l'onglet de la radio. Les deux radios sont maintenant reconnues.

### Corrigé : parité de la HG-UV98
- Avec la parité paire, le port de Windows contrôle maintenant aussi les erreurs de parité. Avant, le contrôle était activé puis aussitôt désactivé.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus avec les firmwares libres
- Nouveaux onglets pour les firmwares egzumer, F4HWN et LOSEHU/LTS/UVE.
- F4HWN utilise le pilote F4HWN d'armel (hors de CHIRP) : 8 niveaux de puissance, blocage de l'émission, 3 listes de balayage et les réglages du firmware.
- Chaque onglet lit et écrit seulement son firmware. Après la mise à jour du firmware, le programme indique l'onglet à utiliser.

### Nouvelles radios
- Wouxun (20 modèles) : KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (USA)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (firmware Radtel et firmware libre M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

Pour les radios issues des pilotes de CHIRP, canaux et réglages sont vérifiés par les tests automatiques face à CHIRP. Les radios GMRS, FRS et MURS sont pour les États-Unis : en Italie, l'émission n'est pas autorisée sur ces fréquences.

## Nouveautés de la 4.8.0

### Mise à jour du firmware .BTU (Radtel RT-910 et RT-900 sans Bluetooth, Bajeton BJ-7700 sans Bluetooth)
- Les firmwares .BTU se mettent à jour avec le protocole du programme « Radtel Bootload Tools ». Le chargeur est celui de la Baofeng UV-5RM, avec la présentation « PROGRAMBT80U ».
- Mode automatique, avec la radio allumée normalement. Mode manuel : radio éteinte, allumée en maintenant les deux touches latérales, jusqu'à ce que l'écran affiche « UPDATE ».
- Le chargeur ne contrôle pas le modèle. C'est pourquoi le programme reconnaît la radio et la version de l'électronique d'après le nom du fichier, affiche les avertissements et demande toujours d'écrire le modèle.
- Dans la galerie, RT-910 (sans Bluetooth, .BTU) et RT-910B (avec Bluetooth, .anc) ont maintenant deux fiches. De même pour BJ-7700 et BJ-7700 sans Bluetooth.

### Corrigé : Radtel RT-900 sans Bluetooth avec le firmware à 512 canaux
- Jusqu'à la V1.09P la radio a 512 canaux ; à partir de la V1.14P elle en a 999.
- Les canaux 513-999 ne s'écrivent plus que si vous les modifiez. Le programme vérifie d'abord dans la radio que cette zone est vide ou contient des canaux.

### Nouvelles radios (pilotes de CHIRP)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (la nouvelle version de la H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

Pour toutes, canaux et réglages sont vérifiés par les tests automatiques face à CHIRP.

## Nouveautés de la 4.7.0

### Corrigé : identification automatique bloquée ou très lente
- Une radio allumée en mode de mise à jour du firmware envoie d'elle-même des messages périodiques : avant, elle semblait répondre à chaque modèle et la recherche durait des minutes. Maintenant le programme écoute le port avant d'essayer et ne se laisse pas tromper : la recherche se termine en moins d'une minute et l'onglet Identifier explique ce qui se passe.
- L'interface ne se fige plus pendant le lancement des recherches ; les radios les plus répandues sont essayées en premier.
- Corrigé le faux avis « Onglet de la radio déjà ouvert » pour un onglet fermé auparavant.

### Radios gérées de la même façon
- La fenêtre de choix et l'onglet Identifier listent toutes les radios possibles (p. ex. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), avec le menu pour choisir et une brève indication sur l'endroit où voir la version du firmware sur la radio.

### Corrigé : Radtel RT-950 Pro / Bajeton BJ-9000 avec le firmware jusqu'à 0.23
- Le bit « émission autorisée » des canaux est maintenant toujours écrit, avec les deux versions (comme les programmes Radtel jusqu'à la v1.2.x ; la v1.3.3 l'ignore). Avec l'ancien firmware, les canaux modifiés dans l'onglet à 990 canaux pouvaient ne pas émettre. Le bit est aussi rétabli dans les canaux 1-960 qui en étaient privés.
- Vérifié les versions réelles de la Bajeton BJ-9000 : firmware V0.22, V0.27 et V0.29 (identique à celui de Radtel), programme v1.1.1 (960 canaux) et v1.3.3 (990 canaux, c'est le programme Radtel). Canaux et réglages vérifiés aussi avec le code des programmes Bajeton.

### Mise à jour du firmware pour d'autres radios
- Radtel RT-950 Pro, Radtel RT-950 et Hiroyasu HI-UV850 Pro (fichiers .BTF, même chargeur que la BJ-9000).
- Radtel RT-920 et RT-910B (fichiers .anc, même protocole que BJ-7800 et BJ-7700).

### Nouvelles radios (pilotes de CHIRP)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446) ; dans la galerie aussi Hiroyasu HI-UV850 Pro.
- Réglages aussi pour SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 et Anysecu AC-580.

## Nouveautés de la 4.6.0

### Corrigé : Radtel RT-950 Pro / Bajeton BJ-9000 avec le firmware récent
- Depuis le firmware **0.24** (décembre 2025) la radio a **990 canaux en 10 zones de 99** (ou une liste unique de 990) ; jusqu'à la **0.23** elle en avait **960 en 15 zones de 64** (ou une liste unique de 960). Le programme gère maintenant les deux versions.
- La version est reconnue d'après les données lues dans la radio (la radio ne la communique pas) : la lecture ouvre les canaux dans le bon onglet. Si les données ne suffisent pas, une note explique comment vérifier la version sur la radio.
- Les canaux 961-990 ne s'écrivent que s'ils changent, après les avoir relus dans la radio ; si la radio ne les a pas (firmware jusqu'à 0.23), rien n'est écrit.
- Canaux et réglages vérifiés avec le code d'origine des programmes Radtel v1.2.0 (960 canaux) et v1.3.3 (990 canaux).

### Nouveau : Radtel RT-950 (sans « Pro »)
- 960 canaux (15 zones de 64) ou une liste unique, réglages généraux ; vérifiée avec le code d'origine du programme « BT-9000_CPS » v1.3.0.

### Corrigé : nom exact de la radio
- Des radios différentes gérées de la même façon (p. ex. **Radtel RT-950 Pro** et **Bajeton BJ-9000**, **Baofeng K67 Pro** et **5RH Pro**, Radtel RT-910 et Bajeton BJ-7700) s'appellent maintenant par leur nom dans l'onglet, la lecture, l'écriture, l'identification et les fichiers de sauvegarde.
- À la première identification d'une de ces radios, le programme demande laquelle c'est, puis s'en souvient.

### Corrigé : avertissement de données non enregistrées
- Après une lecture, une écriture ou un transfert, les canaux sont dans la radio : en fermant le programme, l'avertissement n'apparaît que s'il y a de vraies modifications.

### Nouvelles radios (pilotes de CHIRP)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920 : les 990 canaux en 10 zones de 99, comme sur la radio.

Galerie : 489 modèles de 44 marques (451 avec canaux, 386 avec réglages, 52 avec la mise à jour du firmware). Nouvelles marques : Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
