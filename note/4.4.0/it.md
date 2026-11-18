# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 per Microsoft Windows 10/11

Questa versione comprende anche le novità della 4.3.0, 4.2.0, 4.1.0 e 4.0.0, che non hanno avuto una release stabile propria.

## Novità della 4.4.0


### Corretto: riquadro «Identificazione Radio In Corso» senza radio
- Il riquadro poteva comparire anche con un cavo senza radio: i disturbi della linea e l'eco dei comandi inviati sembravano una radio che risponde.
- Ora compare **solo per il cavo a cui è collegata una radio che risponde davvero**. Con due cavi e una sola radio (COM6 senza radio, COM7 con la radio, o il contrario) compare solo per la porta della radio.
- Senza radio il giro delle prove è anche più veloce.

### Barra in alto, Home, scheda Identifica
- Il pulsante si chiama **«🔎 Identifica Radio»** ed è a destra, accanto ad «Aggiornamenti».
- Con molte schede aperte il pulsante si restringe da solo, fino alla sola lente.
- Home: tornano le descrizioni dei riquadri della 4.2. In basso restano visibili copyright, sito web e sorgenti, anche a 1280×720.
- Scheda Identifica: spiegazioni brevi e chiare.

### Firmware Retevis (43 modelli con l'aggiornamento del firmware)
- **Retevis RA79 V3** (e UV-K5 V3 / UV-K1, caricatore 7.x) con i file .bin di Retevis, per esempio «RA79 V3.bin». Il firmware viene inviato solo se il caricatore della radio è della stessa famiglia.
- **Retevis RA89** con i file .bin di Retevis, per esempio «Eng_RA89_Upadat_20250524_B91.38.bin». Usa il protocollo del programma di Retevis «RA89-update-software», che è stato fatto girare su una radio simulata. I comandi inviati sono identici byte per byte.

### Altro
- I test con le radio simulate non impediscono più la compilazione dei test su Windows.

Galleria: 386 modelli di 27 marche (348 con canali, 302 con impostazioni, 43 con l'aggiornamento del firmware).

## Novità della 4.3.0


### Identificazione automatica: più veloce, più radio, più porte
- Un giro completo di prove senza radio dura circa **50 secondi** (prima 135).
- Il programma **ascolta di continuo tutte le porte dei cavi**, ognuna per conto suo e tutte insieme:
  - riconosce anche una radio collegata spenta e accesa dopo;
  - riconosce **più radio su cavi diversi** nello stesso momento, ne mostra i dati e ne apre tutte le schede.
- Le **radio già riconosciute** (es. K67 Pro, Bajeton 8300, BJ-9000) vengono provate per prime, poi tutte le altre.
- Se la scheda della radio è già aperta compare l'avviso «Radio già identificata!! Scheda Radio già aperta» (con più radio «Radio già identificate!! Schede radio già aperte»).
- Scheda Identifica: porte in ordine (COM6, COM7… COM10) e scritte come «COM6 — USB-SERIAL CH340»; «Identifica Nuova Radio Automaticamente» in grassetto.

### Aggiornamento del firmware: 8 radio in più (41 in tutto)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** con i file .kdhx di Radtel (caricatore «JJCC Bootloader Pro»). RT-470: il firmware 1.x è solo per la scheda vecchia, il 2.x solo per la nuova. Il programma lo indica e chiede conferma scritta.
- **TIDRADIO TD-H3, TD-H8 e Radtel RT-730** con i file .BIN del costruttore: avvia l'aggiornamento e accendi la radio tenendo premuto il PTT.

### Altro
- **Bajeton BJ-9000**: come la BJ-8300, tutti i 960 canali in un unico elenco senza zone («Canali: tutti (960)»).
- **Home** più compatta: in basso si vedono sempre versione, copyright, sito e sorgenti.
- **Galleria**: una linea spessa separa i modelli compatibili da quelli non compatibili.
- **LEGGIMI**: tolta una lunga parte ripetuta per errore nella 4.2.0.

Galleria: 386 modelli di 27 marche (348 con canali, 302 con impostazioni, 41 con l'aggiornamento del firmware).

## Novità della 4.2.0


- Corretta l'identificazione automatica: il riquadro «Identificazione Radio In Corso» compare solo quando una radio collegata risponde (non più all'avvio con il solo cavo) e sparisce quando la radio è riconosciuta o non viene trovata. Dopo il riconoscimento si aprono i dati e la scheda della radio e la ricerca si ferma, senza disattivare la funzione.
- Nuovo pulsante «🔎 Identifica Nuova Radio Automaticamente» nella barra in alto, tra «Info» e «Aggiornamenti»: cerca una nuova radio con le stesse regole ed è grigio se l'identificazione automatica è disattivata.
- 35 radio nuove (36 versioni) e 8 marche in più: famiglia TYT TH-UV88 (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), famiglia Radtel RT-490 (con JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 radio della famiglia Retevis RT21 (con Abbree AR-63) e Quansheng TK11.
- Aggiornamento del firmware anche per Radtel RT-590 / RT-590 Plus e RT-600 / RT-600 Plus con i file .bin di Radtel.
- Corretti alcuni difetti nascosti nella lettura dei nomi dei canali con lettere accentate o cinesi.
- Galleria di 386 modelli di 27 marche: 348 con canali e ponti, 302 con le impostazioni, 33 con l'aggiornamento del firmware.

## Novità della 4.1.0


- Corretta l'identificazione automatica: funziona in qualsiasi scheda e, quando scolleghi la radio, la spegni o la cambi, la ricerca riparte da sola e riconosce la radio collegata dopo (anche se è la stessa o la accendi dopo il cavo). Prima, dopo il primo riconoscimento, non cercava più altre radio.
- Durante la ricerca automatica compare il riquadro «Identificazione Radio In Corso. Attendere prego.», con il pulsante «Interrompi».
- 14 radio nuove e 3 marche in più: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 e RT-730, Hiroyasu HI-8811, Talkpod A36plus e A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (2ª e 3ª generazione) e TD-H3, nelle versioni base, HAM e GMRS.
- Aggiornamento del firmware delle Quansheng UV-K5, UV-K5(8), UV-K6 e UV-5R Plus con i file .bin di Quansheng, con il controllo che il firmware sia per il tipo di radio collegata.
- Schermata Home e sito: «GESTIONE CANALI - FREQUENZE - PONTI RADIO - FIRMWARE» e i nomi di tutte le marche in una cornice in stile videogioco anni '80.
- Galleria di 349 modelli di 19 marche: 311 con canali e ponti, 265 con le impostazioni, 31 con l'aggiornamento del firmware.

## Novità della 4.0.0


- Programma plurimarca: 335 modelli di 16 marche nella galleria, 297 con canali e frequenze via cavo e 251 con le impostazioni della radio.
- Nuove radio: Quansheng UV-K5, UV-K6, UV-5R Plus e UV-K1; 31 modelli Retevis (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920 e T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S e GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton e TID.
- Identificazione automatica (scheda «Identifica»): colleghi il cavo e la radio viene riconosciuta da sola, con la sua scheda aperta e i canali letti. La scelta resta memorizzata e si disattiva dallo stesso pulsante.
- Galleria dei modelli con il menu delle marche e un colore per ogni marca.
- «Leggi dalla radio» e «Scrivi nella radio» in blu nelle schede delle radio.
- Home, Info, guida, sito e documentazione aggiornati al programma plurimarca.
