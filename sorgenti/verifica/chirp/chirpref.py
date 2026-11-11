#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
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
# le versioni recenti di CHIRP usano nei driver la funzione di traduzione «_»,
# installata dall'interfaccia grafica: qui i testi restano quelli originali
import builtins  # noqa: E402
if not hasattr(builtins, "_"):
    builtins._ = lambda s: s
logging.disable(logging.CRITICAL)

from chirp import chirp_common, directory, errors, memmap  # noqa: E402
import chirp.drivers  # noqa: E402

for m in pkgutil.iter_modules(chirp.drivers.__path__):
    try:
        importlib.import_module("chirp.drivers." + m.name)
    except Exception:
        pass

# 4.9: driver esterni (CHIRP_EXTRA, vedi chirpextra.py)
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import chirpextra  # noqa: E402
chirpextra.load_extra(directory)

MAGIC = chirp_common.CloneModeRadio.MAGIC


def find_class(name):
    found = [cls for cls in directory.DRV_TO_RADIO.values() if cls.__name__ == name]
    # 4.15: due classi con lo stesso nome (Kenwood TH-D74 e TH-D72: copia della
    # memoria in thd74.py / thd72.py, comandi diretti in kenwood_live.py): vale
    # quella che copia la memoria
    clone = [cls for cls in found if issubclass(cls, chirp_common.CloneModeRadio)]
    if clone:
        return clone[0]
    if found:
        return found[0]
    raise SystemExit("classe sconosciuta: " + name)


def load(cls, path):
    data = open(path, "rb").read()
    meta = {}
    if MAGIC in data:
        try:
            meta = json.loads(base64.b64decode(data[data.index(MAGIC) + len(MAGIC):]).decode())
        except Exception:
            meta = {}
        data = data[:data.index(MAGIC)]
    r = cls(None)
    # 4.14: metadati che cambiano la lettura della memoria (Yaesu FT-65: sottotipo)
    if "ft4_subtype" in meta:
        r._metadata["ft4_subtype"] = meta["ft4_subtype"]
    if r.NEEDS_COMPAT_SERIAL:
        r._mmap = memmap.MemoryMap(data)
    else:
        r._mmap = memmap.MemoryMapBytes(bytes(data))
    r.process_mmap()
    return r


def raw_mmap(r):
    m = r.get_mmap()
    if getattr(r, "_highbit_flip", False):
        # 4.15 Icom ID-51 Plus2: get_mmap() restituisce i dati come viaggiano sul
        # cavo (bit alto rovesciato); il file .img contiene la memoria in chiaro
        m = r._mmap
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
        # radio senza tono di ricezione separato (has_ctone falso): in CHIRP
        # il tono del TSQL è rtone
        t = mem.ctone if (rf is None or rf.has_ctone) else mem.rtone
        tx = rx = "%.1f" % t
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
    # 4.9: int(): retevis_ha1g.py lascia in mem.freq il campo della memoria
    return {"rx": int(mem.freq), "tx": int(tx), "txOff": txoff, "txTone": t, "rxTone": rxt,
            "power": power, "mode": mem.mode, "scan": mem.skip != "S", "skip": mem.skip, "name": mem.name}


def zoned(r):
    """4.18: radio di CHIRP con una «sotto-radio» per zona e i canali in una
    tabella di posizioni fisiche: in IU0VYT Flash la lista è quella delle
    posizioni fisiche. Restituisce "x140" (Kenwood TK-2140/3140, tk3140.py),
    "x80" (Kenwood TK-280/380, tk280.py), "x180" (4.21: Kenwood TK-x180,
    tk8180.py: 250 posizioni per ognuna delle prime 32 zone, la posizione è il
    numero del canale nella zona) oppure None."""
    if hasattr(r, "probe_layout") and hasattr(r, "expand_mmap") and hasattr(r, "shuffle_zone"):
        return "x180"
    if hasattr(r, "_expand_system") and hasattr(r, "_reduce_system"):
        # 4.23: Kenwood TK-481 / TK-981 (tk481.py): 250 posizioni per ognuno dei 32 sistemi
        return "x481"
    if hasattr(r, "check_index") and hasattr(r, "_get_zone") and hasattr(r, "next_slot"):
        return "x140"
    if hasattr(r, "_get_memory_mapping") and hasattr(r, "_compact_mappings") and r.get_features().has_sub_devices:
        return "x80"
    return None


def x80_groups(r):
    return dict((int(s.group), s) for s in r.get_sub_devices())


def x80_slot(r, pos):
    """Voce della tabella dei gruppi che usa la posizione fisica pos."""
    for k in range(250):
        m = r._memobj.group_mapping[k]
        if int(m.group) < 255 and int(m.number) < 255 and int(m.index) == pos:
            return m
    return None


def zoned_target(r, pos, delete=False):
    """Sotto-radio e numero del canale per la posizione fisica pos (da 0): quelli
    del canale presente; per una posizione vuota la zona del canale che la
    precede (la prima zona se non c'è) e il primo numero libero. Restituisce
    (None, None) per la cancellazione di una posizione già vuota."""
    if zoned(r) == "x481":
        z, num = pos // 250, pos % 250 + 1
        for sub in r.get_sub_devices():
            if sub._system == z + 1:
                if delete:
                    try:
                        sub._get_memory(num)
                    except IndexError:
                        return None, None
                return sub, num
        return None, None
    if zoned(r) == "x180":
        subs = r.get_sub_devices()
        z, num = pos // 250, pos % 250 + 1
        if z >= len(subs) or (delete and subs[z]._get_raw_memory(num) is None):
            return None, None
        return subs[z], num
    if zoned(r) == "x80":
        subs = x80_groups(r)
        cur = x80_slot(r, pos)
        if cur is not None:
            return subs[int(cur.group)], int(cur.number)
        if delete:
            return None, None
        group = None
        for k in range(pos - 1, -1, -1):
            m = x80_slot(r, k)
            if m is not None:
                group = int(m.group)
                break
        gi = None
        for i in range(250):
            n = int(r._memobj.groups[i].number)
            if n > 250:
                continue
            if group is None:
                group = n
            if n == group:
                gi = i
                break
        used = set(int(m.number) for m in r._memobj.group_mapping
                   if int(m.group) == group and int(m.number) < 255)
        num = min(n for n in range(1, 251) if n not in used)
        # come _get_memory_mapping(allocate=True) di CHIRP, ma nella posizione
        # fisica scelta dalla lista (CHIRP prende la prima libera)
        for i in range(250):
            m = r._memobj.group_mapping[i]
            if int(m.group) == 0xFF:
                m.group = group
                m.number = num
                m.group_index = gi
                m.index = pos
                break
        r._memobj.groups[gi].channels += 1
        r._compact_mappings()
        return subs[group], num
    mems = r._memobj.memories
    if int(mems[pos].memory) != 0xFF:
        return r._get_zone(int(mems[pos].zone)), int(mems[pos].memory)
    if delete:
        return None, None
    zone = int(r._memobj.zoneinfo[0].zone)
    for k in range(pos - 1, -1, -1):
        if int(mems[k].memory) != 0xFF:
            zone = int(mems[k].zone)
            break
    used = set(int(x.memory) for x in mems if int(x.memory) != 0xFF and int(x.zone) == zone)
    num = min(n for n in range(1, 251) if n not in used)
    sub = r._get_zone(zone)
    sub.next_slot = lambda number, p=pos: p
    return sub, num


def decode(r):
    if zoned(r) == "x481":
        out = [None] * (32 * 250)
        for sub in r.get_sub_devices():
            z = sub._system - 1
            system = getattr(r._memobj, "system%i" % z)
            for i in range(int(system.sys.channels)):
                n = int(system.channels[i].number)
                # l'immagine di prova della TK-981 ha nei sistemi 2 e 3 un record
                # tutto a FF (numero 255): CHIRP mostra solo i numeri da 1 a 250
                if 1 <= n <= 250:
                    out[z * 250 + n - 1] = mem_to_json(sub, sub.get_memory(n))
        return out
    if zoned(r) == "x180":
        out = [None] * (32 * 250)
        for z, sub in enumerate(r.get_sub_devices()[:32]):
            for i in range(int(sub.raw_zoneinfo.count)):
                n = int(sub.raw_memories[i].number)
                out[z * 250 + n - 1] = mem_to_json(sub, sub.get_memory(n))
        return out
    if zoned(r) == "x80":
        out = [None] * 250
        for group, sub in x80_groups(r).items():
            for n in range(1, 251):
                try:
                    mapping = r._get_memory_mapping(group, n)
                except IndexError:
                    continue
                out[int(mapping.index)] = mem_to_json(sub, sub.get_memory(n))
        return out
    if zoned(r):
        out = [None] * len(r._memobj.memories)
        for zone in r._get_zone_info():
            sub = r._get_zone(zone)
            for n in range(1, 251):
                _i, slot = sub.check_index(n)
                if slot is not None:
                    out[int(slot)] = mem_to_json(sub, sub.get_memory(n))
        return out
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
    # come l'interfaccia di CHIRP: un campo non modificabile (immutable) non
    # viene toccato se il valore non cambia (canali fissi FRS, PMR, GMRS)
    fixed = getattr(mem, "immutable", []) or []
    # 4.9: radio che ridefiniscono check_set_memory_immutable_policy (es. Retevis
    # RB17P, BTECH GMRS-V2: nessun campo bloccato): come un canale incollato o
    # importato nell'interfaccia di CHIRP, i campi si impostano liberamente e poi
    # decide la regola della radio
    pol = getattr(type(r), "check_set_memory_immutable_policy", None)
    if fixed and pol is not None and type(r).__module__.split(".")[-1] in ("gmrsv2", "retevis_rb17p"):
        existing = mem.dupe()
        mem.__dict__["immutable"] = []
        fixed = []
        r._chirpref_existing = existing

    def setf(name, val):
        if not (name in fixed and getattr(mem, name) == val):
            setattr(mem, name, val)
    setf("empty", False)
    setf("freq", c["rx"])
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
        setf("power", rf.valid_power_levels[c["power"]])
    setf("mode", c.get("mode") or "FM")
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
    # 4.15: icx90.py chiama la mappa «memobj»
    walk("", getattr(r, "_memobj", None) or r.memobj)
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
        top = r
        for e in edits:
            try:
                if zoned(top):
                    r, num = zoned_target(top, e["n"] - 1, e["chan"] is None)
                    if r is None:
                        continue
                    e = dict(e, n=num)
                if e["chan"] is None:
                    # cancellazione come nell'interfaccia di CHIRP
                    r.erase_memory(e["n"])
                    continue
                mem = json_to_mem(r, e["n"], e["chan"])
                if zoned(top) == "x80":
                    # la posizione appena assegnata è ancora vuota (FF): il numero
                    # è quello chiesto, come nell'interfaccia di CHIRP
                    mem.number = e["n"]
                ex0 = r.__dict__.pop("_chirpref_existing", None)
                if ex0 is not None:
                    r.check_set_memory_immutable_policy(ex0, mem)
                msgs = [m for m in r.validate_memory(mem) if isinstance(m, chirp_common.ValidationError)]
                if msgs:
                    errs.append([e["n"], [str(m) for m in msgs]])
                scr = None
                if type(r).__module__.endswith("retevis_rt87") and not mem.empty and mem.extra:
                    # 4.9: retevis_rt87.py scrive lo scrambler aumentato di 1 a ogni
                    # modifica (confronta un RadioSetting con una stringa): IU0VYT
                    # Flash lo conserva, e qui si corregge il difetto di CHIRP
                    scr = int(r._memobj.channels[e["n"] - 1].scrambler)
                r.set_memory(mem)
                if scr is not None:
                    r._memobj.channels[e["n"] - 1].scrambler = scr
                if type(r).__module__.endswith("anytone_ht") and not mem.empty and mem.duplex != "off":
                    # 4.9: anytone_ht.py non toglie mai «tx_off»: un canale messo
                    # in sola ricezione non torna più a trasmettere. IU0VYT Flash
                    # lo toglie, e qui si corregge il difetto di CHIRP
                    r._memobj.memory[e["n"]].tx_off = False
            except Exception as ex:
                errs.append([e["n"], ["%s: %s" % (type(ex).__name__, ex)]])
        open(sys.argv[5], "wb").write(raw_mmap(top))
        print(json.dumps({"errors": errs}))
    elif cmd == "clone":
        addr, op = sys.argv[3], sys.argv[4]
        pipe = SockPipe(addr, cls.BAUD_RATE)
        if op == "download":
            r = cls(pipe)
            r.status_fn = lambda s: None
            mod = sys.modules[cls.__module__]
            det = getattr(cls, "_DETECTED_BY", None)
            if det is not None and det is not cls and hasattr(det, "detect_from_serial") and (
                    mod.__name__.endswith("ga510") or mod.__name__.endswith("retevis_h777v4")
                    or mod.__name__.endswith("tdm11")):
                # 4.7 GA-510 V2, 4.8 Retevis H777 V4: classe riconosciuta da
                # detect_from_serial di un'altra classe (come l'interfaccia di
                # CHIRP), che entra in programmazione; poi la lettura
                # 4.7: GA-510 V2 (famiglia UV-17): riconosciuta da detect_from_serial
                # della GA-510, che entra in programmazione; poi la lettura
                rc = det.detect_from_serial(pipe)
                r = rc(pipe)
                r.status_fn = lambda s: None
            elif mod.__name__.endswith("ga510") and hasattr(mod, "start_program") and hasattr(r, "_magic"):
                # ga510.py: l'ingresso in programmazione è fatto da
                # detect_from_serial (scelta del modello in CHIRP), non da sync_in
                mod.start_program(r)
            elif len(cls.detected_models()) > 1 or mod.__name__.endswith("tdh8"):
                # classe che riconosce più modelli (es. Retevis H777): come
                # l'interfaccia di CHIRP, prima detect_from_serial (che entra in
                # programmazione), poi la lettura; tdh8.py: l'ingresso in
                # programmazione è sempre fatto da detect_from_serial
                rc = cls.detect_from_serial(pipe)
                r = rc(pipe)
                r.status_fn = lambda s: None
            r.sync_in()
            open(sys.argv[5], "wb").write(raw_mmap(r))
        else:
            r = load(cls, sys.argv[5])
            # 4.9: metadati del file (alcuni driver scrivono solo un'immagine
            # letta dalla radio: TYT UVL-15W)
            raw = open(sys.argv[5], "rb").read()
            if MAGIC in raw:
                import base64
                try:
                    r.metadata = json.loads(base64.b64decode(raw[raw.index(MAGIC) + len(MAGIC):]))
                except Exception:
                    pass
            r.set_pipe(pipe) if hasattr(r, "set_pipe") else setattr(r, "pipe", pipe)
            r.status_fn = lambda s: None
            if hasattr(r, "_encode_current_time") and hasattr(r._memobj, "progtime"):
                # 4.9 tdm11.py: CHIRP scrive data e ora in un punto sbagliato
                # della memoria; IU0VYT Flash non le scrive: per il confronto
                # si lasciano i byte come sono
                r._encode_current_time = lambda: [int(x) for x in r._memobj.progtime.time]
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
