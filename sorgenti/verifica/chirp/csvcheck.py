#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
# Software libero: GNU General Public License versione 3 o successiva, SENZA
# ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
# in AUTORI_E_LICENZE.txt.

# Confronto del CSV di CHIRP scritto e letto da IU0VYT Flash con il codice di CHIRP
# (driver generic_csv, https://chirpmyradio.com, GNU GPL):
#   csvcheck.py parse  <file.csv>                    canali letti da CHIRP (JSON)
#   csvcheck.py export <classe> <immagine> <uscita>  CSV scritto da CHIRP con i canali
#                                                    dell'immagine di una radio
# I canali sono descritti come in IU0VYT Flash: frequenze RX/TX in Hz, toni come
# testo («88.5», «D023N», «D023I»), larghezza, scansione, sola ricezione, DMR.
import json
import os
import sys

sys.path.insert(0, os.environ.get("CHIRP_DIR", "."))
import logging  # noqa: E402
logging.disable(logging.CRITICAL)

from chirp import chirp_common, directory  # noqa: E402
from chirp.drivers import generic_csv  # noqa: E402
import chirpsettings as cs  # noqa: E402


def tone(kind, ctcs, code, pol):
    if kind == "Tone":
        return "%.1f" % ctcs
    if kind == "DTCS":
        return "D%03d%s" % (code, "I" if pol == "R" else "N")
    return ""


def describe(m):
    rx = int(m.freq)
    tx, txoff = rx, False
    if m.duplex == "+":
        tx = rx + int(m.offset)
    elif m.duplex == "-":
        tx = rx - int(m.offset)
    elif m.duplex == "split":
        tx = int(m.offset)
    elif m.duplex == "off":
        txoff = True
    pol = (m.dtcs_polarity or "NN") + "NN"
    t, r = "", ""
    if m.tmode == "Tone":
        t = tone("Tone", m.rtone, 0, "N")
    elif m.tmode == "TSQL":
        t = r = tone("Tone", m.ctone, 0, "N")
    elif m.tmode == "DTCS":
        t = tone("DTCS", 0, m.dtcs, pol[0])
        r = tone("DTCS", 0, m.dtcs, pol[1])
    elif m.tmode == "Cross":
        a, b = (m.cross_mode.split("->") + [""])[:2]
        if a == "Tone":
            t = tone("Tone", m.rtone, 0, "N")
        elif a == "DTCS":
            t = tone("DTCS", 0, m.dtcs, pol[0])
        if b == "Tone":
            r = tone("Tone", m.ctone, 0, "N")
        elif b == "DTCS":
            r = tone("DTCS", 0, m.rx_dtcs, pol[1])
    return {"loc": m.number, "name": m.name.rstrip(), "rx": rx, "tx": tx, "txTone": t, "rxTone": r,
            "narrow": m.mode in ("NFM", "NAM", "DMR"), "dig": m.mode == "DMR", "scan": m.skip != "S", "txOff": txoff,
            "power": str(m.power) if m.power else ""}


def main():
    cmd = sys.argv[1]
    if cmd == "parse":
        r = generic_csv.CSVRadio(sys.argv[2])
        out = []
        lo, hi = r.get_features().memory_bounds
        for n in range(lo, hi + 1):
            m = r.get_memory(n)
            if not m.empty:
                out.append(describe(m))
        json.dump(out, sys.stdout)
    elif cmd == "export":
        cls = cs.find_class(sys.argv[2])
        radio = cs.load(cls, cs.read_data(sys.argv[3]))
        csvr = generic_csv.CSVRadio(None)
        lo, hi = radio.get_features().memory_bounds
        for n in range(lo, hi + 1):
            m = radio.get_memory(n)
            if m.empty:
                continue
            m.number = n - lo + 1
            csvr.set_memory(m)
        csvr.save(sys.argv[4])
    else:
        raise SystemExit("comando sconosciuto")


if __name__ == "__main__":
    main()
