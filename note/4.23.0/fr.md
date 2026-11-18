# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 pour Microsoft Windows 10/11

### Nouveautés
- **Kenwood TK-481 et TK-981** (900 MHz, canaux répartis en systèmes) et les mobiles **TK-780 et TK-880** : canaux et réglages, avec les pilotes de CHIRP.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180** : maintenant aussi les réglages de la radio.
- **Radtel RT-13** : canaux (22, sans nom), tirés du programme du constructeur.
- **Radtel RT-69 GPS** : canaux (999, avec des noms de 12 caractères), tirés du programme du constructeur ; avant d'écrire, on vérifie que chaque canal est dans les bandes de la radio.
- Plus robuste face aux fichiers abîmés : corrigé deux plantages à la lecture des tonalités des Baofeng UV-B5 / UV-B6 et UV-3R.
- Conversions : un canal que la radio de destination ne peut pas mémoriser est sauté avec un avertissement au lieu d'arrêter la conversion (firmwares IJV, nicFW, OpenGD77, UV-KX, R3b0rn, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Galerie : 569 modèles de 47 marques (515 avec canaux, 433 avec réglages).

### Notes
- TK-481 / TK-981 : seuls les systèmes que la radio a déjà sont utilisés ; ceux à ressources partagées (trunking) ne sont pas modifiés.
- RT-13 et RT-69 GPS : si un mot de passe est défini sur la radio, retirez-le d'abord avec le programme de Radtel. Tirées des programmes du constructeur et testées seulement avec une radio simulée : faites d'abord une simple lecture.
- Programme indépendant, non affilié aux constructeurs des radios.

---

Cette version comprend aussi les nouveautés des versions 4.13.0 à 4.22.0, publiées seulement comme versions d'essai. Les voici, de la plus récente.

## IU0VYT Flash 4.22.0

### Nouveautés
- **Radtel RT-480 et RT-68** : canaux et noms (256 canaux, numérotés de 0 à 255, avec des noms de 12 caractères), tirés du programme du constructeur. On modifie fréquences, nom, puissance, largeur de bande, balayage et tonalités ; les autres éléments du canal restent tels quels.
- **Radtel RT-830 et RT-850** : canaux (128, sans nom), tirés des programmes du constructeur. On modifie fréquences, puissance, largeur de bande, balayage et tonalités.
- **Radtel RT-20M** : s'ouvre avec l'onglet Ruyage UV58Plus (même mémoire que l'iRadio UV-5118plus).

### Notes
- RT-480 / RT-68 : seuls les canaux et les noms sont écrits ; réglages, VFO et DTMF de la radio ne sont pas touchés. Fréquences par pas de 1,25 kHz ; chaque canal doit avoir une fréquence d'émission.
- RT-830 / RT-850 : réglages et VFO reviennent tels quels ; le mot de passe de la radio n'est jamais écrit. Si un mot de passe est défini sur la radio, retirez-le d'abord avec le programme de Radtel.
- Toutes tirées des programmes du constructeur et testées seulement avec une radio simulée : faites d'abord une simple lecture.

## IU0VYT Flash 4.21.0

### Nouveautés
- **Radtel RT-5D (DMR)** : canaux (16 zones de 64, avec des noms de 12 caractères), tirés du programme du constructeur. On modifie fréquences, nom, puissance, largeur de bande, balayage et tonalités ; les autres éléments du canal restent tels quels.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** et les mobiles **TK-7180, TK-8180** (E aussi) : canaux, avec les pilotes de CHIRP. La liste a 250 positions pour chacune des 32 premières zones ; seules les zones que la radio a déjà sont utilisées, 512 canaux en tout.

### Notes
- RT-5D : la lecture prend toute la mémoire et l'écriture la renvoie entière telle qu'elle est dans le fichier, comme le fait le programme de Radtel. Avant d'écrire, faites une lecture récente : un ancien fichier remettrait les contacts et les réglages de l'époque. Testée seulement avec une radio simulée.
- TK-x180 : canaux seulement, pas de réglages ; les zones ne se créent ni ne se suppriment. Si la radio a des zones non conventionnelles (trunking), la lecture est refusée.

## IU0VYT Flash 4.20.0

### Nouveautés
- **Radtel RT-4D et RT-6D (DMR) et RT-1000 Pro** : canaux (1024, avec des noms de 16 caractères) et zones (250), tirés des programmes du constructeur. On modifie fréquences, nom, puissance, largeur de bande, tonalités et, sur RT-4D et RT-6D, le balayage ; les autres éléments du canal restent tels quels. Seuls les canaux et les zones sont écrits : contacts, groupes, messages, radio FM et réglages ne sont pas touchés.

### Notes
- Testées seulement avec une radio simulée : faites d'abord une simple lecture et signalez tout problème.
- « Identifier » ne distingue pas ces radios des autres Radtel et Bajeton qui ont la même connexion : l'onglet doit être choisi à la main.
- RT-1000 Pro : l'exclusion du balayage n'est pas gérée (elle reste telle quelle sur la radio).

## IU0VYT Flash 4.19.0

### Nouveautés
- **Radtel RT-817A / RT-871A** : canaux (1024, avec des noms de 16 caractères), tirés du programme du constructeur. Même connexion et même format de canal que la RT-880G, sans les zones. Seules les zones utilisées par le programme de Radtel sont lues et écrites (réglages, canaux, radio FM) ; le logo n'est pas touché.

### Notes
- Testée seulement avec une radio simulée : faites d'abord une simple lecture et signalez tout problème.
- « Identifier » ne distingue pas la RT-817A / RT-871A de la RT-880G : l'onglet doit être choisi à la main.

## IU0VYT Flash 4.18.0

### Nouveautés
- **Kenwood TK-2140, TK-3140, TK-280, TK-380** : portatifs professionnels avec les canaux répartis en zones (groupes), canaux et réglages. La liste montre les 250 positions de la mémoire ; un nouveau canal entre dans la zone du canal qui le précède.
- **Quansheng UV-K5 / UV-K6 avec le firmware R3b0rn de fagci** : canaux (298 avec la mémoire d'origine, jusqu'à 1024 avec une mémoire plus grande). Réglages, VFO, préréglages et correction SSB ne sont jamais écrits.
- **Radtel RT-330 et RT-400** : canaux, tirés des programmes du constructeur.

### Notes
- Les pilotes de CHIRP des portatifs TK sont déclarés expérimentaux. Les zones ne se créent ni ne se suppriment depuis le programme.
- Le firmware R3b0rn, la RT-330 et la RT-400 sont testés seulement avec une radio simulée : faites d'abord une simple lecture et signalez tout problème.
- Avec le firmware R3b0rn, un canal sans nom prend la fréquence comme nom (pour ce firmware, un canal sans nom n'existe pas).
- L'onglet RT-400 refuse les radios qui s'identifient comme RT-470, RT-630, RT-495 ou HI-8811 : utilisez-le seulement avec la RT-400.

## IU0VYT Flash 4.17.0

### Nouveautés
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2** : canaux lus et écrits un par un, avec la radio allumée normalement (9600 bauds). À l'écriture, seuls les canaux modifiés sont envoyés.
- **Portatifs professionnels Kenwood TK** : canaux et réglages.
  - Série 60 : TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - Série 60G : TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860** : on utilise l'onglet « Radtel RT-880G » (même connexion et même mémoire des canaux).

### Notes
- Les Kenwood TH-F6/F7, TH-G71, TH-D7 et TH-K2 utilisent les commandes des pilotes de CHIRP et sont testées seulement avec une radio simulée : faites d'abord une simple lecture et signalez tout problème.
- Les pilotes de CHIRP des portatifs TK sont déclarés expérimentaux.
- La RT-860 vient de la comparaison avec le programme du constructeur et n'a pas été testée sur une vraie radio.

## IU0VYT Flash 4.16.0

### Nouveautés
- **OpenGD77** : canaux des radios DMR avec le firmware libre OpenGD77.
  - Radioddity GD-77 et GD-77S, Baofeng DM-1801 et RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 et MD-2017, Retevis RT3S et RT84.
  - La radio apparaît comme un port série USB ; seuls les canaux et les zones modifiés sont écrits, jamais les réglages, contacts, groupes et l'étalonnage.
- **Quansheng UV-K5, d'autres firmwares libres** :
  - **NUNU** de kamilsss655 : canaux et réglages.
  - **UV-KX** de joaquim.org : canaux et réglages ; l'étalonnage n'est jamais écrit.
  - L'onglet « OSFW » vaut aussi pour les firmwares d'origine modifiés avec les bandes élargies (uvmod-kitchen, collection de Tunas1337).
- **Radtel RT-880** sans GPS avec le firmware d'origine : on utilise l'onglet « Radtel RT-880G ».

### Notes
- OpenGD77 est écrit à partir du protocole décrit dans le programme libre qdmr et testé seulement avec une radio simulée : faites d'abord une simple lecture, conservez la copie de sauvegarde et signalez tout problème.
- NUNU et UV-KX sont comparés octet par octet avec les pilotes pour CHIRP de leurs auteurs.

## IU0VYT Flash 4.15.0

### Nouveautés
- **D-STAR et C4FM** : canaux par câble pour les portatifs radioamateurs numériques, avec des fichiers .img compatibles avec CHIRP. La colonne « Mode » affiche « D-STAR » ou « C4FM ».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) et IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (la copie part de la radio : les instructions sont dans la fenêtre du câble).
  - **Kenwood** (nouvelle marque, 47 en tout) TH-D72, TH-D74, TH-D75, avec le câble USB.
  - Réglages pour IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D et TH-D72.
- **Firmwares libres** :
  - **nicFW 2.x** de la TIDRADIO TD-H3 : canaux et réglages principaux.
  - **nicFW880 / nicFW950** des Radtel RT-880, RT-880G et RT-950 Pro : canaux.
  - **F4HWN Fusion 5.x** des Quansheng UV-K1 et UV-K5 V3 : 1024 canaux et réglages.
  - Avec nicFW, l'étalonnage de la radio n'est jamais écrit.
- **Kenwood dans la galerie** : aussi les portatifs PMR446 et DMR/NXDN, parmi les non compatibles (ils ne se programment qu'avec le programme KPG réservé aux revendeurs).

### Corrections
- Réglages : un champ de texte jamais utilisé se lit vide au lieu de « valeur non prévue » ; une valeur non modifiée n'est pas réécrite.
- Puxing PX-888K : retiré six éléments des réglages qui pouvaient écrire une donnée erronée (codes BOT/EOT et temps DTMF).
- Yaesu VX-3 : l'espace n'est plus admis dans les codes DTMF (la radio ne l'a pas).

### Notes
- Pilotes écrits de zéro à partir de ceux de CHIRP (GNU GPL), de la documentation publiée par nicsure (nicFW) et du pilote de F4HWN, et comparés octet par octet à l'aide de radios simulées. Les vraies radios de cette version n'ont pas encore été testées : conservez la copie de sauvegarde et signalez tout problème.
- Pour Icom, Yaesu et Kenwood, il n'y a pas de mise à jour du firmware (elle se fait avec les programmes des constructeurs).

## IU0VYT Flash 4.14.0

### Nouveautés
- **Yaesu et Icom** : deux nouvelles marques (46 en tout) et 13 radios de plus, avec canaux par câble et fichiers .img compatibles avec CHIRP.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (câble SCU-35) et FT-60R/E (copie lancée depuis la radio, « F8 CLONE »).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (câble OPC-478).
  - Réglages pour toutes sauf l'IC-T70. Pour ces radios il n'y a pas de mise à jour du firmware.
- **Deux langues de plus** : malais (Bahasa Melayu) et hindi (हिन्दी), 13 en tout : programme, site, guide et fichiers d'information, README et politique de sécurité.
- **Convertisseur et « Transférer vers une autre radio »** vers les Yaesu et les Icom : les tonalités que la radio n'a pas sont retirées avec un avertissement et les canaux qu'elle ne peut pas mémoriser sont sautés avec un avertissement, au lieu d'arrêter toute l'opération.

### Corrections
- Traductions : cinq mots en gras des guides restaient en italien dans toutes les langues ; deux phrases erronées en coréen et en arabe.
- Wouxun KG-UV8D : noms internes des champs alignés sur le CHIRP actuel (aucun effet sur les canaux).

### Notes
- Pilotes écrits de zéro à partir de ceux de CHIRP (GNU GPL) et comparés octet par octet avec CHIRP à l'aide de radios simulées. Les vraies radios Yaesu et Icom n'ont pas encore été testées : conservez la copie de sauvegarde et signalez tout problème.
- Les traductions en malais et en hindi sont nouvelles : signalez les textes peu clairs.

## IU0VYT Flash 4.13.1

### Nouveautés
- **Firmware IJV** pour les Quansheng UV-K5, UV-K5 V3 et UV-K1 (firmware non officiel d'IU0IJV, [ijvradio.com](https://www.ijvradio.com/)) : cinq onglets, un par version du firmware.
  - **V3** à 200 ou 999 canaux sur la UV-K5 classique (matériel V1).
  - **V4**, **V5 (« FW5 »)** et **V6** sur la UV-K5 V3 et la UV-K1 : les 500 canaux de la banque choisie sur la radio, avec les noms des groupes et des banques.
  - On lit et on écrit les canaux et les réglages principaux (menu, affichage, VOX, touches latérales). Avant d'écrire, le programme vérifie que la même banque est toujours choisie sur la radio et n'écrit que les parties modifiées ; les étalonnages ne sont jamais touchés.
  - **V6 (nouveau dans la 4.13.1)** : liaison de données de la V6, écriture par secteurs de 4 Kio, chacun relu et contrôlé. Si la radio dit que la liaison de données manque, mettez-la à jour avec FLASH V6 de RadioManager.
- **Mise à jour du firmware** : un fichier .bin avec une version IJV est signalé, avec l'onglet à utiliser pour les canaux et un avertissement pour V5 et V6.
- Traductions dans les 11 langues.

### Notes
- Protocole et mémoire tirés des modules CHIRP publiés par IJV et de l'analyse de RadioManager 4.3.2 ; code écrit de zéro, testé avec des radios simulées. Les vraies radios avec IJV n'ont pas encore été testées : signalez tout problème.
- Modules, éléments, licence et firmware IJV restent à RadioManager d'IJV.

## IU0VYT Flash 4.13.0

### Nouveautés
- **Firmware IJV** pour les Quansheng UV-K5, UV-K5 V3 et UV-K1 (firmware non officiel d'IU0IJV, [ijvradio.com](https://www.ijvradio.com/)) : quatre nouveaux onglets, un par version du firmware.
  - **V3** à 200 ou 999 canaux sur la UV-K5 classique (matériel V1).
  - **V4** et **V5 (« FW5 »)** sur la UV-K5 V3 et la UV-K1 : les 500 canaux de la banque choisie sur la radio, avec les noms des groupes et des banques.
  - On lit et on écrit les canaux et les réglages principaux (menu, affichage, VOX, touches latérales). Avant d'écrire, le programme vérifie que la même banque est toujours choisie sur la radio et n'écrit que les blocs modifiés ; les étalonnages ne sont jamais touchés.
  - La **V6** est reconnue : canaux et réglages de la V6 se gèrent avec RadioManager d'IJV.
- **Mise à jour du firmware** : un fichier .bin avec une version IJV est signalé, avec l'onglet à utiliser pour les canaux et un avertissement pour V5 et V6.
- Traductions dans les 11 langues.

### Notes
- Protocole et mémoire tirés des modules CHIRP publiés par IJV et de l'analyse de RadioManager 4.3.2 ; code écrit de zéro, testé avec des radios simulées. Les vraies radios avec IJV n'ont pas encore été testées : signalez tout problème.
