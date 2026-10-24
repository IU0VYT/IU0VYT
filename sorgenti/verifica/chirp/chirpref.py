#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
# Software libero: GNU General Public License versione 3 o successiva, SENZA
# ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
# in AUTORI_E_LICENZE.txt.

# Programma di confronto con CHIRP (https://chirpmyradio.com), usato dai test
# automatici di IU0VYT Flash per verificare i driver delle radio Baofeng e
# Bajeton: fa leggere, modificare, scaricare e caricare le memorie delle radio
# al codice di CHIRP e restituisce il risultato nel formato dei canali di
# IU0VYT Flash.
#
# Serve una copia dei sorgenti di CHIRP (git clone https://github.com/kk7ds/chirp)
# indicata dalla variabile CHIRP_DIR.
#
#   chirpref.py layout  <classe> <immagine>                 posizione di ogni campo della mappa
#   chirpref.py decode  <classe> <immagine>                 canali (JSON)
#   chirpref.py encode  <classe> <immagine> <modifiche.json> <uscita>
#   chirpref.py clone   <classe> <host:porta> download <uscita>
#   chirpref.py clone   <classe> <host:porta> upload <immagine>
#   chirpref.py models                                       elenco delle classi Baofeng/Bajeton
import base64
import importlib
import json
import logging
import os
import pkgutil
import socket
import sys
import time

sys.path.insert(0, os.environ.get("CHIRP_DIR", "."))
logging.disable(logging.CRITICAL)

from chirp import chirp_common, directory, errors, memmap  # noqa: E402
import chirp.drivers  # noqa: E402

for m in pkgutil.iter_modules(chirp.drivers.__path__):
    try:
        importlib.import_module("chirp.drivers." + m.name)
    except Exception:
        pass

MAGIC = chirp_common.CloneModeRadio.MAGIC


def find_class(name):
    for cls in directory.DRV_TO_RADIO.values():
        if cls.__name__ == name:
            return cls
    raise SystemExit("classe sconosciuta: " + name)


def load(cls, path):
    data = open(path, "rb").read()
    if MAGIC in data:
        data = data[:data.index(MAGIC)]
    r = cls(None)
    if r.NEEDS_COMPAT_SERIAL:
        r._mmap = memmap.MemoryMap(data)
    else:
        r._mmap = memmap.MemoryMapBytes(bytes(data))
    r.process_mmap()
    return r


def raw_mmap(r):
    m = r.get_mmap()
    if hasattr(m, "get_byte_compatible"):
        m = m.get_byte_compatible()
    return m.get_packed()


def dcs(code, pol):
    return "D%03d%s" % (code, "I" if pol == "R" else "N")


def tones(mem, rf=None):
    """(tono TX, tono RX) nel formato di IU0VYT Flash."""
    tx = rx = ""
    # radio senza un codice DCS separato per la ricezione: CHIRP usa mem.dtcs
    rxd = mem.rx_dtcs if (rf is None or rf.has_rx_dtcs) else mem.dtcs
    pol = mem.dtcs_polarity or "NN"
    if mem.tmode == "Tone":
        tx = "%.1f" % mem.rtone
    elif mem.tmode == "TSQL":
        tx = rx = "%.1f" % mem.ctone
    elif mem.tmode == "DTCS":
        tx = dcs(mem.dtcs, pol[0])
        rx = dcs(mem.dtcs, pol[1])
    elif mem.tmode == "Cross":
        a, b = mem.cross_mode.split("->")
        if a == "Tone":
            tx = "%.1f" % mem.rtone
        elif a == "DTCS":
            tx = dcs(mem.dtcs, pol[0])
        if b == "Tone":
            rx = "%.1f" % mem.ctone
        elif b == "DTCS":
            rx = dcs(rxd, pol[1])
    return tx, rx


def mem_to_json(r, mem):
    rf = r.get_features()
    if mem.empty:
        return None
    tx = mem.freq
    txoff = False
    if mem.duplex == "+":
        tx = mem.freq + mem.offset
    elif mem.duplex == "-":
        tx = mem.freq - mem.offset
    elif mem.duplex == "split":
        tx = mem.offset
    elif mem.duplex == "off":
        txoff = True
    t, rxt = tones(mem, rf)
    levels = [str(p) for p in rf.valid_power_levels]
    power = levels.index(str(mem.power)) if mem.power is not None and str(mem.power) in levels else -1
    return {"rx": mem.freq, "tx": tx, "txOff": txoff, "txTone": t, "rxTone": rxt,
            "power": power, "mode": mem.mode, "scan": mem.skip != "S", "name": mem.name}


def decode(r):
    rf = r.get_features()
    lo, hi = rf.memory_bounds
    out = []
    for n in range(lo, hi + 1):
        out.append(mem_to_json(r, r.get_memory(n)))
    return out


def parse_tone(s):
    if not s:
        return None
    if s.startswith("D"):
        return ("DTCS", int(s[1:4]), "R" if s.endswith("I") else "N")
    return ("Tone", float(s), None)


def json_to_mem(r, n, c):
    rf = r.get_features()
    mem = r.get_memory(n)
    if c is None:
        mem.empty = True
        return mem
    mem.empty = False
    mem.freq = c["rx"]
    # come l'interfaccia di CHIRP: un campo non modificabile (immutable) non
    # viene toccato se il valore non cambia
    if c.get("txOff"):
        dup, off = "off", 0
    elif c["tx"] == c["rx"]:
        # come nell'interfaccia di CHIRP: lo scostamento memorizzato resta
        dup, off = "", None
    elif abs(c["tx"] - c["rx"]) > 70000000 or "split" in c.get("dup", ""):
        dup, off = "split", c["tx"]
    elif c["tx"] > c["rx"]:
        dup, off = "+", c["tx"] - c["rx"]
    else:
        dup, off = "-", c["rx"] - c["tx"]
    fixed = getattr(mem, "immutable", []) or []
    if not ("duplex" in fixed and mem.duplex == dup):
        mem.duplex = dup
    if off is not None and not ("offset" in fixed and (mem.offset == off or dup == "off")):
        mem.offset = off
    t, rx = parse_tone(c["txTone"]), parse_tone(c["rxTone"])
    pol = [t[2] if t and t[0] == "DTCS" else "N", rx[2] if rx and rx[0] == "DTCS" else "N"]
    mem.dtcs_polarity = "".join(pol)
    if t is None and rx is None:
        mem.tmode = ""
    elif t and t[0] == "Tone" and rx is None:
        mem.tmode, mem.rtone = "Tone", t[1]
    elif t and rx and t[0] == "Tone" and rx[0] == "Tone" and t[1] == rx[1]:
        mem.tmode, mem.ctone = "TSQL", t[1]
        mem.rtone = t[1]
    elif t and rx and t[0] == "DTCS" and rx[0] == "DTCS" and t[1] == rx[1]:
        mem.tmode, mem.dtcs = "DTCS", t[1]
        mem.rx_dtcs = t[1]
    else:
        mem.tmode = "Cross"
        a = t[0] if t else ""
        b = rx[0] if rx else ""
        mem.cross_mode = "%s->%s" % (a, b)
        if t and t[0] == "Tone":
            mem.rtone = t[1]
        if t and t[0] == "DTCS":
            mem.dtcs = t[1]
        if rx and rx[0] == "Tone":
            mem.ctone = rx[1]
        if rx and rx[0] == "DTCS":
            mem.rx_dtcs = rx[1]
            if not rf.has_rx_dtcs:
                mem.dtcs = rx[1]
    if c.get("power", -1) >= 0 and rf.valid_power_levels:
        mem.power = rf.valid_power_levels[c["power"]]
    mem.mode = c.get("mode") or "FM"
    mem.skip = "" if c.get("scan", True) else "S"
    mem.name = c.get("name", "")
    return mem


class SockPipe:
    """Porta seriale simulata: collegamento TCP con la radio simulata."""

    def __init__(self, addr, baud):
        host, port = addr.rsplit(":", 1)
        self.s = socket.create_connection((host, int(port)))
        self.s.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)
        self.timeout = 1.0
        self._baud = baud
        self.rts = True
        self.dtr = True
        self.port = addr
        self.buf = b""
        self._ctl("baud %d" % baud)

    def _ctl(self, msg):
        # i comandi per la radio simulata viaggiano su un secondo canale: qui li
        # si registra soltanto sullo standard error
        sys.stderr.write("ctl: %s\n" % msg)

    @property
    def baudrate(self):
        return self._baud

    @baudrate.setter
    def baudrate(self, v):
        self._baud = v
        self._ctl("baud %d" % v)

    def setBaudrate(self, v):
        self.baudrate = v

    def write(self, data):
        if isinstance(data, str):
            data = data.encode("latin-1")
        self.s.sendall(bytes(data))
        return len(data)

    def read(self, size=1):
        deadline = time.time() + (self.timeout if self.timeout is not None else 30)
        while len(self.buf) < size:
            left = deadline - time.time()
            if left <= 0:
                break
            self.s.settimeout(left)
            try:
                d = self.s.recv(65536)
            except socket.timeout:
                break
            if not d:
                break
            self.buf += d
        out, self.buf = self.buf[:size], self.buf[size:]
        return out

    def inWaiting(self):
        self.s.settimeout(0.05)
        try:
            d = self.s.recv(65536)
            if d:
                self.buf += d
        except (socket.timeout, BlockingIOError):
            pass
        return len(self.buf)

    in_waiting = property(inWaiting)

    def reset_input_buffer(self):
        self.inWaiting()
        self.buf = b""

    flushInput = reset_input_buffer

    def reset_output_buffer(self):
        pass

    flushOutput = reset_output_buffer

    def flush(self):
        pass

    def log(self, *a):
        pass

    def close(self):
        self.s.close()


def layout(r):
    out = {}

    def walk(path, el):
        from chirp import bitwise
        if isinstance(el, bitwise.structDataElement):
            for k in el._keys:
                walk(path + ("." if path else "") + k, el._generators[k])
        elif isinstance(el, bitwise.arrayDataElement):
            if len(el) and not isinstance(el[0], (bitwise.structDataElement, bitwise.arrayDataElement)) and \
                    not isinstance(el[0], bitwise.bitDataElement):
                out[path] = [el._offset, len(el) * el[0].size() // 8, 0, 0, type(el[0]).__name__]
                return
            for i, x in enumerate(el):
                walk("%s[%d]" % (path, i), x)
        else:
            nb = getattr(el, "_nbits", 0) if isinstance(el, bitwise.bitDataElement) else 0
            sh = getattr(el, "_shift", 0) if nb else 0
            out[path] = [el._offset, el.size() // 8 if not nb else 0, nb, sh, type(el).__name__]
    walk("", r._memobj)
    return out


def main():
    cmd = sys.argv[1]
    if cmd == "models":
        res = []
        for cls in directory.DRV_TO_RADIO.values():
            if "aofeng" in cls.VENDOR or "ajeton" in cls.VENDOR:
                res.append([cls.__name__, cls.VENDOR, cls.MODEL, cls.VARIANT])
        print(json.dumps(sorted(res)))
        return
    cls = find_class(sys.argv[2])
    if cmd == "layout":
        print(json.dumps(layout(load(cls, sys.argv[3]))))
    elif cmd == "decode":
        print(json.dumps(decode(load(cls, sys.argv[3]))))
    elif cmd == "encode":
        r = load(cls, sys.argv[3])
        edits = json.load(open(sys.argv[4]))
        errs = []
        for e in edits:
            try:
                if e["chan"] is None:
                    # cancellazione come nell'interfaccia di CHIRP
                    r.erase_memory(e["n"])
                    continue
                mem = json_to_mem(r, e["n"], e["chan"])
                msgs = [m for m in r.validate_memory(mem) if isinstance(m, chirp_common.ValidationError)]
                if msgs:
                    errs.append([e["n"], [str(m) for m in msgs]])
                r.set_memory(mem)
            except Exception as ex:
                errs.append([e["n"], ["%s: %s" % (type(ex).__name__, ex)]])
        open(sys.argv[5], "wb").write(raw_mmap(r))
        print(json.dumps({"errors": errs}))
    elif cmd == "clone":
        addr, op = sys.argv[3], sys.argv[4]
        pipe = SockPipe(addr, cls.BAUD_RATE)
        if op == "download":
            r = cls(pipe)
            r.status_fn = lambda s: None
            mod = sys.modules[cls.__module__]
            if mod.__name__.endswith("ga510") and hasattr(mod, "start_program"):
                # ga510.py: l'ingresso in programmazione è fatto da
                # detect_from_serial (scelta del modello in CHIRP), non da sync_in
                mod.start_program(r)
            r.sync_in()
            open(sys.argv[5], "wb").write(raw_mmap(r))
        else:
            r = load(cls, sys.argv[5])
            r.set_pipe(pipe) if hasattr(r, "set_pipe") else setattr(r, "pipe", pipe)
            r.status_fn = lambda s: None
            r.sync_out()
        pipe.close()
        print(json.dumps({"ok": True}))
    else:
        raise SystemExit("comando sconosciuto")


if __name__ == "__main__":
    try:
        main()
    except errors.RadioError as e:
        print(json.dumps({"radioError": str(e)}))
        sys.exit(3)
