// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Copie di prova e pre-release (4.47).
//
// Una copia è «di prova» quando il suo numero è più alto di quello dell'ultima
// versione stabile pubblicata: l'interfaccia mostra allora una barra di avviso.
// L'eseguibile della pre-release e della stabile è lo stesso: quando la versione
// diventa stabile i numeri coincidono e la barra sparisce da sola.
//
// Solo in una copia di prova il programma legge anche l'elenco delle release
// (indirizzo fisso, non modificabile) e propone la pre-release più recente, se
// ha un numero più alto. Chi usa la stabile non legge mai l'elenco e non vede le
// pre-release. Il pacchetto si scarica sempre e solo da GitHub, da un indirizzo
// costruito dal programma con il numero della pre-release.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const (
	defaultReleasesURL = "https://api.github.com/repos/IU0VYT/IU0VYT/releases"
	defaultRcBase      = "https://github.com/IU0VYT/IU0VYT/releases/download/"
	rcAssetName        = "IU0VYT.Flash-rc.zip"
)

// indirizzi fissi (cambiano solo nei test)
var (
	releasesSource = defaultReleasesURL
	rcDownloadBase = defaultRcBase
)

// tag delle pre-release: V.x.y.z-rc (solo cifre e punti nel numero)
var reRcTag = regexp.MustCompile(`^V\.(\d+(?:\.\d+)*)-rc$`)

type ghAsset struct {
	Name string `json:"name"`
}

type ghRelease struct {
	Tag        string    `json:"tag_name"`
	Prerelease bool      `json:"prerelease"`
	Draft      bool      `json:"draft"`
	Assets     []ghAsset `json:"assets"`
}

// fetchReleases legge l'elenco delle release dall'indirizzo fisso
func fetchReleases(u string) ([]ghRelease, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "IU0VYT-Flash/"+appVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := updHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("collegamento al server degli aggiornamenti non riuscito: controlla la connessione a Internet (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, updStatusErr(resp)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("lettura interrotta: %v", err)
	}
	var list []ghRelease
	// errore mai mostrato: senza elenco valido la pre-release semplicemente non si propone
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// latestPreRelease: fra le pre-release pubblicate (non bozze) con l'allegato
// IU0VYT.Flash-rc.zip, quella con il numero più alto
func latestPreRelease(list []ghRelease) (version, tag string) {
	for _, r := range list {
		if !r.Prerelease || r.Draft {
			continue
		}
		m := reRcTag.FindStringSubmatch(r.Tag)
		if m == nil {
			continue
		}
		has := false
		for _, a := range r.Assets {
			if a.Name == rcAssetName {
				has = true
				break
			}
		}
		if has && (version == "" || cmpVersion(m[1], version) > 0) {
			version, tag = m[1], r.Tag
		}
	}
	return
}

// chooseUpdate: la versione da proporre a chi usa «local». Se sono più recenti
// sia la stabile sia la pre-release vince il numero più alto; a parità la
// stabile. "" se non c'è nulla di più recente.
func chooseUpdate(local, stable, pre string) (version string, isPre bool) {
	sNew := stable != "" && cmpVersion(stable, local) > 0
	pNew := pre != "" && cmpVersion(pre, local) > 0
	switch {
	case sNew && pNew:
		if cmpVersion(pre, stable) > 0 {
			return pre, true
		}
		return stable, false
	case pNew:
		return pre, true
	case sNew:
		return stable, false
	}
	return "", false
}

type preInfo struct {
	version, built, url string
	size, read          int64
}

// checkPreRelease: la pre-release più recente, controllata leggendo la versione
// scritta nel suo pacchetto (senza scaricarlo). ok = false se non c'è o se il
// pacchetto non corrisponde al numero del tag.
func checkPreRelease() (p preInfo, ok bool) {
	list, err := fetchReleases(releasesSource)
	if err != nil {
		return p, false
	}
	v, tag := latestPreRelease(list)
	if v == "" {
		return p, false
	}
	u := rcDownloadBase + tag + "/" + rcAssetName
	pv, built, size, read, err := remoteVersion(u)
	if err != nil || cmpVersion(pv, v) != 0 {
		return p, false
	}
	return preInfo{version: v, built: built, url: u, size: size, read: read}, true
}
