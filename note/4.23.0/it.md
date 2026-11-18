# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 per Microsoft Windows 10/11

### Novità
- **Kenwood TK-481 e TK-981** (900 MHz, canali divisi in sistemi) e le veicolari **TK-780 e TK-880**: canali e impostazioni, con i driver di CHIRP.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: ora anche le impostazioni della radio.
- **Radtel RT-13**: canali (22, senza nome), ricavati dal programma del costruttore.
- **Radtel RT-69 GPS**: canali (999, con nomi di 12 caratteri), ricavati dal programma del costruttore; prima di scrivere si controlla che ogni canale stia nelle bande della radio.
- Più robusto con i file rovinati: corretti due arresti nella lettura dei toni di Baofeng UV-B5 / UV-B6 e UV-3R.
- Conversioni: un canale che la radio di arrivo non può memorizzare viene saltato con un avviso invece di fermare la conversione (firmware IJV, nicFW, OpenGD77, UV-KX, R3b0rn, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Galleria: 569 modelli di 47 marche (515 con canali, 433 con impostazioni).

### Note
- TK-481 / TK-981: si usano solo i sistemi che la radio ha già; quelli a risorse condivise (trunking) non si modificano.
- RT-13 e RT-69 GPS: se sulla radio è impostata una password, toglila prima con il programma di Radtel. Ricavate dai programmi del costruttore e provate solo con una radio simulata: fai prima la sola lettura.
- Programma indipendente, non affiliato ai costruttori delle radio.

---

Questa versione comprende anche le novità delle versioni dalla 4.13.0 alla 4.22.0, pubblicate solo come versioni di prova. Eccole, dalla più recente.

## IU0VYT Flash 4.22.0

### Novità
- **Radtel RT-480 e RT-68**: canali e nomi (256 canali, numerati da 0 a 255, con nomi di 12 caratteri), ricavati dal programma del costruttore. Si modificano frequenze, nome, potenza, banda, scansione e toni; le altre voci del canale restano come sono.
- **Radtel RT-830 e RT-850**: canali (128, senza nome), ricavati dai programmi del costruttore. Si modificano frequenze, potenza, banda, scansione e toni.
- **Radtel RT-20M**: si apre con la scheda Ruyage UV58Plus (stessa memoria della iRadio UV-5118plus).

### Note
- RT-480 / RT-68: si scrivono solo canali e nomi; impostazioni, VFO e DTMF della radio non si toccano. Frequenze a passi di 1,25 kHz; ogni canale deve avere la frequenza di trasmissione.
- RT-830 / RT-850: impostazioni e VFO tornano come sono; la password della radio non si scrive mai. Se sulla radio è impostata una password, toglila prima con il programma di Radtel.
- Tutte ricavate dai programmi del costruttore e provate solo con una radio simulata: fai prima la sola lettura.

## IU0VYT Flash 4.21.0

### Novità
- **Radtel RT-5D (DMR)**: canali (16 zone da 64, con nomi di 12 caratteri), ricavati dal programma del costruttore. Si modificano frequenze, nome, potenza, banda, scansione e toni; le altre voci del canale restano come sono.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** e i veicolari **TK-7180, TK-8180** (anche E): canali, con i driver di CHIRP. La lista ha 250 posizioni per ognuna delle prime 32 zone; si usano solo le zone che la radio ha già, 512 canali in tutto.

### Note
- RT-5D: la lettura prende tutta la memoria e la scrittura la rimanda tutta com'è nel file, come fa il programma di Radtel. Prima di scrivere fai una lettura fresca: un file vecchio riporterebbe contatti e impostazioni di allora. Provata solo con una radio simulata.
- TK-x180: solo i canali, niente impostazioni; le zone non si creano e non si cancellano. Se la radio ha zone non convenzionali (trunking) la lettura viene rifiutata.

## IU0VYT Flash 4.20.0

### Novità
- **Radtel RT-4D e RT-6D (DMR) e RT-1000 Pro**: canali (1024, con nomi di 16 caratteri) e zone (250), ricavati dai programmi del costruttore. Si modificano frequenze, nome, potenza, banda, toni e, su RT-4D e RT-6D, la scansione; le altre voci del canale restano come sono. Si scrivono solo canali e zone: contatti, gruppi, messaggi, radio FM e impostazioni non si toccano.

### Note
- Provate solo con una radio simulata: prima fai la sola lettura e segnala ogni problema.
- «Identifica» non distingue queste radio dalle altre Radtel e Bajeton con lo stesso collegamento: la scheda va scelta a mano.
- RT-1000 Pro: l'esclusione dalla scansione non si gestisce (resta com'è sulla radio).

## IU0VYT Flash 4.19.0

### Novità
- **Radtel RT-817A / RT-871A**: canali (1024, con nomi di 16 caratteri), ricavati dal programma del costruttore. Stesso collegamento e stesso canale della RT-880G, senza le zone. Si leggono e si scrivono solo le aree che usa il programma di Radtel (impostazioni, canali, radio FM); il logo non viene toccato.

### Note
- Provata solo con una radio simulata: prima fai la sola lettura e segnala ogni problema.
- «Identifica» non distingue la RT-817A / RT-871A dalla RT-880G: la scheda va scelta a mano.

## IU0VYT Flash 4.18.0

### Novità
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: palmari professionali con i canali divisi in zone (gruppi), canali e impostazioni. La lista mostra le 250 posizioni della memoria; un canale nuovo entra nella zona del canale che lo precede.
- **Quansheng UV-K5 / UV-K6 con il firmware R3b0rn di fagci**: canali (298 con la memoria originale, fino a 1024 con una memoria più grande). Impostazioni, VFO, preset e correzione SSB non vengono mai scritti.
- **Radtel RT-330 e RT-400**: canali, ricavati dai programmi del costruttore.

### Note
- I driver di CHIRP dei palmari TK sono dichiarati sperimentali. Le zone non si creano e non si cancellano dal programma.
- Il firmware R3b0rn, la RT-330 e la RT-400 sono provati solo con una radio simulata: prima fai la sola lettura e segnala ogni problema.
- Con il firmware R3b0rn un canale senza nome prende come nome la frequenza (per quel firmware un canale senza nome non esiste).
- La scheda RT-400 rifiuta le radio che si identificano come RT-470, RT-630, RT-495 o HI-8811: usala solo con la RT-400.

## IU0VYT Flash 4.17.0

### Novità
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: canali letti e scritti un canale per volta, con la radio accesa normalmente (9600 baud). In scrittura si inviano solo i canali cambiati.
- **Palmari professionali Kenwood TK**: canali e impostazioni.
  - Serie 60: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - Serie 60G: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: si usa la scheda «Radtel RT-880G» (stesso collegamento e stessa memoria dei canali).

### Note
- Le Kenwood TH-F6/F7, TH-G71, TH-D7 e TH-K2 usano i comandi dei driver di CHIRP e sono provate solo con una radio simulata: prima fai la sola lettura e segnala ogni problema.
- I driver di CHIRP dei palmari TK sono dichiarati sperimentali.
- La RT-860 nasce dal confronto con il programma del costruttore e non è stata provata su una radio vera.

## IU0VYT Flash 4.16.0

### Novità
- **OpenGD77**: canali delle radio DMR con il firmware libero OpenGD77.
  - Radioddity GD-77 e GD-77S, Baofeng DM-1801 e RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 e MD-2017, Retevis RT3S e RT84.
  - La radio compare come porta seriale USB; si scrivono solo i canali e le zone cambiati, mai impostazioni, contatti, gruppi e taratura.
- **Quansheng UV-K5, altri firmware liberi**:
  - **NUNU** di kamilsss655: canali e impostazioni.
  - **UV-KX** di joaquim.org: canali e impostazioni; la taratura non viene mai scritta.
  - La scheda «OSFW» vale anche per i firmware originali modificati con le bande ampliate (uvmod-kitchen, raccolta di Tunas1337).
- **Radtel RT-880** senza GPS con il firmware originale: si usa la scheda «Radtel RT-880G».

### Note
- OpenGD77 è scritto dal protocollo descritto nel programma libero qdmr ed è provato solo con una radio simulata: prima fai la sola lettura, conserva la copia di sicurezza e segnala ogni problema.
- NUNU e UV-KX sono confrontati byte per byte con i driver per CHIRP dei loro autori.

## IU0VYT Flash 4.15.0

### Novità
- **D-STAR e C4FM**: canali via cavo per i palmari radioamatoriali digitali, con file .img compatibili con CHIRP. La colonna «Modo» mostra «D-STAR» o «C4FM».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) e IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (la copia parte dalla radio: le istruzioni sono nella finestra del cavo).
  - **Kenwood** (marca nuova, 47 in tutto) TH-D72, TH-D74, TH-D75, con il cavo USB.
  - Impostazioni per IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D e TH-D72.
- **Firmware liberi**:
  - **nicFW 2.x** della TIDRADIO TD-H3: canali e impostazioni principali.
  - **nicFW880 / nicFW950** di Radtel RT-880, RT-880G e RT-950 Pro: canali.
  - **F4HWN Fusion 5.x** di Quansheng UV-K1 e UV-K5 V3: 1024 canali e impostazioni.
  - Con nicFW la taratura della radio non viene mai scritta.
- **Kenwood in galleria**: anche i palmari PMR446 e DMR/NXDN, tra i non compatibili (si programmano solo con il programma KPG riservato ai rivenditori).

### Correzioni
- Impostazioni: un campo di testo mai usato si legge vuoto invece di «valore non previsto»; un valore non cambiato non viene riscritto.
- Puxing PX-888K: tolte sei voci delle impostazioni che potevano scrivere un dato sbagliato (codici BOT/EOT e tempi DTMF).
- Yaesu VX-3: nei codici DTMF lo spazio non è più ammesso (la radio non lo ha).

### Note
- Driver scritti da zero a partire da quelli di CHIRP (GNU GPL), dalla documentazione pubblicata da nicsure (nicFW) e dal driver di F4HWN, e confrontati byte per byte usando radio simulate. Le radio vere di questa versione non sono ancora state provate: conserva la copia di sicurezza e segnala ogni problema.
- Per Icom, Yaesu e Kenwood non c'è l'aggiornamento del firmware (si fa con i programmi dei costruttori).

## IU0VYT Flash 4.14.0

### Novità
- **Yaesu e Icom**: due marche nuove (46 in tutto) e 13 radio in più, con canali via cavo e file .img compatibili con CHIRP.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (cavo SCU-35) e FT-60R/E (copia avviata dalla radio, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (cavo OPC-478).
  - Impostazioni per tutte tranne la IC-T70. Per queste radio non c'è l'aggiornamento del firmware.
- **Due lingue in più**: malese (Bahasa Melayu) e hindi (हिन्दी), 13 in tutto: programma, sito, guida e file informativi, README e politica di sicurezza.
- **Convertitore e «Inoltra ad altra radio»** verso le Yaesu e le Icom: i toni che la radio non ha vengono tolti con un avviso e i canali che non può memorizzare vengono saltati con un avviso, invece di fermare tutta l'operazione.

### Correzioni
- Traduzioni: cinque parole in grassetto delle guide restavano in italiano in tutte le lingue; due frasi sbagliate in coreano e in arabo.
- Wouxun KG-UV8D: nomi interni dei campi allineati a CHIRP di oggi (nessun effetto sui canali).

### Note
- Driver scritti da zero a partire da quelli di CHIRP (GNU GPL) e confrontati byte per byte con CHIRP usando radio simulate. Le radio Yaesu e Icom vere non sono ancora state provate: conserva la copia di sicurezza e segnala ogni problema.
- Le traduzioni in malese e in hindi sono nuove: segnala i testi poco chiari.

## IU0VYT Flash 4.13.1

### Novità
- **Firmware IJV** per le Quansheng UV-K5, UV-K5 V3 e UV-K1 (firmware non originale di IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): cinque schede, una per versione del firmware.
  - **V3** da 200 o 999 canali sulla UV-K5 classica (hardware V1).
  - **V4**, **V5 («FW5»)** e **V6** sulla UV-K5 V3 e sulla UV-K1: i 500 canali del banco scelto sulla radio, con i nomi dei gruppi e dei banchi.
  - Si leggono e si scrivono canali e impostazioni principali (menu, display, VOX, tasti laterali). Prima di scrivere il programma controlla che sulla radio sia ancora scelto lo stesso banco e scrive solo le parti cambiate; le tarature non si toccano mai.
  - **V6 (nuovo nella 4.13.1)**: collegamento dati della V6, scrittura per settori di 4 KiB, ognuno riletto e controllato. Se la radio dice che manca il collegamento dati, aggiornala con FLASH V6 di RadioManager.
- **Aggiornamento del firmware**: un file .bin con una versione IJV viene segnalato, con la scheda da usare per i canali e un avviso per V5 e V6.
- Traduzioni in tutte le 11 lingue.

### Note
- Protocollo e memoria ricavati dai moduli CHIRP pubblicati da IJV e dall'analisi di RadioManager 4.3.2; codice scritto da zero, provato con radio simulate. Le radio vere con IJV non sono ancora state provate: segnala ogni problema.
- Moduli, voci, licenza e firmware IJV restano a RadioManager di IJV.

## IU0VYT Flash 4.13.0

### Novità
- **Firmware IJV** per le Quansheng UV-K5, UV-K5 V3 e UV-K1 (firmware non originale di IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): quattro schede nuove, una per versione del firmware.
  - **V3** da 200 o 999 canali sulla UV-K5 classica (hardware V1).
  - **V4** e **V5 («FW5»)** sulla UV-K5 V3 e sulla UV-K1: i 500 canali del banco scelto sulla radio, con i nomi dei gruppi e dei banchi.
  - Si leggono e si scrivono canali e impostazioni principali (menu, display, VOX, tasti laterali). Prima di scrivere il programma controlla che sulla radio sia ancora scelto lo stesso banco e scrive solo i blocchi cambiati; le tarature non si toccano mai.
  - La **V6** viene riconosciuta: canali e impostazioni della V6 si gestiscono con RadioManager di IJV.
- **Aggiornamento del firmware**: un file .bin con una versione IJV viene segnalato, con la scheda da usare per i canali e un avviso per V5 e V6.
- Traduzioni in tutte le 11 lingue.

### Note
- Protocollo e memoria ricavati dai moduli CHIRP pubblicati da IJV e dall'analisi di RadioManager 4.3.2; codice scritto da zero, provato con radio simulate. Le radio vere con IJV non sono ancora state provate: segnala ogni problema.
