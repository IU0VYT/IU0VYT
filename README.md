IU0VYT FLASH - CANALI, PONTI E FIRMWARE DELLE RADIO BAOFENG E BAJETON
=====================================================================
Versione 3.1.2 - un solo file eseguibile, nessuna installazione, interfaccia in italiano.
Autore: Daniele "IU0VYT" Bonifazi - cbmanz@gmail.com - https://www.qrz.com/db/IU0VYT
Licenza: GNU General Public License versione 3 o successiva (software libero, SENZA
ALCUNA GARANZIA): vedi LICENSE.txt e AUTORI_E_LICENZE.txt.
Novità della 3.1.2:
- corretta la verifica degli aggiornamenti: il server dei file di GitHub rifiutava la
  richiesta con «501 Not Implemented» (non accetta le richieste "ultimi N byte"). Ora
  il programma chiede intervalli espliciti e, se un server non accetta richieste
  parziali, legge il pacchetto intero: in ogni caso dice se il software è aggiornato o
  propone di scaricare la versione più recente.
Novità della 3.1.1:
- aggiornamenti: controllo automatico a ogni avvio e pulsante "🔄 Aggiornamenti" (in
  alto e nella scheda Info); se è disponibile una versione più recente compare un
  avviso con "Scarica l'aggiornamento", altrimenti non viene scaricato nulla. Il
  controllo legge solo il numero di versione pubblicato; l'indirizzo è fisso e non si
  modifica più dal programma (vedi "SCHEDA INFO E AGGIORNAMENTI");
- archivi .rar aperti dal programma stesso con un decodificatore libero (rardecode,
  licenza BSD) al posto di UnRAR, la cui licenza non è compatibile con la GPL;
- firmware .kdhx, .kdh e .Firmware riconosciuti anche dentro archivi .zip e .rar;
- licenza GNU GPL v3 resa esplicita: testo ufficiale (LICENSE.txt), autori e licenze
  dei componenti (AUTORI_E_LICENZE.txt), note di copyright in ogni file sorgente;
- nota sui marchi: programma indipendente, non affiliato a Baofeng o Bajeton.
Novità della 3.1 (dai programmi ufficiali scaricati dai siti Baofeng e Bajeton):
- canali di 15 radio in più (75 modelli della galleria in tutto): UV-25 Plus, AR-152 Pro,
  UV-29 Pro / SW-9, UV-7B, UV-17R, UV-21R, UV-21L, T22, BF-C2, BF-88ST, BF-88A/E,
  serie H, GM-15 Pro, Bajeton BJ-7700 (con e senza Bluetooth) e BJ-9000 (vedi "RADIO
  RICAVATE DAI PROGRAMMI DEI COSTRUTTORI" e "RADIO GESTITE CON I DRIVER DI CHIRP");
- aggiornamento del firmware anche per UV-25 Plus (.kdhx), UV-9G (.kdh), Bajeton
  BJ-9000 (.Firmware) e BJ-7700 con Bluetooth (.anc) (vedi "AGGIORNAMENTO DEL
  FIRMWARE BAOFENG");
- nuovo indirizzo per la verifica degli aggiornamenti (rilascio GitHub, vedi "SCHEDA
  INFO E AGGIORNAMENTI").
Novità della 3.0:
- nuova scheda "📻 Modelli": galleria di 105 radio Baofeng e Bajeton (97 Baofeng, 8
  Bajeton) con ricerca, filtri e, per ogni radio, le funzioni disponibili e i link ai
  programmi e ai manuali ufficiali (vedi "GALLERIA DEI MODELLI");
- canali e ponti di 59 modelli: oltre a K67 Pro e Bajeton 8300, 57 radio gestite con
  driver ricavati da CHIRP (lettura e scrittura via cavo, file .img compatibili con
  CHIRP), tra cui UV-5R e derivate, UV-82, UV-5RM, UV-5RH, UV-17 Pro, UV-21 Pro, UV-25,
  K5 Plus, K6, UV-5R Mini, UV-9R, UV-9G, BF-888S, BJ-7800 (vedi "RADIO GESTITE CON I
  DRIVER DI CHIRP");
- aggiornamento del firmware anche per Bajeton BJ-7800 (.anc, stesso programma della
  BJ-8300), Baofeng UV-5RM/UV-5RH (.BF) e Baofeng 5RH Pro / K67 Pro (.dat) (vedi
  "AGGIORNAMENTO DEL FIRMWARE BAOFENG");
- canali copiabili tra tutte le radio aperte ("Copia in…") e dall'archivio ponti
  ("→ Altra radio…"); canali di sola ricezione (TX = OFF) dove la radio lo permette.
Novità della 2.2: il programma si chiama "IU0VYT Flash" (prima "Gestione canali");
scheda "ℹ Info" con versione, data di compilazione, autore e verifica degli
aggiornamenti (vedi "SCHEDA INFO E AGGIORNAMENTI").
Novità della 2.1: aggiornamento del firmware della Bajeton 8300 (BJ-8300).
Novità della 2.0: lettura e scrittura delle radio con il cavo di programmazione.

AVVIO
- Fai doppio clic su "IU0VYT Flash.exe". Si apre una finestra con le schede
  "📻 Modelli", "Baofeng K67 Pro", "Bajeton 8300", "Archivio ponti", "ℹ Info"; le
  radio scelte dalla galleria si aprono in schede in più (chiudibili con ×).
- Aggiornando dalla versione 2.x basta sostituire l'eseguibile: archivio, impostazioni
  e copie restano dove sono.
- Aggiornando dalla versione 2.1 o precedenti (GestioneCanali.exe): archivio ponti,
  impostazioni e copie restano gli stessi (cartella %APPDATA%\GestioneCanaliRadio);
  il vecchio GestioneCanali.exe si può cancellare.
- Il programma non ha firma digitale: al primo avvio Windows può mostrare
  "Windows ha protetto il PC". Clicca "Ulteriori informazioni" e poi "Esegui comunque".
- La finestra usa il componente WebView2, già presente in Windows 11. Se mancasse,
  il programma si apre nel browser predefinito e una finestrella permette di chiuderlo.
- Si può aprire una sola copia del programma alla volta.

COME SI LAVORA
Con il cavo di programmazione:
1. "📻 Leggi dalla radio…": i canali della radio compaiono nella tabella.
2. Modifica i canali; salvali anche in un file (Salva con nome…), per averne copia.
3. "📻 Scrivi nella radio…".
Con i file (come nella versione 1.0):
- Apri un file .xlc (K67 Pro) o .dat (Bajeton 8300) salvato con il programma
  originale o con questo, modifica, salva (Ctrl+S). I file salvati qui si possono
  aprire e scrivere anche con il programma originale.
Quando salvi sopra un file esistente, la versione precedente viene conservata
accanto con estensione .bak (es. k67.xlc.bak).

GALLERIA DEI MODELLI
- La scheda "📻 Modelli" mostra tutte le radio Baofeng (catalogo e pagina dei download
  di baofengradio.com) e Bajeton (bajeton.com). Cerca per nome, anche alternativo
  (es. UV-5RA, MINI-5, GT-3TP), e filtra per marca, tipo (radioamatoriali, GMRS,
  PMR/FRS, DMR, per bambini), radio con canali via cavo o con firmware.
- In alto le radio gestite da questo programma (etichette "Canali" e "Firmware"),
  in basso quelle che si programmano solo con il programma del costruttore (protocollo
  non documentato o radio digitali DMR): per queste ci sono i link ufficiali a
  programmi, firmware e manuali.
- Clic su una radio: si apre la sua scheda dei canali, l'aggiornamento del firmware o
  la lista dei collegamenti. Se la radio esiste in più versioni (es. UV-5RX3, UV-5R+Pro)
  scegli la tua: se non è quella giusta la radio non risponde e non viene modificata.

RADIO GESTITE CON I DRIVER DI CHIRP
- Per 61 modelli il protocollo via cavo e la mappa della memoria sono quelli dei
  driver di CHIRP (https://chirpmyradio.com), riscritti nel linguaggio di questo
  programma. Famiglie: UV-5R (UV-5R, F-11, UV-82, UV-82HP, UV-82X3, UV-6, BF-F8HP,
  UV-5RX3, GT-5R, UV-5G, UV-5G Pro), UV-17 Pro (UV-17Pro, UV-17Pro GPS, UV-21Pro V2,
  UV-21Pro GPS, UV-25, UV-28 Plus, UV-32, UV-5RM, UV-5RM Plus, K5 Plus, UV-5RH, GM-5RH,
  GM-21, UV-5G Plus, UV-17R Plus, BF-F8HP Pro, K6/K63, UV-5R Mini, UV-5G Mini),
  impermeabili (BF-A58, BF-A58S, UV-82WP, GT-3WP, UV-S9X3, UV-5RX, UV-9R, UV-9G, UV-6R,
  UV-5X3), UV-17 e UV-13 Pro, BF-888S/BF-M4/MP31/BF-1901/1904/1909, UV-B5/B6, BF-T1,
  BF-T8/U9, UV-3R, BF-V8A, BF-T20FRS, BF-T20, W31E, Bajeton BJ-7800 e (dalla 3.1)
  Bajeton BJ-7700 (Radtel RT-910 / RT-910 BT), serie H (H-31, H-32, BF-H…: stessa
  radio della Radioddity GA-510) e GM-15 Pro.
- Si lavora sui canali: "Leggi dalla radio" (oppure apri un file .img, anche salvato
  con CHIRP), modifica, "Scrivi nella radio" e/o "Salva" (file .img apribile con CHIRP).
  Le impostazioni della radio restano quelle lette: per questo non c'è "Nuovo file".
- Come per le altre radio: copia di sicurezza prima della scrittura (file .img nella
  cartella copie_radio) e rilettura di verifica dopo.
- Nomi adattati ai caratteri e alla lunghezza ammessi dalla radio, potenze con i
  livelli della radio (es. "Alta 5 W"), DCS solo tra i codici della radio, gamme di
  frequenza controllate. Nella colonna TX "OFF" = canale di sola ricezione.
- Comportamenti particolari ripresi da CHIRP: UV-5G Mini (canali GMRS 1-30 fissi,
  8-14 stretti a bassa potenza), GM-21 (limiti GMRS), BJ-7800 (15 zone da 64 canali;
  canali FM/AM/HF ausiliari non gestiti), UV-5R (numerazione da 0, come la radio),
  GM-15 Pro (radio GMRS: canali 1-30 fissi, 8-14 e 55-250 solo ricezione, 23-54 ponti
  GMRS con TX = RX + 5 MHz), serie H (canali 0-127).
- I driver di CHIRP sono software libero (licenza GNU GPL v3): le parti di questo
  programma derivate da essi sono distribuite con la stessa licenza e il codice
  sorgente completo è nella cartella "sorgenti". Grazie agli autori di CHIRP.

RADIO RICAVATE DAI PROGRAMMI DEI COSTRUTTORI (3.1)
Per queste radio CHIRP non ha un driver: protocollo e memoria sono stati ricavati dal
codice dei programmi ufficiali scaricati dai siti Baofeng e Bajeton, e confrontati con
la famiglia di CHIRP che ha la stessa mappa (i file .img si aprono anche con CHIRP,
con il nome di quel modello).
- UV-25 Plus, AR-152 Pro, UV-29 Pro / SW-9, UV-7B (programma «BF-UVs CPS»): stessa
  memoria e stesso protocollo della UV-17 Pro GPS; la radio deve rispondere con il suo
  modello (es. «UV25PRO», «AR152»), altrimenti lettura e scrittura non partono.
- UV-17R, UV-21R, UV-21L (programma «T6UV Series CPS»): come la UV-17 Pro.
- T22 (PMR446, versione PMR del programma): come la W31E di CHIRP, 16 canali.
- BF-C2 / BF-V2 e BF-88ST (programmi «BF_V1_FH» e «BF-889»): come la BF-888S, 16
  canali senza nome.
- BF-88A / BF-88E: il pacchetto del costruttore dice di usare CHIRP come «BF-888»: si
  usa il driver della BF-888S.
- Bajeton BJ-9000 (Radtel RT-950 Pro, programma «BJ-9000_CPS» v1.1.1): 960 canali in
  15 zone da 64, stesso protocollo della BJ-7800 con il controllo del modello; si
  scrivono solo canali e nomi (impostazioni e APRS con il programma del costruttore).
Non gestite (programmi senza protocollo ricostruibile e radio non in CHIRP): BF-9700
(programma nativo Wp970), serie BF-D e V1D (app «intercom-configuration»): link ai
programmi ufficiali nella galleria.

AGGIORNAMENTO DEL FIRMWARE BAOFENG
- Baofeng UV-5RM / UV-5RH (serie con display TFT "330L"), file .BF: protocollo del
  programma del costruttore "BF Upgrade Tool for TFT 330L Series" (caricatore KDH
  V1.0). Due modi, come nell'originale: automatico (radio accesa: viene portata nel
  caricatore; prima, a scelta, se ne salva una copia dei canali) e manuale (radio
  accesa tenendo premuto il PTT, schermo spento: anche per una radio che non si
  avvia). ATTENZIONE: esistono versioni con chip radio diversi (BK4819, AT1846S,
  FD6818) e i firmware NON sono intercambiabili; il nome del file di solito lo indica
  (es. 5RH_BK_V0.14_231123.BF). Il programma controlla il file (intestazione, aree,
  programma ARM valido, parte cifrata coerente) e riconosce i firmware raccolti dalla
  comunità (impronta SHA-256); Baofeng non li pubblica sul suo sito.
- Baofeng 5RH Pro (e radio con lo stesso CPS "5RH_PRO", come il K67 Pro), file .dat
  "BF_5RH": protocollo del programma di aggiornamento del costruttore (Radio.exe).
  La radio va messa in modalità aggiornamento come indicato dal costruttore; all'inizio
  dichiara la sua elettronica (V1 = firmware 1.x, V2 = firmware 2.x) e se non è quella
  del file l'aggiornamento non parte. Dopo la cancellazione della memoria non si può
  più interrompere.
- Bajeton BJ-7800: stesso programma ("BinTe ANC Bootloader") e stessi file .anc della
  BJ-8300 (firmware ufficiale BJ7800_CN_EN_V0.17_20251126.anc, sul Google Drive del
  costruttore); il modello è riconosciuto dal nome del file. La copia di sicurezza
  prima dell'aggiornamento è un file .img (driver della BJ-7800).
- Dalla 3.1, con i protocolli dei programmi di aggiornamento ufficiali (classe
  «KDH_Bootloader», ricavata dal loro codice):
    UV-25 Plus, file .kdhx («Baofeng Bootloader Pro»; ufficiale V0.23 del 22/05/2025);
    UV-9G, file .kdh («Bootloader_BaudRate38400», aggiornamento per trasmettere sul
      canale 15; ufficiale T52GMRS_M8_V02_231004.kdh);
    Bajeton BJ-9000, file .Firmware («BJ9000_EnUPDATE»; ufficiale V0.27 del 13/03/2026;
      la radio controlla anche il modello scritto nel file).
  Modo automatico (radio accesa) o manuale (radio accesa con i tasti indicati dal
  costruttore, mostrati nella finestra: UV-25 Plus SK2+SK3, UV-9G tasti 12+13, BJ-9000
  tasti laterali 3+4). Il file viene controllato (programma ARM valido, modello nel
  nome, impronta SHA-256 dei firmware ufficiali).
- Bajeton BJ-7700 con Bluetooth: file .anc come BJ-8300/BJ-7800 (ufficiale
  BJ7700PRO_V0.06_251124.anc); la radio viene identificata con il protocollo della
  BJ-7700 e la copia di sicurezza è un file .img.
- BJ-7700 senza Bluetooth (file .BTU), BJ-8200, DM-5000: aggiornamento solo con i
  programmi del costruttore (link nella galleria; formati non documentati).

LETTURA E SCRITTURA VIA CAVO
- Collega il cavo, accendi la radio, scegli la porta COM del cavo nella finestra
  (il nome del driver aiuta a riconoscerla, es. "USB-SERIAL CH340 (COM3)").
  Se la porta non compare, installa il driver del cavo e premi "Aggiorna".
- Il protocollo è quello dei programmi originali dei costruttori (CPS del K67 Pro,
  5RH_PRO, e BT-11_CPS della Bajeton), ricavato dal loro codice: stessi comandi,
  stessa cifratura, stessi controlli, stesse pause.
    K67 Pro: 19200 baud (la radio può chiedere da sola 115200, come nell'originale),
             password di programmazione facoltativa (max 8 caratteri).
    Bajeton 8300: 57600 baud.
- Prima di scrivere, di serie, il programma legge la radio e ne salva una copia di
  sicurezza; dopo la scrittura rilegge la radio e confronta i canali.
  Copie e letture si trovano in %APPDATA%\GestioneCanaliRadio\copie_radio
  (si aprono come normali file .xlc/.dat). Se la copia di sicurezza non riesce,
  non viene scritto nulla.
- Viene scritta tutta la memoria come fa il programma originale: insieme ai canali
  anche le impostazioni generali contenute nel file o nella lettura su cui lavori
  (menu, VFO, zone, DTMF…). Per non cambiarle, parti da una lettura della radio.
  Sul K67 i primi 128 byte (dati della radio, modello, gamme) non vengono mai
  sovrascritti: il programma li legge dalla radio, come l'originale.
- Controlli prima di scrivere (come nell'originale): K67 con banda di frequenza
  diversa da quella del file o modello "BF-K6" non letto prima → scrittura
  rifiutata; dati non convertibili → errore prima di collegarsi alla radio.
- Lettura e copia di sicurezza si possono interrompere; la scrittura no (la radio
  resterebbe con dati incompleti). Se la scrittura si interrompe (cavo staccato,
  radio spenta) ripetila: la copia di sicurezza è nella cartella indicata sopra.
- Dopo una lettura i canali risultano "non ancora salvati in un file".
- Quando la radio è collegata al PC non usare il programma originale nello stesso
  momento (la porta COM può essere aperta da un solo programma).

AGGIORNAMENTO DEL FIRMWARE (BAJETON 8300)
Nella scheda "Bajeton 8300" premi "⬆ Aggiorna firmware…". La finestra ha tre passi.
1. Firmware
   - "File sul PC…": scegli il file del firmware (.anc, es. BJ8300_V0.12_20251112.anc)
     oppure un archivio .zip o .rar che lo contiene, anche con archivi uno dentro
     l'altro (come il pacchetto del costruttore: .rar con dentro uno .zip con il .anc).
   - "Da Internet": incolla l'indirizzo del file (.anc, .zip o .rar) e premi "Scarica".
     Funzionano anche i link di condivisione di un singolo file di Google Drive e di
     Dropbox. Un link a una CARTELLA di Google Drive non si può scaricare: apri la
     cartella nel browser e copia il link del file, oppure scaricalo e caricalo dal PC.
     I firmware scaricati vengono conservati in %APPDATA%\GestioneCanaliRadio\firmware.
   - Se negli archivi ci sono più firmware, il programma chiede quale usare.
   - Controlli sul file: formato (firma "JLUFW" dei firmware della radio), decifratura
     (il file .anc è cifrato: il programma lo decifra come quello originale e controlla
     che il contenuto sia leggibile), piattaforma uguale a quella del firmware ufficiale
     della BJ-8300 (YHD15C8_soundbox), nome del file, impronta SHA-256 dei firmware
     ufficiali conosciuti. Il firmware BJ8300 V0.12 del 12/11/2025 (l'ultimo pubblicato
     da Bajeton) è riconosciuto come ufficiale. Un file non valido non può essere inviato;
     un file con segnalazioni (es. piattaforma diversa) richiede una conferma in più.
2. Radio
   - Scegli la porta del cavo e lo stato della radio:
     "Funziona normalmente" (come il pulsante "Update" del programma originale): prima
       di tutto la radio viene IDENTIFICATA con il protocollo di programmazione della
       Bajeton 8300 (lo stesso di "Leggi dalla radio") e, di serie, ne viene letta e
       salvata una copia dei canali e delle impostazioni (Bajeton_prima_del_firmware_….dat
       in copie_radio). Se la radio non risponde come una Bajeton 8300 l'aggiornamento
       NON parte e la radio non viene toccata. Poi la radio viene portata in modalità
       aggiornamento e riceve il firmware.
       "Identifica la radio" fa solo il controllo, senza aggiornare.
     "Non si avvia" (modalità di recupero, come il pulsante "Click Here If Your Radio
       Can't Work"): per una radio rimasta senza firmware valido. La radio non può essere
       identificata: il programma, come l'originale, aspetta circa 12 secondi che la radio
       si presenti in modalità aggiornamento. Premi "Avvia aggiornamento" e porta subito
       la radio in modalità aggiornamento secondo le istruzioni del rivenditore (di solito
       accendendola con un tasto premuto, spesso il PTT: la procedura non è documentata).
   - Velocità del trasferimento: 115200 baud (predefinita del programma originale) o 57600.
3. Conferma
   - Va sempre spuntato "Confermo che la radio collegata è una Bajeton BJ-8300".
   - Quando la radio non viene identificata dal programma (modalità di recupero, oppure
     "salta l'identificazione automatica") bisogna anche scrivere BJ8300 nella casella.
Durante l'aggiornamento una finestra mostra avanzamento e messaggi. Non spegnere la
radio e non scollegare il cavo; usa una batteria carica. L'operazione si può
interrompere solo finché la radio non ha iniziato a rispondere. Alla fine spegni e
riaccendi la radio e controlla la versione all'accensione; se canali o impostazioni
fossero cambiati, scrivi nella radio la copia salvata prima dell'aggiornamento.
Se l'aggiornamento si interrompe: spegni e riaccendi la radio e ripetilo; se la radio
non si avvia più, ripetilo in modalità di recupero.
Protocollo: quello del programma del costruttore "BinTe ANC Bootloader" V1.1 (BinTalk),
ricavato dal suo codice: pacchetti AA 55 con CRC16, 9600 baud, comando di ingresso in
modalità aggiornamento, passaggio a 10000 baud e poi alla velocità scelta, invio dei
dati richiesti dalla radio, 5 nuovi tentativi a intervalli di 2 secondi. Il programma
originale non controlla né il file né il modello della radio: i controlli descritti
sopra sono aggiunti da questo programma.
Gli archivi .rar (formati RAR 1.5-5, anche "solidi") sono aperti dal programma con il
decodificatore libero rardecode (licenza BSD). Non sono supportati gli archivi protetti
da password e quelli divisi in più parti.

SCHEDA INFO E AGGIORNAMENTI
- La scheda "ℹ Info" mostra nome, versione, data di compilazione, autore, e-mail,
  pagina QRZ, licenza e note legali (i collegamenti si aprono nel browser / nel
  programma di posta).
- Verifica degli aggiornamenti: a ogni avvio, in automatico, e con il pulsante
  "🔄 Aggiornamenti" in alto (o "🔄 Verifica aggiornamenti" nella scheda Info).
  Il programma confronta la sua versione con quella pubblicata dall'autore:
    · se quella in uso è la più recente lo dice e non scarica nulla;
    · se ne esiste una più recente compare in alto un avviso giallo con
      "⬇ Scarica l'aggiornamento" (e "Dettagli"); la × lo nasconde fino al
      prossimo avvio.
  Il controllo legge solo il numero di versione (pochi KB), non il programma.
  Senza connessione a Internet il controllo all'avvio non mostra errori.
- "Scarica l'aggiornamento": scegli dove salvarlo (di solito la cartella Download),
  il pacchetto .zip viene scaricato e controllato. Poi chiudi il programma, estrai lo
  zip e sostituisci l'eseguibile: archivio ponti e impostazioni restano dove sono.
- L'indirizzo da cui si verificano e scaricano gli aggiornamenti è fisso nel programma
  e non si può modificare.
- Per l'autore, pubblicare una nuova versione: aumentare appVersion (server.go) e la
  versione in res/risorse.rc, ricompilare, aggiornare versione.txt (righe
  "versione=x.y" e "compilazione=...") e sostituire il file "IU0VYT.Flash.zip" nel
  rilascio "NEW" del repository GitHub (che deve essere pubblico). La versione viene
  letta da versione.txt, oppure dalla riga "Versione x.y" del LEGGIMI.txt.

TABELLE DEI CANALI
- Doppio clic, Invio o F2 per modificare una cella; puoi anche iniziare a scrivere.
  Esc annulla la modifica. Su una riga vuota si comincia scrivendo la frequenza RX.
- Frequenze in MHz (145.600 oppure 145,600). Lo shift si scrive -0.6, +1.6, +5:
  il TX viene calcolato. Per un TX in un'altra banda scrivi direttamente la frequenza TX.
- Toni: subtoni CTCSS standard (67.0 … 254.1) oppure codici DCS: D023N normale,
  D023I invertito (D023 = normale); vuoto = OFF. "Tono TX" apre il ponte,
  "Tono RX" è lo squelch a tono in ricezione. Il K67 accetta qualsiasi codice DCS
  ottale; la Bajeton i 104 codici standard (D023 … D754).
- Scan: canale incluso o no nella scansione (ora anche sul K67).
- Selezione: clic, Maiusc+clic (intervallo), Ctrl+clic (aggiungi), Ctrl+A (tutti).
- Pulsanti e tasti: Aggiungi; Inserisci riga (Ins); Cancella (Canc);
  Elimina e compatta (Maiusc+Canc); Sposta su/giù (Alt+↑ / Alt+↓);
  Ordina (per frequenza, nome, tono: se hai selezionato più canali ordina solo quelli);
  Compatta (toglie i vuoti); Copia/Taglia/Incolla (Ctrl+C/X/V); Annulla/Ripeti (Ctrl+Z/Y).
  Il tasto destro sulla tabella apre un menu con le stesse operazioni.
- Incollando su posizioni occupate il programma chiede se sovrascrivere, inserire
  spostando in basso o usare solo le posizioni libere.
- Bajeton 8300: 15 zone da 64 canali. Scegli la zona dalla barra sopra la tabella,
  "Tutte" per vederle insieme; "Sposta in zona" sposta i canali selezionati.
- K67 Pro: 640 canali. Le gamme di frequenza della radio sono lette dal file:
  le frequenze fuori gamma sono evidenziate in rosso.

COPIA TRA LE DUE RADIO
"Copia in Bajeton… / Copia in K67…" (oppure Copia nella prima scheda e Incolla
nell'altra) adatta i canali con la logica del convertitore e mostra gli avvisi:
- nomi: max 16 caratteri sul K67, 12 sulla Bajeton (abbreviazioni M. S. D. e
  numero finale conservato, es. "CAMPOCATINO 3" → "CAMPOCATIN 3");
- potenza: la Media della Bajeton diventa Alta sul K67;
- toni DCS: trasferiti se la radio di destinazione ha quel codice, altrimenti OFF
  con avviso;
- frequenze fuori dalle gamme del K67.
Nella finestra di inserimento scegli la zona (Bajeton), la posizione di partenza e
se usare solo le posizioni libere o sovrascrivere.

COSA VIENE CONSERVATO
Il programma modifica solo ciò che cambi. Un file aperto e salvato senza modifiche
resta identico byte per byte. I dati che il programma non interpreta (altri byte del
record K67, i campi segnalazione/PTT ID/busy lock/ecc. della Bajeton) restano invariati
e seguono il canale quando lo sposti o lo copi nella stessa radio. Le impostazioni
generali delle radio non vengono toccate.
K67: quando aggiungi o cancelli canali, le liste delle zone, i canali A/B e il canale
di emergenza vengono aggiornati come fa il programma originale (le zone non si
modificano da qui).

CORREZIONE DI UN ERRORE DELLA VERSIONE 1.0 (K67 Pro)
Analizzando il programma originale si è scoperto che il file .xlc ha due tabelle di
bit per canale ("posizione occupata" e "in scansione"). La versione 1.0 usava tabelle
sbagliate: un canale aggiunto in una posizione che era VUOTA poteva risultare
invisibile alla radio e al programma originale (modifiche e cancellazioni di canali
esistenti invece erano corrette). Se hai salvato file .xlc con la 1.0 aggiungendo
canali in posizioni vuote, aprili con questa versione: le posizioni interessate
vengono segnalate e basta reinserire quei canali (oppure ripeti l'aggiunta).

ARCHIVIO PONTI
- È condiviso dalle due radio e si salva da solo, sul PC, in:
  %APPDATA%\GestioneCanaliRadio\archivio_ponti.json (copia di sicurezza: .bak)
- "Importa…": da Internet (indirizzo preimpostato https://www.ik2ane.it/pontixls.xls,
  modificabile; gli indirizzi usati vengono ricordati) oppure da file .xls, .xlsx, .csv.
  Riconosce l'elenco di IK2ANE (foglio "ponti"), i CSV di CHIRP, i file esportati da
  questo programma e fogli generici con intestazioni come Frequenza, Shift, Tono, Nome,
  Località, Regione, Provincia, Tipo, Rete.
- Valori irregolari interpretati con buon senso e segnalati (righe gialle con ⚠):
    "+5.0MHz", "-600 kHz"         → shift normali
    "+1.362,5 MHz"                → 1,3625 MHz (segnalato)
    "79,7" "123"                  → 79.7 / 123.0
    "77.0-141.3" "110.9/162.2 loc"→ usato il primo tono, gli altri indicati nell'avviso
    "82.2" (non standard)         → 82.5, segnalato
    "CC1" (color code DMR)        → nessun subtono, annotato nelle note
    "DCS023", "1750", "DG-ID:10"  → annotati/segnalati
    "input" / "output"            → isofrequenza, segnalato
  Il valore originale resta visibile nel pannello di dettaglio. Le righe senza una
  frequenza valida (es. i segnaposto 9999999 e le righe di testo finali) non sono
  importabili: sono elencate nella finestra, non scartate in silenzio.
- Dal file di IK2ANE vengono letti anche i colori della legenda: ponti non più attivi,
  momentaneamente non in funzione, non H24, tono non regionale, shift non standard.
- Filtri (in importazione, in esportazione e nell'archivio): banda (VHF/UHF/Altre),
  tipo (Analogico, EchoLink, SvxLink, DMR, C4FM, D-Star, NXDN, Multimodo, HotSpot,
  Transponder, Winlink), stato, regione, provincia, rete, testo libero, solo righe con
  avvisi; più la spunta riga per riga. In importazione i filtri partono da "solo ponti
  utilizzabili in analogico, esclusi i non attivi": puoi cambiarli.
- Nome del canale: modello a scelta con {LOC} {SIGLA} {NOMINATIVO} {PROV}
  (es. "{NOMINATIVO} {LOC}" → "IR0CM M.CAVO"). "Rigenera nomi…" lo riapplica.
- Ponti già presenti (stessa frequenza, TX, località e sigla): puoi aggiungere solo i
  nuovi, aggiornare quelli presenti (il nome che hai dato resta) o sostituire tutto.
- Seleziona i ponti e usa "→ K67…" / "→ Bajeton…" per inserirli nelle radio.
  "Segna come verificati" toglie gli avvisi dopo il controllo (restano nelle note).

ESPORTAZIONE
Dall'archivio (con filtri e spunte) o dai canali di una radio (tutti o selezionati):
- Excel .xlsx e Excel 97-2003 .xls (prima riga bloccata, righe con avvisi in giallo);
- CSV per Excel italiano (separatore ";" e virgola decimale);
- CSV per CHIRP: colonne standard di CHIRP. La potenza è scritta in watt
  (Alta 5.0W, Media 2.0W, Bassa 1.0W) come si aspettano le versioni recenti di CHIRP;
  i toni usano Tone (solo TX), TSQL (TX e RX uguali) o Cross (diversi).
  CHIRP adatta poi la potenza ai livelli della radio di destinazione.

DIFFERENZE RISPETTO AL CONVERTITORE PRECEDENTE
- Il K67 ora conserva i byte non interpretati di ogni canale invece di azzerarli.
- CSV CHIRP: potenza in watt al posto di High/Mid/Low (rifiutati da CHIRP recente) e
  toni diversi in TX e RX scritti come Cross (prima TSQL, che in CHIRP usa un solo tono).

COME È STATO VERIFICATO
Il codice dei due programmi originali è stato decompilato e fatto girare su radio
simulate, accanto a questo programma:
- scrittura: memoria della radio identica byte per byte e stessa sequenza di comandi
  (K67: 17 comandi, 49.152 byte; Bajeton: 505 comandi, 49.984 byte), sia con i file di
  esempio sia con file modificati (spostamenti, cancellazioni, nuovi canali, DCS);
- lettura: file identico a quello che salva il programma originale (data a parte);
  Bajeton: tutti i 15.600 campi del .dat uguali;
- il programma originale riapre e risalva i file di questo programma senza altre
  modifiche;
- password errata, radio spenta, banda diversa, rifiuto dei dati, interruzioni,
  copia di sicurezza e verifica sono provati con test automatici.
NON È STATO POSSIBILE PROVARLO CON LE RADIO VERE. Alla prima prova: parti da una
lettura, salvala in un file e tienila da parte; poi prova a scrivere.
Aggiornamento del firmware (2.1): il codice ORIGINALE del programma "BinTe ANC
Bootloader" (caricato direttamente dal suo eseguibile) e questo programma sono stati
collegati alla stessa Bajeton simulata in 10 situazioni (aggiornamento a 115200 e a
57600 baud, modalità di recupero, file di lunghezza qualsiasi, risposte perse, CRC
errati, cambio di velocità perso, radio spenta, firmware vero V0.12 con 700 KB di dati):
la radio simulata ha ricevuto in tutti i casi gli stessi pacchetti, alle stesse velocità
e con gli stessi dati, e gli esiti sono identici (stessi tentativi, stessi errori).
L'eseguibile Windows è stato provato anche sotto Wine con porte seriali vere
(pseudo-terminali): aggiornamento e recupero completati, cambi di velocità corretti.
Anche l'aggiornamento del firmware NON è stato provato con una radio vera: il
comportamento della radio è ricostruito da ciò che il programma originale si aspetta.

Novità della 3.0, stesse verifiche:
- driver di CHIRP: CHIRP vero (cartella dei sorgenti) e questo programma confrontati
  su ogni modello: lettura dei canali dalle immagini di prova di CHIRP (tutti i canali
  uguali), fino a 18 modifiche per modello (nuovi canali, ponti, toni CTCSS/DCS normali e
  invertiti, split, sola ricezione, cancellazioni, ultima posizione) con immagine
  risultante identica byte per byte a quella di CHIRP, posizione di ogni campo della
  mappa uguale, e lettura/scrittura via cavo con radio simulate: byte inviati alla
  radio identici a quelli di CHIRP e memoria finale identica;
- firmware .BF: il codice ORIGINALE del "BF Upgrade Tool for TFT 330L Series" (caricato
  dal suo eseguibile) e questo programma collegati alla stessa radio simulata in 9
  situazioni (automatico, manuale, risposte perse, errori di controllo, CRC errati,
  presentazione persa, errore di scrittura, radio spenta) e con 7 firmware veri:
  pacchetti e esiti identici;
- firmware 5RH Pro: il codice originale di aggiornamento (Radio.exe) e questo programma
  confrontati con la stessa radio simulata (elettronica V1 e V2, 115200 e 38400 baud,
  elettronica diversa dal file, blocco rifiutato) e con 6 firmware veri: identici;
- interfaccia provata in un browser automatico con le radio simulate (galleria, lettura,
  modifica, scrittura con copia e verifica, copia tra radio, firmware).
Novità della 3.1, stesse verifiche:
- BJ-7700, serie H, GM-15 Pro: confronto completo con CHIRP come sopra (lettura,
  modifiche, mappa, protocollo via cavo: identici); per la GM-15 Pro anche le regole dei
  canali GMRS;
- radio ricavate dai programmi dei costruttori: lettura, modifica e scrittura con radio
  simulate costruite sul codice dei programmi ufficiali (memoria finale uguale a quella
  scritta; una radio che risponde con un altro modello viene rifiutata);
- firmware UV-25 Plus, UV-9G e BJ-9000: il codice ORIGINALE dei tre programmi di
  aggiornamento (caricato dai loro eseguibili) e questo programma collegati alla stessa
  radio simulata, in modo automatico e manuale, con risposte perse ed errori, e con i
  firmware ufficiali: pacchetti, velocità ed esiti identici.
Anche queste parti NON sono state provate con radio vere.

LIMITI
- I protocolli e i formati sono proprietari e ricostruiti dai programmi originali:
  conserva sempre le copie di sicurezza.
- Le zone del K67 non si modificano da qui (vengono solo tenute coerenti).
- Non si sa se il K67 Pro possa identificarsi come "BF-K6": i relativi controlli
  sono riprodotti come nel programma originale.
- Il download da Internet usa la connessione diretta (non il proxy di sistema).
- Radio gestite con i driver di CHIRP: valgono i limiti di CHIRP per quel modello (es.
  BJ-7800 scritto per il firmware V0.15P; se la radio ha un firmware più recente e non
  risponde come previsto, la lettura si ferma senza modificare la radio).
- Firmware: la procedura per far entrare in modalità aggiornamento una radio bloccata
  non è documentata dal costruttore. L'identificazione automatica riconosce le radio
  che usano il protocollo di programmazione della Bajeton 8300 (BT-11): una radio
  diversa con lo stesso protocollo non sarebbe distinguibile, per questo serve anche
  la tua conferma.

RICOMPILARE DAI SORGENTI (facoltativo)
Serve Go 1.22 o successivo. Nella cartella "sorgenti":
  go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w -X main.buildDate=AAAA-MM-GGTHH:MM" -o "IU0VYT Flash.exe" .
(buildDate è la data di compilazione mostrata nella scheda Info, es. 2026-10-20T21:45).
Icona e proprietà del file .exe (nome, versione) sono in res/risorse.rc, compilato in
rsrc_windows_amd64.syso (il comando è scritto in testa al file .rc).
(le librerie sono già incluse nella cartella vendor; l'icona è in rsrc_windows_amd64.syso).
Su Linux "go test -mod=vendor ." esegue i test automatici sui file di esempio in testdata
e con le radio simulate. La cartella "verifica" contiene i programmi di confronto
con il codice originale (vedi verifica/LEGGIMI_VERIFICA.txt).
Le prove sull'importazione dell'elenco dei ponti di IK2ANE richiedono il file
testdata/pontixls.xls (non incluso: dati di terzi, si scarica da www.ik2ane.it).

LICENZA
IU0VYT Flash, Copyright (C) 2026 Daniele "IU0VYT" Bonifazi. Software libero: puoi
ridistribuirlo e/o modificarlo secondo i termini della GNU General Public License
pubblicata dalla Free Software Foundation, versione 3 o (a tua scelta) successiva.
È distribuito nella speranza che sia utile, ma SENZA ALCUNA GARANZIA, neppure
implicita di COMMERCIABILITÀ o di IDONEITÀ A UN FINE PARTICOLARE. Testo completo in
LICENSE.txt. Chi lo ridistribuisce (anche modificato) deve usare la stessa licenza,
fornire il codice sorgente e conservare le note di copyright.
I driver delle radio derivano in gran parte da CHIRP (GNU GPL): autori e licenze di
tutti i componenti in AUTORI_E_LICENZE.txt.
Programma indipendente, non affiliato né approvato da Baofeng o da Bajeton: i nomi
dei modelli indicano solo con quali radio funziona; i marchi appartengono ai
rispettivi titolari.

