# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Pagina della release su GitHub (in inglese), con il download: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 per Microsoft Windows 10/11

Questa versione comprende anche le novità della 4.8.0, della 4.7.0 e della 4.6.0, pubblicate solo come versioni di prova.

## Novità della 4.9.0

### Corretto: Identifica con Lanchonlh HG-UV98 e Quansheng TG-UV2+
- Durante «Identifica» la porta parlava sempre in 8N1. Ora usa la parità pari (HG-UV98) e i due bit di stop (TG-UV2+), come la lettura dalla scheda della radio. Le due radio ora vengono riconosciute.

### Corretto: parità della HG-UV98
- Con la parità pari la porta di Windows ora controlla anche gli errori di parità. Prima il controllo veniva acceso e subito spento.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus con i firmware liberi
- Nuove schede per i firmware egzumer, F4HWN e LOSEHU/LTS/UVE.
- F4HWN usa il driver di armel F4HWN (fuori da CHIRP): 8 livelli di potenza, blocco della trasmissione, 3 liste di scansione e le impostazioni del firmware.
- Ogni scheda legge e scrive solo il suo firmware. Dopo l'aggiornamento del firmware il programma indica la scheda da usare.

### Nuove radio
- Wouxun (20 modelli): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (USA)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (firmware Radtel e firmware libero M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

Per le radio dai driver di CHIRP, canali e impostazioni sono verificati con le prove automatiche contro CHIRP. Le radio GMRS, FRS e MURS sono per gli Stati Uniti: in Italia non sono ammesse in trasmissione su quelle frequenze.

## Novità della 4.8.0

### Aggiornamento del firmware .BTU (Radtel RT-910 e RT-900 senza Bluetooth, Bajeton BJ-7700 senza Bluetooth)
- Si aggiornano i firmware .BTU con il protocollo del programma «Radtel Bootload Tools». Il caricatore è quello della Baofeng UV-5RM, con la presentazione «PROGRAMBT80U».
- Modalità automatica, con la radio accesa normalmente. Modalità manuale: radio spenta, accesa tenendo premuti i due tasti laterali, finché il display mostra «UPDATE».
- Il caricatore non controlla il modello. Per questo il programma riconosce dal nome del file la radio e la versione dell'elettronica, mostra gli avvisi e chiede sempre di scrivere il modello.
- In galleria RT-910 (senza Bluetooth, .BTU) e RT-910B (con Bluetooth, .anc) ora hanno due schede. Lo stesso vale per BJ-7700 e BJ-7700 senza Bluetooth.

### Corretto: Radtel RT-900 senza Bluetooth con il firmware da 512 canali
- Fino alla V1.09P la radio ha 512 canali; dalla V1.14P ne ha 999.
- Ora i canali 513-999 si scrivono solo se li modifichi. Prima il programma controlla nella radio che quell'area sia vuota o contenga canali.

### Nuove radio (driver di CHIRP)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (la nuova versione della H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

Per tutte, canali e impostazioni sono verificati con le prove automatiche contro CHIRP.

## Novità della 4.7.0

### Corretto: identificazione automatica bloccata o lentissima
- Una radio accesa in modalità aggiornamento del firmware trasmette da sola dei messaggi periodici: prima sembrava rispondere a ogni modello e la ricerca durava minuti. Ora il programma ascolta la porta prima di provare e non si fa ingannare: la ricerca finisce in meno di un minuto e la scheda Identifica spiega cosa succede.
- L'interfaccia non si blocca più mentre partono le ricerche; le radio più diffuse si provano per prime.
- Corretto il falso avviso «Scheda Radio già aperta» per una scheda chiusa in precedenza.

### Radio con la stessa gestione
- La finestra di scelta e la scheda Identifica elencano tutte le radio possibili (es. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), con il menu per scegliere e una breve indicazione su dove vedere la versione del firmware sulla radio.

### Corretto: Radtel RT-950 Pro / Bajeton BJ-9000 con il firmware fino a 0.23
- Il bit «trasmissione abilitata» dei canali ora si scrive sempre, con tutte e due le versioni (come i programmi Radtel fino alla v1.2.x; la v1.3.3 lo ignora). Con il firmware vecchio i canali modificati nella scheda da 990 canali potevano non trasmettere. Il bit viene rimesso anche nei canali 1-960 che ne erano privi.
- Verificate le versioni reali della Bajeton BJ-9000: firmware V0.22, V0.27 e V0.29 (identico a quello Radtel), programma v1.1.1 (960 canali) e v1.3.3 (990 canali, è il programma Radtel). Canali e impostazioni verificati anche con il codice dei programmi Bajeton.

### Aggiornamento del firmware per altre radio
- Radtel RT-950 Pro, Radtel RT-950 e Hiroyasu HI-UV850 Pro (file .BTF, stesso caricatore della BJ-9000).
- Radtel RT-920 e RT-910B (file .anc, stesso protocollo di BJ-7800 e BJ-7700).

### Nuove radio (driver di CHIRP)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); in galleria anche Hiroyasu HI-UV850 Pro.
- Impostazioni anche per SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 e Anysecu AC-580.

## Novità della 4.6.0

### Corretto: Radtel RT-950 Pro / Bajeton BJ-9000 con il firmware recente
- Dal firmware **0.24** (dicembre 2025) la radio ha **990 canali in 10 zone da 99** (oppure un unico elenco di 990); fino alla **0.23** ne aveva **960 in 15 zone da 64** (oppure un unico elenco di 960). Il programma ora gestisce le due versioni.
- La versione si riconosce dai dati letti dalla radio (la radio non la comunica): la lettura apre i canali nella scheda giusta. Se i dati non bastano, una nota spiega come controllare la versione sulla radio.
- I canali 961-990 si scrivono solo se cambiano, dopo averli riletti dalla radio; se la radio non li ha (firmware fino a 0.23) non si scrive nulla.
- Canali e impostazioni verificati con il codice originale dei programmi Radtel v1.2.0 (960 canali) e v1.3.3 (990 canali).

### Nuova: Radtel RT-950 (senza «Pro»)
- 960 canali (15 zone da 64) o un unico elenco, impostazioni generali; verificata con il codice originale del programma «BT-9000_CPS» v1.3.0.

### Corretto: nome esatto della radio
- Radio diverse con la stessa gestione (es. **Radtel RT-950 Pro** e **Bajeton BJ-9000**, **Baofeng K67 Pro** e **5RH Pro**, Radtel RT-910 e Bajeton BJ-7700) ora si chiamano con il loro nome in scheda, lettura, scrittura, identificazione e nei file di copia.
- Alla prima identificazione di una di queste radio il programma chiede quale sia e poi lo ricorda.

### Corretto: avviso di dati non salvati
- Dopo una lettura, una scrittura o un inoltro i canali sono nella radio: chiudendo il programma l'avviso compare solo se ci sono modifiche vere.

### Nuove radio (driver di CHIRP)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: i 990 canali in 10 zone da 99, come sulla radio.

Galleria: 489 modelli di 44 marche (451 con canali, 386 con impostazioni, 52 con l'aggiornamento del firmware). Nuove marche: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
