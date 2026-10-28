#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
# Software libero: GNU General Public License versione 3 o successiva, SENZA
# ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
# in AUTORI_E_LICENZE.txt.

# Impostazioni delle radio ricavate dal codice di CHIRP (https://chirpmyradio.com,
# GNU GPL): per ogni impostazione della scheda «Settings» di CHIRP si fa scrivere
# ogni valore possibile al codice di CHIRP (get_settings / set_settings) e si
# guarda quali bit della memoria cambiano. Il risultato (posizione dei bit e
# valore per ogni scelta, oppure la codifica dei numeri e dei testi) è la tabella
# usata da IU0VYT Flash (settings_chirp.json); la stessa procedura, con
# combinazioni casuali di modifiche, verifica la tabella.
#
#   chirpsettings.py extract <classe> <immagine>              tabella (JSON)
#   chirpsettings.py apply   <classe> <immagine> <mod.json> <uscita>
#                     (mod.json: {"chiave": valore}; l'immagine di partenza è
#                      quella già normalizzata da set_settings, vedi «base»)
#   chirpsettings.py base    <classe> <immagine> <uscita>     immagine normalizzata
#   chirpsettings.py values  <classe> <immagine>              valori attuali (JSON)
import importlib
import json
import logging
import os
import pkgutil
import random
import re
import sys

sys.path.insert(0, os.environ.get("CHIRP_DIR", "."))
logging.disable(logging.CRITICAL)
import warnings  # noqa: E402
warnings.simplefilter("ignore")

from chirp import chirp_common, directory, memmap, settings  # noqa: E402
import chirp.drivers  # noqa: E402

for m in pkgutil.iter_modules(chirp.drivers.__path__):
    try:
        importlib.import_module("chirp.drivers." + m.name)
    except Exception:
        pass

MAGIC = chirp_common.CloneModeRadio.MAGIC
# gruppi di servizio (taratura): non si toccano
SKIP_GROUP = re.compile(r"(?i)service|calibrat|factory|tuning|engineer|test mode")


def find_class(name):
    for cls in directory.DRV_TO_RADIO.values():
        if cls.__name__ == name:
            return cls
    raise SystemExit("classe sconosciuta: " + name)


def read_data(path):
    data = open(path, "rb").read()
    if MAGIC in data:
        data = data[:data.index(MAGIC)]
    return data


def load(cls, data):
    r = cls(None)
    if r.NEEDS_COMPAT_SERIAL:
        r._mmap = memmap.MemoryMap(data)
    else:
        r._mmap = memmap.MemoryMapBytes(bytes(data))
    r.process_mmap()
    return r


def raw(r):
    m = r.get_mmap()
    if hasattr(m, "get_byte_compatible"):
        m = m.get_byte_compatible()
    return bytes(m.get_packed())


def flat(tree, path=()):
    out = []
    for e in tree:
        if isinstance(e, settings.RadioSetting):
            out.append((path, e))
        elif isinstance(e, settings.RadioSettingGroup):
            out.extend(flat(e, path + (e.get_shortname(),)))
    return out


def keyed(r):
    """impostazioni con una chiave unica: nome CHIRP (+ #n se ripetuto)"""
    tree = r.get_settings()
    seen = {}
    items = []
    for path, e in flat(tree):
        k = e.get_name()
        if k in seen:
            seen[k] += 1
            k = "%s#%d" % (k, seen[k])
        else:
            seen[k] = 0
        items.append((k, path, e))
    return tree, items


def kind(v):
    if isinstance(v, settings.RadioSettingValueBoolean):
        return "bool"
    if isinstance(v, settings.RadioSettingValueList):
        return "list"
    if isinstance(v, settings.RadioSettingValueInteger):
        return "int"
    if isinstance(v, settings.RadioSettingValueFloat):
        return "float"
    if isinstance(v, settings.RadioSettingValueString):
        return "str"
    return None


def current(v):
    k = kind(v)
    if k == "bool":
        return bool(v.get_value())
    if k in ("list",):
        return str(v.get_value())
    if k == "int":
        return int(v.get_value())
    if k == "float":
        return float(v.get_value())
    if k == "str":
        return str(v.get_value())
    return None


def apply(cls, base, changes):
    r = load(cls, base)
    tree, items = keyed(r)
    idx = {k: e for k, _, e in items}
    for k, val in changes.items():
        idx[k].value.set_value(val)
    r.set_settings(tree)
    return raw(r)


class FastApply:
    """apply() senza rileggere ogni volta la memoria con process_mmap (lento): si
    usa sempre lo stesso oggetto radio e si rimette la memoria di partenza al suo
    posto. Si usa solo se, per questa classe, dà gli stessi risultati di apply()
    su un campione di modifiche; la verifica finale (settings_test.go) usa
    comunque apply() con una lettura nuova."""

    def __init__(self, cls, b0, items, rnd):
        self.cls, self.b0, self.ok = cls, b0, False
        try:
            self.r = load(cls, b0)
        except Exception:  # noqa: BLE001
            return
        samples = []
        for key, path_, e in items:
            v = e.value
            if SKIP_GROUP.search("/".join(path_)) or len(e) != 1 or not v.get_mutable():
                continue
            k = kind(v)
            if k == "bool":
                samples.append({key: not bool(v.get_value())})
            elif k == "list" and len(v.get_options()) > 1:
                samples.append({key: rnd.choice([o for o in v.get_options() if o != str(v.get_value())])})
            elif k == "int" and v.get_max() > v.get_min():
                samples.append({key: v.get_min() if int(v.get_value()) != v.get_min() else v.get_max()})
        rnd.shuffle(samples)
        samples = samples[:12]
        if len(samples) >= 2:
            samples.append(dict(kv for s_ in samples[:6] for kv in s_.items()))
        samples.append({})
        self.ok = True
        for ch in samples:
            try:
                want = apply(cls, b0, ch)
            except Exception:  # noqa: BLE001
                continue
            try:
                got = self.apply(ch)
            except Exception:  # noqa: BLE001
                got = None
            if got != want:
                self.ok = False
                return

    def apply(self, changes):
        if not self.ok and hasattr(self, "_checked"):
            return apply(self.cls, self.b0, changes)
        r = self.r
        m = r._mmap
        m._data[:] = list(self.b0)
        tree, items = keyed(r)
        idx = {k: e for k, _, e in items}
        for k, val in changes.items():
            idx[k].value.set_value(val)
        r.set_settings(tree)
        return raw(r)


FAST = None


def apply_fast(cls, base, changes):
    if FAST is not None and FAST.ok and FAST.cls is cls and FAST.b0 == base:
        return FAST.apply(changes)
    return apply(cls, base, changes)


def fix_none(cls, data):
    """immagine di prova di un altro modello della stessa famiglia: le voci con un
    valore che CHIRP non riconosce (None) bloccano set_settings; si mettono al
    primo valore ammesso (serve solo come punto di partenza: le codifiche si
    ricavano cambiando i valori)"""
    r = load(cls, data)
    tree, items = keyed(r)
    n = 0
    for _, _, e in items:
        for v in e:
            if not isinstance(v, settings.RadioSettingValue) or not v.get_mutable():
                continue
            try:
                cur = v.get_value()
            except Exception:  # noqa: BLE001
                cur = None
            if cur is not None:
                continue
            try:
                if isinstance(v, settings.RadioSettingValueList):
                    v.set_value(v.get_options()[0])
                elif isinstance(v, (settings.RadioSettingValueInteger, settings.RadioSettingValueFloat)):
                    v.set_value(v.get_min())
                elif isinstance(v, settings.RadioSettingValueBoolean):
                    v.set_value(False)
                elif isinstance(v, settings.RadioSettingValueString):
                    v.set_value("A" * max(v.minlength, 0))
                else:
                    continue
                n += 1
            except Exception:  # noqa: BLE001
                pass
    if n == 0:
        return data
    r.set_settings(tree)
    return raw(r)


def normalize(cls, data):
    """immagine dopo un set_settings senza modifiche (stabile)"""
    try:
        b0 = apply(cls, data, {})
    except Exception:  # noqa: BLE001
        data = fix_none(cls, data)
        b0 = apply(cls, data, {})
    b1 = apply(cls, b0, {})
    return b0, b0 == b1


def diff_mask(base, outs):
    mask = {}
    for o in outs:
        if len(o) != len(base):
            return None
        for i, (a, b) in enumerate(zip(base, o)):
            if a != b:
                mask[i] = mask.get(i, 0) | (a ^ b)
    return mask


def fields_of(mask):
    """intervalli contigui: [(inizio, [maschere])]"""
    out = []
    for off in sorted(mask):
        if out and out[-1][0] + len(out[-1][1]) == off:
            out[-1][1].append(mask[off])
        else:
            out.append((off, [mask[off]]))
    return out


def masked(data, fields):
    b = []
    for off, ms in fields:
        for k, m in enumerate(ms):
            b.append(data[off + k] & m)
    return bytes(b)


# ------------------------------------------------------------------ numeri

def num_raw(data, fields, enc):
    off, ms = fields[0]
    bs = [data[off + k] & ms[k] for k in range(len(ms))]
    if enc["fmt"] == "digits":
        # una cifra per byte (la prima è la più significativa)
        order = bs if enc["endian"] == "be" else list(reversed(bs))
        v = 0
        for x in order:
            if x > 9:
                return None
            v = v * 10 + x
        return v
    if enc["endian"] == "be":
        v = 0
        for x in bs:
            v = v << 8 | x
        mv = 0
        for x in ms:
            mv = mv << 8 | x
    else:
        v = 0
        for x in reversed(bs):
            v = v << 8 | x
        mv = 0
        for x in reversed(ms):
            mv = mv << 8 | x
    sh = (mv & -mv).bit_length() - 1
    v >>= sh
    if enc["fmt"] == "bcd":
        d = 0
        mul = 1
        while v:
            n = v & 15
            if n > 9:
                return None
            d += n * mul
            mul *= 10
            v >>= 4
        v = d
    return v


def find_num(base, fields, samples):
    """samples: [(intero, bytes)] -> codifica raw = v*num/den + add"""
    from fractions import Fraction
    if len(fields) != 1 or len(fields[0][1]) > 10:
        return None
    for fmt in ("bin", "bcd", "digits"):
        for endian in ("le", "be"):
            enc = {"endian": endian, "fmt": fmt}
            pts = [(v, num_raw(d, fields, enc)) for v, d in samples]
            if any(p[1] is None for p in pts):
                continue
            (v0, r0), (v1, r1) = pts[0], pts[-1]
            if v1 == v0:
                continue
            mul = Fraction(r1 - r0, v1 - v0)
            add = r0 - mul * v0
            if mul == 0 or add.denominator != 1:
                continue
            if all(r == v * mul + add for v, r in pts):
                enc.update(num=mul.numerator, den=mul.denominator, add=int(add))
                return enc
    return None


# ------------------------------------------------------------------ estrazione

def int_values(v):
    lo, hi, st = v.get_min(), v.get_max(), v.get_step() or 1
    return list(range(lo, hi + 1, st))


def extract(cls, path):
    data = read_data(path)
    b0, stable = normalize(cls, data)
    r = load(cls, b0)
    _, items = keyed(r)
    out = []
    skipped = []
    rnd = random.Random(1)
    global FAST
    FAST = FastApply(cls, b0, items, random.Random(2))
    FAST._checked = True
    for key, path_, e in items:
        v = e.value
        group = "/".join(path_)
        ent = {"key": key, "group": group, "label": e.get_shortname()}
        if SKIP_GROUP.search(group):
            skipped.append({"key": key, "why": "gruppo di servizio"})
            continue
        if len(e) != 1:
            skipped.append({"key": key, "why": "impostazione multipla"})
            continue
        k = kind(v)
        if k is None:
            skipped.append({"key": key, "why": "tipo non gestito"})
            continue
        try:
            w = e.get_warning(None)
        except Exception:
            w = None
        if w:
            ent["warn"] = w
        ent["type"] = k
        cur = current(v)
        if not v.get_mutable():
            ent["ro"] = True
            ent["type"] = "info"
            ent["value"] = str(cur).strip()
            out.append(ent)
            continue
        try:
            if k == "bool":
                vals = [False, True]
            elif k == "list":
                vals = list(v.get_options())
                ent["options"] = vals
            elif k == "int":
                vals = int_values(v)
                ent.update(min=v.get_min(), max=v.get_max(), step=v.get_step() or 1)
            elif k == "float":
                res = v._res
                dec = max(0, -int(round(__import__("math").log10(res)))) if res < 1 else 0
                lo, hi = v.get_min(), v.get_max()
                n = int(round((hi - lo) / res))
                ent.update(min=lo, max=hi, res=res, dec=dec)
                vals = None
            else:
                vals = None
            if k == "int" and len(vals) > 24:
                numenc = num_try(cls, b0, key, v, vals, rnd)
                if numenc is not None:
                    ent["enc"] = numenc
                    out.append(ent)
                    continue
            if k in ("bool", "list") or (k == "int" and len(vals) <= 512):
                outs = {}
                for x in vals:
                    try:
                        outs[x] = apply_fast(cls, b0, {key: x})
                    except Exception:
                        outs[x] = None
                ok = {x: o for x, o in outs.items() if o is not None}
                mask = diff_mask(b0, list(ok.values()))
                if mask is None:
                    skipped.append({"key": key, "why": "dimensione cambiata"})
                    continue
                if not mask:
                    skipped.append({"key": key, "why": "nessun effetto sulla memoria"})
                    continue
                fl = fields_of(mask)
                if sum(len(m) for _, m in fl) > 32:
                    skipped.append({"key": key, "why": "troppi byte"})
                    continue
                table = []
                for x in vals:
                    table.append(masked(ok[x], fl).hex() if x in ok else None)
                curm = masked(b0, fl).hex()
                if k == "bool":
                    curv = cur
                    want = table[1 if cur else 0]
                elif k == "list":
                    want = table[vals.index(cur)] if cur in vals else None
                else:
                    want = table[vals.index(cur)] if cur in vals else None
                if want != curm:
                    skipped.append({"key": key, "why": "valore attuale non riconoscibile"})
                    continue
                ent["enc"] = {"kind": "table", "fields": [[off, bytes(ms).hex()] for off, ms in fl], "table": table}
                out.append(ent)
                continue
            if k == "str":
                st = str_enc(cls, b0, key, v, rnd)
                if st is not None:
                    ent.update(maxlen=v.maxlength, minlen=v.minlength, charset=st.pop("charset"), autopad=bool(v.autopad))
                    ent["enc"] = st
                    out.append(ent)
                    continue
            if k in ("int", "float") or (k == "str" and re.fullmatch(r"\s*\d+(\.\d+)?\s*", str(cur) or "")):
                # numeri grandi (anche testi numerici: frequenze del VFO)
                if k == "int":
                    allv = int_values(v)
                    sv = sorted(set([allv[0], allv[-1]] + rnd.sample(allv, min(40, len(allv)))))
                    conv = lambda x: x  # noqa: E731
                    scale = 1
                elif k == "float":
                    steps = list(range(0, n + 1))
                    picks = sorted(set([0, n] + rnd.sample(steps, min(40, len(steps)))))
                    sv = [round(lo + p * res, dec) for p in picks]
                    scale = 10 ** dec
                    conv = lambda x: x  # noqa: E731
                else:
                    s = str(cur).strip()
                    dec = len(s.split(".")[1]) if "." in s else 0
                    width = len(str(cur))
                    ent.update(dec=dec, width=width)
                    base = float(s)
                    scale = 10 ** dec
                    sv = []
                    for _ in range(40):
                        x = round(base + rnd.uniform(-5, 5), dec) if dec else int(base) + rnd.randint(-50, 50)
                        if x > 0:
                            sv.append(x)
                    sv = sorted(set(sv + [base]))
                    conv = lambda x: (("%." + str(dec) + "f") % x if dec else str(int(x))).ljust(width) if v.autopad else (("%." + str(dec) + "f") % x if dec else str(int(x)))  # noqa: E731,E501
                outs = []
                for x in sv:
                    try:
                        outs.append((x, apply_fast(cls, b0, {key: conv(x)})))
                    except Exception:
                        pass
                if len(outs) < 3:
                    skipped.append({"key": key, "why": "valori non accettati"})
                    continue
                mask = diff_mask(b0, [o for _, o in outs])
                if not mask:
                    skipped.append({"key": key, "why": "nessun effetto sulla memoria"})
                    continue
                fl = fields_of(mask)
                samples = [(int(round(x * scale)), o) for x, o in outs]
                enc = find_num(b0, fl, samples)
                if enc is None:
                    skipped.append({"key": key, "why": "codifica numerica non riconosciuta"})
                    continue
                cv = int(round(float(str(cur).strip()) * scale))
                if num_raw(b0, fl, enc) * enc["den"] != cv * enc["num"] + enc["add"] * enc["den"]:
                    skipped.append({"key": key, "why": "valore attuale non riconoscibile"})
                    continue
                enc.update(kind="num", fields=[[off, bytes(ms).hex()] for off, ms in fl], scale=scale)
                if k == "str":
                    ent["type"] = "numstr"
                    ent["autopad"] = bool(v.autopad)
                ent["enc"] = enc
                out.append(ent)
                continue
            skipped.append({"key": key, "why": "tipo non gestito"})
        except Exception as ex:  # noqa: BLE001
            skipped.append({"key": key, "why": "errore: %s" % ex})
    return {"class": cls.__name__, "stable": stable, "size": len(b0), "settings": out, "skipped": skipped}


def num_try(cls, b0, key, v, allv, rnd):
    """intero con molti valori: prima la codifica numerica su un campione
    (molto più veloce della tabella completa); None se non riconosciuta"""
    sv = sorted(set([allv[0], allv[-1]] + rnd.sample(allv, min(30, len(allv)))))
    outs = []
    for x in sv:
        try:
            outs.append((x, apply_fast(cls, b0, {key: x})))
        except Exception:
            pass
    if len(outs) < 3:
        return None
    mask = diff_mask(b0, [o for _, o in outs])
    if not mask:
        return None
    fl = fields_of(mask)
    enc = find_num(b0, fl, [(x, o) for x, o in outs])
    if enc is None:
        return None
    cv = int(v.get_value())
    if num_raw(b0, fl, enc) * enc["den"] != cv * enc["num"] + enc["add"] * enc["den"]:
        return None
    enc.update(kind="num", fields=[[off, bytes(ms).hex()] for off, ms in fl], scale=1)
    return enc


def str_enc(cls, b0, key, v, rnd):
    """codifica di un testo: un byte per carattere con una tabella dei caratteri.
    Si usano solo i caratteri ASCII stampabili del charset di CHIRP (le tabelle con
    migliaia di ideogrammi non servono ai nomi scritti da qui); se due caratteri
    finiscono nello stesso byte (es. minuscole convertite in maiuscole) resta il
    primo del charset."""
    cs = v._charset
    if isinstance(cs, str):
        cs = list(cs)
    cs = [c for c in cs if len(c) == 1 and 32 <= ord(c) < 127]
    ml = v.maxlength
    if not cs or ml == 0:
        return None
    fill = "A" if "A" in cs else cs[0]
    alts = [c for c in cs if c != fill]
    if not alts:
        return None
    try:
        full = apply_fast(cls, b0, {key: fill * ml})
    except Exception:
        return None
    # posizione di ogni carattere
    offs = []
    for p in range(ml):
        s = fill * p + alts[0] + fill * (ml - p - 1)
        try:
            d = diff_mask(full, [apply_fast(cls, b0, {key: s})])
        except Exception:
            return None
        if not d or len(d) != 1:
            return None
        offs.append(list(d)[0])
    if len(set(offs)) != ml:
        return None
    # byte di ogni carattere: ml caratteri per volta
    cmap = {fill: full[offs[0]]}
    if any(full[o] != cmap[fill] for o in offs):
        return None
    pend = list(alts)
    while pend:
        chunk, pend = pend[:ml], pend[ml:]
        s = "".join(chunk) + fill * (ml - len(chunk))
        try:
            o = apply_fast(cls, b0, {key: s})
        except Exception:
            # un carattere non accettato: si provano uno per uno
            for c in chunk:
                try:
                    o1 = apply_fast(cls, b0, {key: c + fill * (ml - 1)})
                except Exception:
                    continue
                if c not in cmap:
                    cmap[c] = o1[offs[0]]
            continue
        for i, c in enumerate(chunk):
            cmap[c] = o[offs[i]]
    seen = {}
    for c in list(cmap):
        if cmap[c] in seen:
            del cmap[c]
        else:
            seen[cmap[c]] = c
    if fill not in cmap:
        return None
    enc = {"kind": "str", "offs": offs, "map": cmap, "charset": "".join(c for c in cs if c in cmap)}
    # riempimento (testo vuoto o corto) e verifica con testi casuali
    pads = set()
    for _ in range(16):
        n = rnd.randint(max(v.minlength, 0), ml)
        s = "".join(rnd.choice(list(cmap)) for _ in range(n))
        try:
            o = apply_fast(cls, b0, {key: s})
        except Exception:
            continue
        padc = s.ljust(ml) if v.autopad else s
        exp = bytearray(b0)
        for p, off in enumerate(offs):
            if p < len(padc):
                if padc[p] not in cmap:
                    return None
                exp[off] = cmap[padc[p]]
            else:
                exp[off] = o[off]
                pads.add(o[off])
        if bytes(exp) != o:
            return None
    if len(pads) > 1:
        return None
    if pads:
        enc["pad"] = pads.pop()
    return enc


def values(cls, data):
    r = load(cls, data)
    _, items = keyed(r)
    out = {}
    for k, _, e in items:
        if len(e) == 1 and kind(e.value):
            c = current(e.value)
            out[k] = c
    return out


def main():
    cmd = sys.argv[1]
    cls = find_class(sys.argv[2])
    data = read_data(sys.argv[3])
    if cmd == "extract":
        json.dump(extract(cls, sys.argv[3]), sys.stdout, ensure_ascii=False)
    elif cmd == "base":
        b0, _ = normalize(cls, data)
        open(sys.argv[4], "wb").write(b0)
    elif cmd == "apply":
        ch = json.load(open(sys.argv[4]))
        open(sys.argv[5], "wb").write(apply(cls, data, ch))
    elif cmd == "values":
        json.dump(values(cls, data), sys.stdout, ensure_ascii=False)
    else:
        raise SystemExit("comando sconosciuto")


if __name__ == "__main__":
    main()
