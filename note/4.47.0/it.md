# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub><b>Italiano</b></sub></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 per Microsoft Windows 10/11

### Novità
- **Retevis, 24 modelli nuovi ricavati dai programmi del costruttore**: B3B, B63B, RT22P, RT622P, RB666, RB666P (stessa memoria dei canali della H777 nuova versione); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (famiglia della T18); RB65, RB66, RB66P (una frequenza e un tono per canale). Collegamento, lettura e scrittura sono riscritti da zero e controllati con i file dati di fabbrica dei programmi. RB66 e RB66P, e RB669, RT68H e RT668H, rispondono con la stessa sigla: la scheda giusta va scelta a mano.
- **Impostazioni dei 24 modelli Retevis**: 241 voci in tutto (squelch, tempo massimo di trasmissione, VOX, scansione, guida vocale, risparmio della batteria, bip, tasti laterali…). Si offrono solo le voci che il programma del modello mostra; i valori dei 24 file di fabbrica sono tutti scelte ammesse.
- **Password di programmazione**: nei 24 modelli Retevis e nella Retevis H777 nuova versione non ferma la lettura né la scrittura, come in CHIRP (il programma del costruttore invece si ferma, se la password non è quella scritta dall'utente). È una voce delle impostazioni, in chiaro: si vede, si cambia e si toglie. Lunghezza e caratteri sono quelli del programma del costruttore: 6 caratteri (3 per RB45, RB645, RB47 e RB66; 4 per RB66P), cifre e lettere (solo cifre per RB65 e RB66). Per B3H, B63H, RB45, RB645, RB66 e RB666, il cui programma non mostra la casella della password, un avviso lo ricorda.
- **Impostazioni per radio che avevano solo i canali**: famiglia AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 e DMR-6X2 PRO: 27 voci, dalle descrizioni della memoria di qdmr e dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 e MD-9600, Retevis RT3S e RT82 (le voci della Baofeng DM-1701); Radtel RT-880 e RT-950 Pro con il firmware nicFW880 / nicFW950 (17 voci, dalla documentazione dell'autore); TID TD-UV68 (le voci della TD-H8); Radtel RT-890 con il firmware M7OCM (53 voci).
- **Versioni di prova**: una copia con un numero più alto dell'ultima versione stabile pubblicata mostra una barra di avviso («Versione di prova (beta testing)…») e, solo lei, propone la versione di prova successiva, scaricata sempre da GitHub. Chi usa la versione stabile non vede le versioni di prova.
- Galleria: 697 modelli di 54 marche (647 con canali, 596 con impostazioni).

### Note
- Nessuna delle 24 radio Retevis è stata provata su una radio vera, né per i canali né per le impostazioni: prima la sola lettura, poi una sola impostazione innocua. Se la radio risponde in un modo inatteso la lettura si ferma con un messaggio e nella radio non viene scritto niente.
- Con una password di programmazione già presente il programma prova comunque a leggere e a scrivere, come CHIRP: su una radio vera non è stato provato.
- RB66 e RB66P, e RB669, RT68H e RT668H, rispondono con la stessa sigla: la scheda giusta va scelta a mano.
- Programma indipendente, non affiliato ai costruttori delle radio.
- Suite completa superata: parte generale; 915 confronti di struttura e 380 letture / scritture identiche a CHIRP.
