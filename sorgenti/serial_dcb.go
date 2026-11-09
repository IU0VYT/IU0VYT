// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Bit del campo Flags della struttura DCB di Windows (porta seriale).
const (
	dcbFBinary     = 0x0001 // modalità binaria (obbligatoria)
	dcbFParity     = 0x0002 // controllo degli errori di parità
	dcbDTREnable   = 0x0010 // DTR attivo
	dcbRTSEnable   = 0x1000 // RTS attivo
	dcbFlowMask    = 0x236C // controlli di flusso: CTS, DSR, sensibilità DSR, XON/XOFF, handshake DTR e RTS
	dcbFlagsSimple = dcbFBinary | dcbDTREnable | dcbRTSEnable
)

// dcbFlags (4.9): i bit della DCB per una porta. Binario, DTR e RTS attivi,
// nessun controllo di flusso (come SerialPort di .NET); con la parità pari
// (Lanchonlh HG-UV98) anche il controllo degli errori di parità.
// Fino alla 4.8 fParity veniva impostato e subito dopo cancellato.
// rtsOff (4.9): RTS spento (Retevis C2, WANTS_RTS = False in CHIRP).
func dcbFlags(even, rtsOff bool) uint32 { return dcbFlagsDTR(even, rtsOff, false) }

// dcbFlagsDTR (4.13): come dcbFlags, con DTR spento se dtrOff (collegamento dati
// del firmware IJV V6)
func dcbFlagsDTR(even, rtsOff, dtrOff bool) uint32 {
	f := uint32(dcbFlagsSimple)
	if dtrOff {
		f &^= dcbDTREnable
	}
	if even {
		f |= dcbFParity
	}
	if rtsOff {
		f &^= dcbRTSEnable
	}
	return f
}
