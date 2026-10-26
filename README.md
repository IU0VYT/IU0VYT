# IU0VYT Flash

**Versione 3.5** · Windows · interfaccia in italiano · un solo file eseguibile, nessuna installazione

Programma per radioamatori e utenti di ricetrasmettitori **Baofeng** e **Bajeton**: legge e scrive canali e frequenze via cavo di programmazione, gestisce un archivio dei ponti radio italiani e aggiorna il firmware delle radio supportate.

## Funzioni

- **Canali e frequenze** di 230 modelli Baofeng e Bajeton (comprese radio PMR digitali e DMR): lettura e scrittura via cavo, file `.img` compatibili con CHIRP, file originali del Baofeng K67 Pro (`.xlc`) e della Bajeton 8300 (`.dat`).
- **Galleria di 271 modelli**, divisi tra compatibili e non compatibili, per tipologia (radioamatoriali, DMR, GMRS/FRS, PMR446, per bambini) e per marca, con ricerca, filtri e collegamenti ai programmi e ai manuali ufficiali.
- **Archivio dei ponti radio** basato sull'elenco aggiornato di [IK2ANE](https://www.ik2ane.it), con filtri e inserimento nelle radio, con il controllo della compatibilità tra ponti e radio prima della scrittura.
- **Aggiornamento del firmware** di 28 modelli, con controllo del file e copia di sicurezza della radio.
- **Copia di sicurezza** automatica prima di ogni scrittura e verifica dopo.
- **Verifica degli aggiornamenti** del programma a ogni avvio.

## Uso

1. Scarica `IU0VYT.Flash.zip` dalla [pagina dei rilasci](https://github.com/IU0VYT/IU0VYT/releases) ed estrai l'archivio.
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
