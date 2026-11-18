# IU0VYT Flash 4.45.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.45.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.45.0)

IU0VYT Flash V.4.45.0 per Microsoft Windows 10/11

### Novità
- **Radio DMR della famiglia AnyTone**: AnyTone AT-D868UV, AT-D878UV / AT-D878UV Plus, AT-D878UV II Plus e BTECH DMR-6X2 e DMR-6X2 PRO (queste due erano «non compatibili»). Canali e zone con il cavo USB della radio; memoria e protocollo del programma libero dmrconfig, verificati con il suo codice originale. Si scrivono solo canali, zone e liste di scansione.
- **TYT MD-UV380, MD-UV390, MD-2017, MD-9600 e Retevis RT3S e RT82** con il firmware originale: stessa memoria e stesso protocollo della Baofeng DM-1701.
- **Baofeng BF-666S** (driver della BF-888S) e **BF-T6** (sperimentale, driver della BF-T8).
- Per le schede che restano «non compatibili» UV-8D, BF-UV8, DM-1702 e DM-1703 ora è scritto il motivo.
- Galleria: 673 modelli di 54 marche (623 con canali, 556 con impostazioni).

### Note
- Nessuna delle radio nuove è stata provata su una radio vera: prima la sola lettura.
- BTECH DMR-6X2 PRO e Baofeng BF-T6 sono sperimentali.
- Programma indipendente, non affiliato ai costruttori delle radio.
- Suite completa superata: parte generale; 918 confronti di struttura e 381 letture / scritture identiche a CHIRP; famiglia AnyTone confrontata con il codice originale di dmrconfig.

---

Questa versione comprende anche le novità delle versioni dalla 4.24.0 alla 4.44.0, che non hanno avuto una release stabile propria. Eccole, dalla più recente.

## IU0VYT Flash 4.44.0

### Novità
- **Impostazioni** per altre otto radio che avevano solo i canali: Radtel RT-480, RT-68, RT-5D, RT-69 GPS, RT-4D, RT-6D, RT-1000 Pro e Bajeton DM-5000, dai programmi dei costruttori.
- **RT-480 e RT-68**: in scrittura ora arriva alla radio anche il blocco delle impostazioni generali.
- **RT-4D, RT-6D, RT-1000 Pro**: il file della radio contiene anche le impostazioni (i file salvati prima restano validi).

### Note
- Nessuna delle impostazioni nuove è stata provata su una radio vera.
- RT-69 GPS: per ora solo le voci verificate una per una sul programma del costruttore.

## IU0VYT Flash 4.43.0

### Novità
- **Errore corretto**: per 33 radio (fra cui Baofeng UV-17 Pro, UV-17 Pro GPS, UV-21 Pro, UV-25, UV-32, UV-5RH, K5 Plus, Retevis RT22 e HA1G, Kenwood TK-2140 / TK-3140) e per le 8 «gemelle» della 4.42 le impostazioni non comparivano dopo la lettura via cavo, ma solo aprendo un file di CHIRP.
- **Impostazioni** per altre cinque Radtel: RT-817A / RT-871A, RT-830, RT-850, RT-330 e RT-400, dai programmi del costruttore.

### Note
- Nessuna delle impostazioni nuove è stata provata su una radio vera.
- Radtel RT-830 e RT-850: una voce per radio è letta dal programma del costruttore in modo diverso da come è scritta; qui si segue la scrittura e la voce ha un avviso.

## IU0VYT Flash 4.42.0

### Novità
- **Scheda Home**: ora entra tutta nella finestra. Il riquadro delle marche è più largo e la pagina si adatta da sola alla dimensione della finestra.
- **Impostazioni** per 30 modelli che avevano solo i canali: Baofeng serie D (BF-V1D … BF-V12D, 888SD, 889D, 1901D, 1904D, 1909D, W31D, BF-T20D), BF-88ST, BF-C2 / BF-V2, GT-68, GT-22, T22, UV-17R, UV-21R, UV-21L, UV-25 Plus, UV-28 Pro, UV-29 Pro, UV-7B, AR-152 Pro e Radtel RT-13.

### Note
- Nessuna delle impostazioni nuove è stata provata su una radio vera.
- Baofeng serie D: nel programma del costruttore «guida vocale» e «suono dei tasti» sono scambiate fra lettura e scrittura; qui si segue la scrittura e le due voci hanno un avviso.

## IU0VYT Flash 4.41.0

### Novità
- **Radio nuove** (i palmari di CHIRP che mancavano): Yaesu VX-1R (canali e impostazioni), Icom IC-91AD e IC-92AD (due bande, anche D-STAR), BTECH UV-PRO (via Bluetooth, senza cavo) e il ricevitore AOR AR8200 (marca nuova).
- **Impostazioni** per radio che avevano solo i canali: Yaesu FT-7800 / FT-7900, Icom IC-Q7A, TIDRADIO TD-M11.

### Note
- Nessuna delle radio nuove è stata provata dal vero: prima la sola lettura.
- BTECH UV-PRO: la scrittura è ammessa solo con la versione del firmware provata dall'autore del driver di CHIRP (147).

## IU0VYT Flash 4.40.0

### Novità
- **Ricerca automatica più veloce con la scheda già aperta**: il controllo breve (prima solo per la Baofeng K67 Pro) ora vale anche per le Icom, le Kenwood TH-F6 / TH-G71 / TH-D7 / TH-K2, le Quansheng UV-K5 e derivate, le Bajeton BJ-9000, BJ-7800, BJ-7700 e BJ-8300, le Radtel RT-900 / RT-950 / RT-470 e le radio con la stessa memoria. Per le altre resta la lettura intera.
- **Marca nuova: Alinco.** Palmari DJ-596, DJ-175, DJ-G7EG e DJ-G7T; veicolari DR-135T, DR-235T, DR-435T, DR-03T e DR-06T. Solo canali. Funziona anche con i cavi che non rimandano indietro i dati.

### Note
- Nessuna Alinco è stata provata dal vero: prima la sola lettura.
- Il controllo breve è provato con le radio simulate; se dopo il controllo una radio non torna al funzionamento normale, segnalalo.

## IU0VYT Flash 4.39.0

### Novità
- **Quattro radio Icom nuove**: ID-80H (palmare D-STAR), ID-880H (veicolare D-STAR), IC-W32A e IC-W32E (palmari bibanda con due metà separate: da 0 a 99 i canali VHF, da 100 a 199 i canali UHF). Solo canali.

### Note
- Nessuna delle radio nuove è stata provata dal vero: prima la sola lettura.

## IU0VYT Flash 4.38.0

### Correzioni
- **Identificazione**: la Bajeton BJ-9000 ogni tanto si bloccava, soprattutto alla prima identificazione. Il programma è tornato ad ascoltare la porta per un secondo prima del primo saluto, come nella 4.35, anche per la radio già riconosciuta su quel cavo. Le altre migliorie della 4.36 restano.

### Note
- Nessuna radio nuova in questa versione.

## IU0VYT Flash 4.37.0

### Novità
- **Otto palmari Yaesu nuovi**: VX-5R e VX-7R (canali); VX-2R, FT-50R, VX-150, VX-8R, VX-8DR e VX-8GE (canali e impostazioni; per le VX-8 anche APRS).
- La copia parte dalla radio (modalità «CLONE»): la scheda si sceglie a mano e le istruzioni sono nella finestra del cavo.

### Note
- Nessuna delle radio nuove è stata provata dal vero: prima la sola lettura.
- Non fatto: Yaesu VX-1.

## IU0VYT Flash 4.36.0

### Correzioni
- **«Identifica» e identificazione automatica**: il programma ricorda su quale cavo ha riconosciuto ogni radio e comincia da quelle; la prima ricerca non perde più secondi in ascolto e in preparazione; la radio accesa dopo il cavo viene riprovata più spesso.
- Il riquadro «Identificazione Radio In Corso» compare solo mentre la radio si fa leggere, con il modello e l'avanzamento. Se la radio è collegata ma spenta il programma avvisa di accenderla e la riconosce appena è accesa.
- Baofeng K67 Pro con la scheda già aperta: riconosciuta in un paio di secondi, senza rileggere tutta la memoria.

### Novità
- Scheda «Identifica»: pulsante **«Salva la diagnosi dell'identificazione»** (tempi e primi byte delle ultime ricerche in un file di testo, da inviare all'autore se l'identificazione è lenta o sbaglia; non contiene canali né dati personali).

### Note
- Nessuna radio nuova in questa versione.

## IU0VYT Flash 4.35.0

### Correzioni
- **«Identifica» e identificazione automatica**: tolto il «giro veloce» della 4.31, che con le radio vere poteva lasciare la radio bloccata in programmazione e rendere la ricerca lenta. Le attese sono di nuovo quelle di prima.
- La ricerca prova gli stessi gruppi di radio su ogni PC (prima cambiavano da un avvio all'altro).
- L'eco di un cavo senza radio non viene più scambiata per una radio.

### Note
- Nessuna radio nuova in questa versione.

## IU0VYT Flash 4.34.0

### Novità
- **Radio nuove**: Yaesu VX-170 e VX-177 (palmari), FT-7800 / FT-7900 e FT-8900 (veicolari), solo canali, con i driver di CHIRP: lettura, scrittura e byte scambiati con la radio identici a CHIRP.

### Note
- La copia parte dalla radio (modalità «CLONE»): la scheda va scelta a mano, «Identifica» non prova queste radio.

## IU0VYT Flash 4.33.0

### Novità
- **Radio nuove**: Icom IC-T7H, IC-T8A, IC-Q7A e IC-P7 (canali), con i driver di CHIRP: lettura, scrittura e byte scambiati con la radio identici a CHIRP.

### Note
- IC-Q7A e IC-P7: il modo (FM, FM larga, AM) segue la frequenza; salto «P» e passo di sintonia del canale si conservano.

## IU0VYT Flash 4.32.0

### Novità
- **Radio nuove**: Baofeng T22 in versione FRS, TIDRADIO TD-M8S, TD-V2, TD-H4 e TD-H11 (canali e impostazioni), dai programmi dei costruttori.
- **«Identifica»** propone anche le radio nuove che si presentano come la BF-888S, senza allungare la ricerca: la radio si legge con il loro driver solo quando si apre quella scheda.

### Correzioni
- TIDRADIO del 2026 (TD-H3 Plus / SE, TD-H7 Plus, TD-H8 4ª generazione, TD-H9, TD-638UV): la potenza media è offerta solo dai modelli che l'hanno davvero, e mai nelle versioni GMRS.
- Baofeng T22 PMR446: collegamento come il programma del costruttore.
- Note di Radtel RT-13, RT-480 / RT-68, RT-830 / RT-850, RT-69 GPS e TIDRADIO TD-M2 Pro: «Identifica» le prova, la frase che diceva il contrario è stata tolta.

## IU0VYT Flash 4.31.0

### Novità
- **«Identifica» è molto più veloce**: prima fa un giro rapido di tutte le radio e solo dopo, se non ha trovato nulla, quello completo di prima. Una radio in fondo all'elenco si trova in meno di mezzo minuto invece di quasi due.
- **«Identifica» ora prova anche le radio aggiunte nelle ultime versioni**: Retevis RB39 / RB639, C6, C66, Radioddity Echo, UB15-G, DB35-G, DB35-UV, DB40-G e i TIDRADIO senza display.

### Correzioni
- Retevis RB39 / RB639: con la scheda del paese sbagliato la radio restava in programmazione fino allo spegnimento.

## IU0VYT Flash 4.30.0

### Novità
- **Impostazioni generali** per TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 di quarta generazione**, **TD-H9** e **TD-638UV**: tasti programmabili, squelch, VOX, tempo massimo di trasmissione, display, lingua, messaggi di accensione e altro.
- **Radio nuove (canali e impostazioni)**: TIDRADIO **TID-712Plus**, **TD-M7**, **TD-M11 Plus**, **TD-BR09**, **TD-777S**, **TD-V900**, **TD-FRS**, **TD-PMR** e **TD-M2 Pro**.

### Note
- Le radio nuove vengono dai programmi del costruttore e non sono ancora provate con una radio vera: fai prima la sola lettura. «Identifica» non le distingue fra loro: scegli tu la scheda.

## IU0VYT Flash 4.29.0

### Novità
- **Radio nuove (solo canali)**: TIDRADIO **TD-H3 Plus**, **TD-H3 SE**, **TD-H7 Plus**, **TD-H8 di quarta generazione**, **TD-H9** e **TD-638UV**, ognuna nelle versioni base, HAM e GMRS.
- La TD-H3 Plus si presenta come la TD-H3: usa le schede «TD-H3 Plus», perché ha tre potenze.

### Note
- Queste radio vengono dal programma del costruttore del 2026 («TIDRadioCPS») e non sono ancora provate con una radio vera: fai prima la sola lettura. Le impostazioni generali non sono ancora gestite.

## IU0VYT Flash 4.28.0

### Novità
- **Radio nuove** (veicolari Radioddity): **UB15-G** (GMRS 15 W), **DB35-G** (GMRS 35 W), **DB35-UV** (VHF / UHF 35 W) e **DB40-G** (GMRS 40 W, 199 canali), con canali e impostazioni generali. La UB15-G ha due schede, secondo il firmware: prima della V1.10 (senza nomi) e V1.10 (con i nomi).
- La **DB40-G** è anche nel CHIRP: qui legge e scrive gli stessi byte e gli stessi file `.img`.

### Correzioni
- Le prove non si compilavano su Windows dalla 4.26.0 (tre file di prova usavano le radio simulate, che su Windows non vengono compilate). Il programma non ne era toccato.

### Note
- UB15-G, DB35-G e DB35-UV vengono dal programma del costruttore e dal suo file di configurazione e non sono ancora provate con una radio vera: fai prima la sola lettura. «Identifica» non riconosce queste radio: la scheda va scelta a mano.

## IU0VYT Flash 4.27.0

### Novità
- **Radio nuove**: **Retevis C6** (FRS / GMRS, 30 canali), **Retevis C66** (UHF, 16 canali) e **Radioddity Echo-DE / Echo-IT** (PMR446, 16 canali), con canali e impostazioni generali.
- Prima di scrivere il programma controlla che gamma, modello e area della radio siano quelli della memoria da scrivere; password e tarature della radio non vengono mai scritte.

### Correzioni
- **Retevis RB39 / RB39P**: ora è fra le radio «GMRS / FRS (USA)» nella scheda Modelli.

### Note
- Le radio nuove vengono dal codice dei programmi dei costruttori e non sono ancora provate con una radio vera: fai prima la sola lettura. «Identifica» non le riconosce: la scheda va scelta a mano.
- Se nella radio è impostata una password di lettura, va tolta prima con il programma del costruttore.

## IU0VYT Flash 4.26.0

### Novità
- **Radio nuove**: **Retevis RB39 / RB39P** (FRS, 22 canali) e **RB639 / RB639P** (PMR446, 16 canali), con canali e impostazioni generali. La radio dichiara per quale paese è fatta: con la scheda sbagliata la lettura si ferma e indica quella giusta.
- **Impostazioni nuove** per **Baofeng BF-T17 / BF-T18** (PMR446 e FRS), **GT-18 PMR446**, **Radioddity PR-T5**, **FS-T3** e **PR-T3**: squelch, tempo massimo di trasmissione, VOX, guida vocale, suono dei tasti, risparmio della batteria e le altre del programma del costruttore.

### Correzioni
- Registri delle modifiche nelle 12 lingue: data della 4.25.0 corretta (12/11/2026).

### Note
- Le radio nuove vengono dal codice dei programmi dei costruttori e non sono ancora provate con una radio vera: fai prima la sola lettura. «Identifica» non le riconosce: la scheda va scelta a mano.

## IU0VYT Flash 4.25.0

### Novità
- **Bajeton BJ-8300 con il firmware nuovo**: scheda nuova per le radio con 990 canali in 10 zone da 99 (programma del costruttore BT-11_CPS v1.2.6): canali, nomi delle zone e impostazioni. La scheda «Bajeton 8300» di sempre (960 canali, 15 zone da 64) resta com'è e ora **avvisa** quando trova il firmware nuovo.
- **Radio nuove**: **Radioddity GS-10B Pro** (999 canali e impostazioni), **Radioddity FS-T3** e **PR-T3** (canali).
- **Bajeton BJ-9000** con 990 canali: cinque impostazioni in più (quelle del programma v1.3.3).

### Correzioni
- **Baofeng BF-T17 / BF-T18, GT-18 e Radioddity PR-T5**: potenza alta e bassa erano scambiate. Chi ha scritto queste radio con una versione precedente le rilegga e scelga di nuovo la potenza.
- «Uscita automatica dal menu» (BJ-8300, BJ-9000, Radtel RT-950): l'ultima scelta è 60 secondi, non 55.

### Note
- Le radio e la scheda nuove vengono dal codice dei programmi dei costruttori e non sono ancora provate con una radio vera: fai prima la sola lettura.
- BJ-8300: per sapere quale scheda usare guarda sulla radio quante zone ci sono (15 da 64 canali, oppure 10 da 99).

## IU0VYT Flash 4.24.0

### Novità
- **Scheda Modelli più leggibile**: le etichette delle tipologie di radio sono al centro della pagina e la riga fra radio compatibili e non compatibili è rossa e più spessa.
- **26 radio in più**, gestite con i driver di radio uguali già presenti:
  - riconosciute confrontando i programmi dei costruttori: **Bajeton BT-8000**, **Radioddity GC-5**, **Trivox**, **PR-T5**, **GT-82**, **Retevis RT28**;
  - dalla tabella dei modelli equivalenti di CHIRP: Radioddity GA-5S, Retevis RT5 (tre potenze), RT5R, RT5RV, RT5, TIDRADIO TD-UV5R TriPower, TechSide TI-F8+ e TS-T9+, Rugged RH5R e RH5X, Ansoko A-5R e A-8S, Arcshell AR-5, AR-6 e AR-8, Greaval GV-8S e GV-9S, Tenway TW-325, UV-82 Pro e UV-5R Pro.

### Note
- Le sei radio riconosciute dal confronto dei programmi dei costruttori non sono ancora provate con una radio vera: fai prima la sola lettura. Se il modello non è quello giusto la radio non risponde e non viene modificata.
