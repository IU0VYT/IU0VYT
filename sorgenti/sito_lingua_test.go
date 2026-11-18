// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"regexp"
	"testing"
)

// TestSitoLinguaRobot: lo script della pagina italiana (docs/index.html) porta i
// visitatori alla loro lingua, ma i motori di ricerca e i loro strumenti di prova
// devono restare sulla pagina italiana. Fino al 18/11/2026 lo strumento di prova di
// Google (Google-InspectionTool, usato da Rich Results Test e Controllo URL), che nel
// nome non ha «bot», veniva portato alla pagina inglese. Senza la cartella docs
// (pacchetto del programma) il test si salta.
func TestSitoLinguaRobot(t *testing.T) {
	b, err := os.ReadFile("../docs/index.html")
	if err != nil {
		t.Skip("docs/index.html assente: ", err)
	}
	m := regexp.MustCompile(`if \(/([^/]+)/i\.test\(navigator\.userAgent`).FindSubmatch(b)
	if m == nil {
		t.Fatal("elenco dei robot dello script della lingua non trovato in docs/index.html")
	}
	robot, err := regexp.Compile("(?i)" + string(m[1]))
	if err != nil {
		t.Fatalf("elenco dei robot non valido: %v", err)
	}
	restano := map[string]string{
		"Googlebot":             "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Googlebot smartphone":  "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		"Google-InspectionTool": "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 (compatible; Google-InspectionTool/1.0;)",
		"Bing":                  "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm) Chrome/120.0.0.0 Safari/537.36",
		"Yandex":                "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
		"Baidu":                 "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
		"DuckDuckGo":            "DuckDuckBot/1.1; (+http://duckduckgo.com/duckduckbot.html)",
		"Lighthouse":            "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Chrome-Lighthouse",
		"Facebook":              "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
	}
	for nome, ua := range restano {
		if !robot.MatchString(ua) {
			t.Errorf("%s verrebbe portato a un'altra lingua: aggiungerlo all'elenco dei robot", nome)
		}
	}
	browser := map[string]string{
		"Chrome Windows": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"Edge":           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
		"Firefox":        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
		"Safari iPhone":  "Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1",
		"Chrome Android": "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Mobile Safari/537.36",
	}
	for nome, ua := range browser {
		if robot.MatchString(ua) {
			t.Errorf("il browser %s verrebbe trattato come un robot e non passerebbe alla sua lingua", nome)
		}
	}
}
