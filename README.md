# IU0VYT Flash

**Versione 3.9** · Windows · interfaccia in italiano · un solo file eseguibile, nessuna installazione

Programma per radioamatori e utenti di ricetrasmettitori **Baofeng** e **Bajeton**: legge e scrive canali e frequenze via cavo di programmazione, gestisce un archivio dei ponti radio italiani e aggiorna il firmware delle radio supportate.

## Funzioni

- **Canali e frequenze** di 235 modelli Baofeng e Bajeton (comprese radio PMR digitali e DMR, anche via USB): lettura e scrittura via cavo, file `.img` compatibili con CHIRP, file originali del Baofeng K67 Pro (`.xlc`) e della Bajeton 8300 (`.dat`).
- **Impostazioni delle radio** (menu, display, VOX, squelch, TOT, DTMF, tasti…) per 190 modelli, con la descrizione di ogni voce e un avviso da confermare per quelle delicate.
- **Identificazione della radio collegata**: il programma la riconosce da solo e mostra marca, modello, tipo, firmware e le altre informazioni che comunica (per le radio DMR USB anche ID DMR e nome).
- **Convertitore universale**: i canali di qualsiasi radio o file (K67 `.xlc`, Bajeton `.dat`, immagini `.img` anche di CHIRP, CSV di CHIRP, Excel/CSV) nel formato di qualsiasi altra radio, adattati alle sue regole.
- **Galleria di 273 modelli**, divisi tra compatibili e non compatibili, per tipologia (radioamatoriali, DMR, GMRS/FRS, PMR446, per bambini) e per marca, con ricerca, filtri e collegamenti ai programmi e ai manuali ufficiali.
- **Archivio dei ponti radio** basato sull'elenco aggiornato di [IK2ANE](https://www.ik2ane.it), con filtri e inserimento nelle radio, con il controllo della compatibilità tra ponti e radio prima della scrittura.
- **Aggiornamento del firmware** di 28 modelli, con controllo del file e copia di sicurezza della radio; aggiornamento guidato con il programma ufficiale per BTECH DMR-6X2 e DMR-6X2 PRO.
- **Copia di sicurezza** automatica prima di ogni scrittura e verifica dopo.
- **Verifica degli aggiornamenti** del programma a ogni avvio.

## Uso

1. Scarica [`IU0VYT.Flash.zip`](https://github.com/IU0VYT/IU0VYT/releases/latest/download/IU0VYT.Flash.zip) (ultima versione) ed estrai l'archivio.
2. Avvia `IU0VYT Flash.exe` (Windows 10/11; usa il componente WebView2 di Windows).
3. Dalla scheda **📻 Modelli** scegli la tua radio, collega il cavo e usa **Leggi dalla radio**.

La guida completa è nel file `LEGGIMI.txt`.

## Compilazione

Serve Go 1.22 o successivo. Nella cartella `sorgenti`:

```
go build -mod=vendor -trimpath -ldflags "-H=windowsgui -s -w" -o "IU0VYT Flash.exe" .
```

Le librerie sono incluse nella cartella `vendor`.

## Avvertenze

- I protocolli delle radio sono stati ricostruiti per l'interoperabilità e verificati con radio simulate: **conserva sempre le copie di sicurezza**. Il programma è fornito senza alcuna garanzia.
- Programma indipendente, **non affiliato né approvato** da Baofeng o Bajeton. I marchi citati appartengono ai rispettivi titolari e servono solo a indicare le radio compatibili.
- Rispetta le norme sull'uso delle frequenze del tuo Paese.

## Ringraziamenti

- [CHIRP](https://chirpmyradio.com) e i suoi autori: gran parte dei driver delle radio deriva dal loro lavoro.
- [IK2ANE](https://www.ik2ane.it), che mette a disposizione di tutti, gratuitamente, l'elenco dei ponti radio italiani.

Autori e licenze dei componenti: `AUTORI_E_LICENZE.txt`.

## Licenza

Copyright © 2026 Daniele "IU0VYT" Bonifazi · [qrz.com/db/IU0VYT](https://www.qrz.com/db/IU0VYT)

Software libero distribuito con la **GNU General Public License versione 3 o successiva** (`GPL-3.0-or-later`): vedi `LICENSE.txt`.
