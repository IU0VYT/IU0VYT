//go:build !windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Impostazioni del K67 Pro confrontate con il codice ORIGINALE del programma
// «5RH PRO» (verifica/k67set): K67SET_HARNESS=…/k67set_harness.dll K67_CPS=…/CPS.exe
// 1) lettura: valori uguali a quelli di ConvertSetting;
// 2) scrittura: combinazioni casuali scritte qui e da SettingConvert: file identico.

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// chiave -> proprietà di CPS_5RH.Data.DatGenSetInfo
var k67SetProp = map[string]string{
	"cha_mode": "ChAMode", "chb_mode": "ChBMode", "blight_time": "BlightTime", "blight_lv": "BlightLv", "cha_disp": "ChADispMode",
	"chb_disp": "ChBDispMode", "disp_dir": "DispDir", "pwon_face": "PownFace", "lang": "LangSel", "dual_mode": "DualMode",
	"main_band": "MainBand", "sql": "Sqlv", "tail": "TailFreq", "busy_lock": "BusyLock", "tot": "TOT", "pretot": "PreTot",
	"vox": "VoxSW", "vox_lv": "VoxLv", "vox_dly": "Voxdettime", "hz1750": "HzTo1750", "voice": "Voice", "beep": "Tone",
	"end_tone": "EndTone", "record": "Reord", "record_mode": "RecordMode", "key_lock": "KeyLock", "auto_lock": "AutoKey",
	"skey1": "Skey1", "lkey1": "Lkey1", "skey2": "Skey2", "lkey2": "Lkey2", "radioname": "Radioname", "pow_pwd": "PowPassword",
	"wr_pwd": "WrPassword", "save": "PoSave", "save_dly": "PoSaveDly", "apo": "APO", "gps": "GpsSW", "gps_mode": "GpsMode",
	"gps_share": "GpsShare", "gps_req": "GpsReq", "gps_zone": "GpsZone", "gps_id": "GpsID", "aprs": "AprsSw", "lone_work": "LoneWork",
	"lone_time": "LoneWorkTim", "lone_resp": "LoneWorkRsp", "fall_alarm": "DaoDi", "noaa": "NOAA", "noaa_ch": "NoaaCh",
	"weather": "TianQi", "bt": "BlueT", "bt_app": "BluetAPP",
}

// valore della proprietà per il valore dell'impostazione
func k67PropVal(d *setDef, v string) string {
	switch d.Type {
	case "bool":
		if v == "true" {
			return "1"
		}
		return "0"
	}
	return v // scelte: indice; numeri e testi così come sono
}

func TestSettingsK67CPS(t *testing.T) {
	h, exe := os.Getenv("K67SET_HARNESS"), os.Getenv("K67_CPS")
	if h == "" || exe == "" {
		t.Skip("K67SET_HARNESS o K67_CPS non impostati")
	}
	base, err := os.ReadFile("templates/k67_base.xlc")
	if err != nil {
		t.Fatal(err)
	}
	tb := settingsFor(kindXLC, base)
	if tb == nil {
		t.Fatal("tabella assente")
	}
	if len(k67SetProp) != len(tb.Settings) {
		t.Fatalf("%d impostazioni, %d proprietà", len(tb.Settings), len(k67SetProp))
	}
	dir := t.TempDir()
	run := func(args ...string) string {
		out, err := exec.Command("dotnet", append([]string{h, exe}, args...)...).Output()
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				t.Fatalf("programma del costruttore: %v\n%s", err, ee.Stderr)
			}
			t.Fatal(err)
		}
		return string(out)
	}
	props := func(file string) map[string]string {
		m := map[string]string{}
		for _, l := range strings.Split(run("read", file), "\n") {
			if k := strings.Index(l, "="); k > 0 {
				m[l[:k]] = strings.TrimRight(l[k+1:], "\r")
			}
		}
		return m
	}
	check := func(img []byte, file string) {
		pm := props(file)
		for i := range tb.Settings {
			d := &tb.Settings[i]
			v, err := readSetting(d, img)
			if err != nil {
				t.Errorf("%s: %v", d.Key, err)
				continue
			}
			got := k67PropVal(d, v)
			want := strings.TrimRight(pm[k67SetProp[d.Key]], " \x00")
			if got != want {
				t.Errorf("%s: qui %q, programma %q", d.Key, got, want)
			}
		}
	}
	// 1) lettura del file di base
	bp := filepath.Join(dir, "base.xlc")
	_ = os.WriteFile(bp, base, 0o644)
	check(base, bp)
	// 2) combinazioni casuali
	rnd := rand.New(rand.NewSource(67))
	total := 0
	for r := 0; r < 12; r++ {
		mine := append([]byte{}, base...)
		var lines []string
		for _, k := range rnd.Perm(len(tb.Settings)) {
			d := &tb.Settings[k]
			var v string
			switch d.Type {
			case "bool":
				v = strconv.FormatBool(rnd.Intn(2) == 1)
			case "list":
				v = d.Options[rnd.Intn(len(d.Options))]
			case "int":
				v = strconv.Itoa(int(d.Min) + rnd.Intn(int(d.Max-d.Min)+1))
			case "str":
				cs := []rune(d.Charset)
				n := rnd.Intn(len(d.Enc.Offs) + 1)
				var s []rune
				for i := 0; i < n; i++ {
					s = append(s, cs[rnd.Intn(len(cs))])
				}
				v = strings.TrimSpace(string(s))
				if d.Key == "radioname" && len(v)%2 == 1 {
					v += "X" // il programma legge i nomi a coppie di byte
				}
			}
			if err := writeSetting(d, mine, v); err != nil {
				t.Fatalf("%s = %q: %v", d.Key, v, err)
			}
			lines = append(lines, k67SetProp[d.Key]+"="+k67PropVal(d, v))
		}
		mp := filepath.Join(dir, fmt.Sprintf("mod%d.txt", r))
		op := filepath.Join(dir, fmt.Sprintf("out%d.xlc", r))
		_ = os.WriteFile(mp, []byte(strings.Join(lines, "\n")), 0o644)
		run("write", bp, mp, op)
		theirs, _ := os.ReadFile(op)
		// il programma riscrive anche le altre voci (BTrxdly letto /500, ecc.):
		// si confrontano i byte delle impostazioni gestite qui
		for i := range tb.Settings {
			d := &tb.Settings[i]
			a, _ := readSetting(d, mine)
			b, err := readSetting(d, theirs)
			if err != nil || a != b {
				t.Errorf("giro %d, %s: qui %q, programma %q (%v)", r, d.Key, a, b, err)
			}
		}
		if diff := k67SetDiff(mine, theirs, tb); diff != "" {
			t.Fatalf("giro %d: byte diversi: %s", r, diff)
		}
		opf := filepath.Join(dir, fmt.Sprintf("mine%d.xlc", r))
		_ = os.WriteFile(opf, mine, 0o644)
		check(mine, opf)
		total += len(lines)
	}
	t.Logf("%d impostazioni lette come il programma del costruttore; %d modifiche in 12 combinazioni: byte identici", len(tb.Settings), total)
}

// k67SetDiff: differenze nei bit usati dalle impostazioni gestite
func k67SetDiff(a, b []byte, tb *setTable) string {
	var out []string
	for i := range tb.Settings {
		e := tb.Settings[i].Enc
		switch e.Kind {
		case "table", "num":
			offs, masks, _ := e.fields()
			for j, off := range offs {
				for k, m := range masks[j] {
					if a[off+k]&m != b[off+k]&m {
						out = append(out, fmt.Sprintf("%s@0x%X", tb.Settings[i].Key, off+k))
					}
				}
			}
		case "str":
			if !bytes.Equal(pick(a, e.Offs), pick(b, e.Offs)) {
				out = append(out, tb.Settings[i].Key)
			}
		}
	}
	return strings.Join(out, ", ")
}

func pick(b []byte, offs []int) []byte {
	out := make([]byte, len(offs))
	for i, o := range offs {
		out[i] = b[o]
	}
	return out
}
