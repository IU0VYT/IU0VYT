#!/usr/bin/env python3
# IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
# Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
# SPDX-License-Identifier: GPL-3.0-or-later
# Software libero: GNU General Public License versione 3 o successiva, SENZA
# ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
# in AUTORI_E_LICENZE.txt.

# Unisce le tabelle prodotte da «chirpsettings.py extract» (un file JSON per
# classe di CHIRP) nel file settings_chirp.json usato dal programma.
#   settings_combine.py <cartella con i file .json> <settings_chirp.json>
import json
import os
import re
import sys

# esclusi anche qui (tabelle estratte prima della regola in chirpsettings.py):
# gruppi di prova/taratura e modalità di fabbrica
SKIP_GROUP = re.compile(r"(?i)service|calibrat|factory|tuning|engineer|test mode")
SKIP_KEY = re.compile(r"^modes\.modetype$")

src, dst = sys.argv[1], sys.argv[2]
out = {}
for f in sorted(os.listdir(src)):
    if not f.endswith(".json"):
        continue
    d = json.load(open(os.path.join(src, f)))
    if not d.get("stable") or not d.get("settings"):
        continue
    sets = [x for x in d["settings"] if not SKIP_GROUP.search(x["group"]) and not SKIP_KEY.search(x["key"])]
    out[d["class"]] = {"class": d["class"], "size": d["size"], "settings": sets}
json.dump(out, open(dst, "w"), ensure_ascii=False, separators=(",", ":"), sort_keys=True)
print(len(out), "classi")
