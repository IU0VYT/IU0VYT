# IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
# Software libero: GNU General Public License versione 3 o successiva, SENZA
# ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
# in AUTORI_E_LICENZE.txt.

# 4.9: driver di CHIRP esterni (non compresi in CHIRP, come quello per il
# firmware F4HWN della Quansheng UV-K5), caricati come i «moduli» dell'interfaccia
# di CHIRP. Variabile CHIRP_EXTRA: elenco «file.py:suffisso» separato da «;».
# Le classi del modulo che hanno lo stesso nome di una classe di CHIRP prendono
# il suffisso (es. UVK5RadioEgzumer → UVK5RadioEgzumerF4HWN).
#
# I moduli esterni possono usare wxPython (finestre dell'interfaccia) e aprire
# pagine web: qui si sostituiscono con funzioni che non fanno nulla.
import importlib.util
import os
import sys
import types
import webbrowser


def _stub_wx():
    if "wx" in sys.modules:
        return
    wx = types.ModuleType("wx")
    wx.OK, wx.CANCEL, wx.CANCEL_DEFAULT, wx.ICON_WARNING = 4, 16, 0x80000000, 0x100
    wx.YES, wx.NO, wx.YES_NO = 2, 8, 10
    wx.MessageBox = lambda *a, **k: wx.CANCEL
    sys.modules["wx"] = wx


def load_extra(directory):
    spec = os.environ.get("CHIRP_EXTRA", "")
    if not spec:
        return
    _stub_wx()
    webbrowser.open = lambda *a, **k: False
    for k, item in enumerate(x for x in spec.split(";") if x):
        path, _, suffix = item.partition(":")
        known = {c.__name__ for c in directory.DRV_TO_RADIO.values()}
        before = set(directory.DRV_TO_RADIO.values())
        name = "chirp_extra_%d" % k
        sp = importlib.util.spec_from_file_location(name, path)
        mod = importlib.util.module_from_spec(sp)
        sys.modules[name] = mod
        sp.loader.exec_module(mod)
        for cls in set(directory.DRV_TO_RADIO.values()) - before:
            if cls.__name__ in known and suffix:
                cls.__name__ = cls.__name__ + suffix
