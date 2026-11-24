# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Page de la version sur GitHub (en anglais), avec le téléchargement : [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 pour Microsoft Windows 10/11

### Nouveautés
- **Retevis, 24 nouveaux modèles tirés des programmes du constructeur** : B3B, B63B, RT22P, RT622P, RB666, RB666P (même mémoire des canaux que la nouvelle version de la H777) ; RB45, RB645, RB47 ; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (famille de la T18) ; RB65, RB66, RB66P (une fréquence et une tonalité par canal). La connexion, la lecture et l'écriture sont réécrites à partir de zéro et contrôlées avec les fichiers de données d'usine des programmes. Les RB66 et RB66P, ainsi que les RB669, RT68H et RT668H, répondent avec le même code : la bonne fiche doit être choisie à la main.
- **Réglages des 24 modèles Retevis** : 241 éléments en tout (squelch, durée maximale d'émission, VOX, balayage, guide vocal, économie de batterie, bip, touches latérales…). Seuls les éléments que le programme du modèle affiche sont proposés ; les valeurs des 24 fichiers d'usine sont toutes des choix admis.
- **Mot de passe de programmation** : sur les 24 modèles Retevis et sur la nouvelle version de la Retevis H777, il n'empêche ni la lecture ni l'écriture, comme dans CHIRP (le programme du constructeur, lui, s'arrête si le mot de passe n'est pas celui saisi par l'utilisateur). C'est un élément des réglages, en clair : on peut le voir, le changer et l'enlever. La longueur et les caractères sont ceux du programme du constructeur : 6 caractères (3 pour RB45, RB645, RB47 et RB66 ; 4 pour RB66P), chiffres et lettres (chiffres uniquement pour RB65 et RB66). Pour B3H, B63H, RB45, RB645, RB66 et RB666, dont le programme n'affiche pas la case du mot de passe, un avertissement le rappelle.
- **Réglages pour des radios qui n'avaient que les canaux** : famille AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 et DMR-6X2 PRO : 27 éléments, d'après les descriptions de la mémoire de qdmr et de dmrconfig) ; TYT MD-UV380, MD-UV390, MD-2017 et MD-9600, Retevis RT3S et RT82 (les éléments de la Baofeng DM-1701) ; Radtel RT-880 et RT-950 Pro avec le firmware nicFW880 / nicFW950 (17 éléments, d'après la documentation de l'auteur) ; TID TD-UV68 (les éléments de la TD-H8) ; Radtel RT-890 avec le firmware M7OCM (53 éléments).
- **Versions d'essai** : une copie dont le numéro est plus élevé que celui de la dernière version stable publiée affiche une barre d'avertissement (« Version d'essai (bêta-test)… ») et, elle seule, propose la version d'essai suivante, toujours téléchargée depuis GitHub. Ceux qui utilisent la version stable ne voient pas les versions d'essai.
- Galerie : 697 modèles de 54 marques (647 avec canaux, 596 avec réglages).

### Notes
- Aucune des 24 radios Retevis n'a été testée sur du matériel réel, ni pour les canaux ni pour les réglages : d'abord la lecture seule, puis un seul réglage sans risque. Si la radio répond de façon inattendue, la lecture s'arrête avec un message et rien n'est écrit dans la radio.
- Avec un mot de passe de programmation déjà présent, le programme essaie quand même de lire et d'écrire, comme CHIRP : cela n'a pas été testé sur une vraie radio.
- Les RB66 et RB66P, ainsi que les RB669, RT68H et RT668H, répondent avec le même code : la bonne fiche doit être choisie à la main.
- Programme indépendant, non affilié aux constructeurs des radios.
- Suite complète de tests réussie : partie générale ; 915 comparaisons de structure et 380 lectures / écritures identiques à CHIRP.
