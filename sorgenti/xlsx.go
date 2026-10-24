// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Lettura e scrittura minima dei file Excel .xlsx (Office Open XML).

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

func zipFile(zr *zip.Reader, name string) ([]byte, error) {
	name = strings.TrimPrefix(name, "/")
	for _, f := range zr.File {
		if strings.EqualFold(f.Name, name) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(io.LimitReader(rc, 200<<20))
		}
	}
	return nil, fmt.Errorf("%s mancante nel file .xlsx", name)
}

// colonna "AB12" -> 27 (0-based) e riga 11
func cellRef(r string) (int, int, bool) {
	c, i := 0, 0
	for i < len(r) && r[i] >= 'A' && r[i] <= 'Z' {
		c = c*26 + int(r[i]-'A'+1)
		i++
	}
	if i == 0 || i == len(r) {
		return 0, 0, false
	}
	row, err := strconv.Atoi(r[i:])
	if err != nil || row < 1 {
		return 0, 0, false
	}
	return c - 1, row - 1, true
}

type xText struct {
	T string   `xml:"t"`
	R []xTextR `xml:"r"`
}
type xTextR struct {
	T string `xml:"t"`
}

func (x xText) String() string {
	if len(x.R) == 0 {
		return x.T
	}
	var b strings.Builder
	b.WriteString(x.T)
	for _, r := range x.R {
		b.WriteString(r.T)
	}
	return b.String()
}

func readXLSX(b []byte) ([]*Sheet, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, errors.New("file .xlsx non valido")
	}
	var wb struct {
		Sheets []struct {
			Name string `xml:"name,attr"`
			RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
		} `xml:"sheets>sheet"`
	}
	wbx, err := zipFile(zr, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	if err := xml.Unmarshal(wbx, &wb); err != nil {
		return nil, fmt.Errorf("workbook.xml: %v", err)
	}
	var rels struct {
		R []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if rx, err := zipFile(zr, "xl/_rels/workbook.xml.rels"); err == nil {
		_ = xml.Unmarshal(rx, &rels)
	}
	var shared []string
	if sx, err := zipFile(zr, "xl/sharedStrings.xml"); err == nil {
		var sst struct {
			SI []xText `xml:"si"`
		}
		if err := xml.Unmarshal(sx, &sst); err != nil {
			return nil, fmt.Errorf("sharedStrings.xml: %v", err)
		}
		for _, s := range sst.SI {
			shared = append(shared, s.String())
		}
	}
	var out []*Sheet
	for i, s := range wb.Sheets {
		target := fmt.Sprintf("worksheets/sheet%d.xml", i+1)
		for _, r := range rels.R {
			if r.ID == s.RID {
				target = r.Target
			}
		}
		p := target
		if strings.HasPrefix(p, "/") {
			p = strings.TrimPrefix(p, "/")
		} else {
			p = path.Join("xl", p)
		}
		sx, err := zipFile(zr, p)
		if err != nil {
			continue
		}
		var ws struct {
			Rows []struct {
				R     int `xml:"r,attr"`
				Cells []struct {
					R  string `xml:"r,attr"`
					T  string `xml:"t,attr"`
					V  string `xml:"v"`
					IS xText  `xml:"is"`
				} `xml:"c"`
			} `xml:"sheetData>row"`
		}
		if err := xml.Unmarshal(sx, &ws); err != nil {
			return nil, fmt.Errorf("foglio %s: %v", s.Name, err)
		}
		sh := &Sheet{Name: s.Name}
		for ri, row := range ws.Rows {
			rn := row.R - 1
			if row.R == 0 {
				rn = ri
			}
			for ci, c := range row.Cells {
				col, r := ci, rn
				if c.R != "" {
					if cc, rr, ok := cellRef(c.R); ok {
						col, r = cc, rr
					}
				}
				var cell Cell
				switch c.T {
				case "s":
					k, _ := strconv.Atoi(c.V)
					if k >= 0 && k < len(shared) {
						cell = Cell{S: shared[k]}
					}
				case "inlineStr":
					cell = Cell{S: c.IS.String()}
				case "str", "e":
					cell = Cell{S: c.V}
				case "b":
					cell = Cell{S: map[bool]string{true: "VERO", false: "FALSO"}[c.V == "1"]}
				default:
					if v, err := strconv.ParseFloat(strings.TrimSpace(c.V), 64); err == nil {
						cell = numCell(v)
					} else {
						cell = Cell{S: c.V}
					}
				}
				cell.Fill, cell.Font = -1, -1
				sh.set(r, col, cell)
			}
		}
		out = append(out, sh)
	}
	if len(out) == 0 {
		return nil, errors.New("nessun foglio leggibile nel file .xlsx")
	}
	return out, nil
}

func xmlEsc(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			continue // caratteri di controllo non ammessi in XML
		}
		_ = xml.EscapeText(&b, []byte(string(r)))
	}
	return b.String()
}

func colName(c int) string {
	s := ""
	for c++; c > 0; c = (c - 1) / 26 {
		s = string(rune('A'+(c-1)%26)) + s
	}
	return s
}

func writeXLSX(sheets []WSheet) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, content string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, content)
		return err
	}
	const hdr = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"
	var ct, wbs, rels strings.Builder
	var sheetXML []string
	ct.WriteString(hdr + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
		`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	wbs.WriteString(hdr + `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	rels.WriteString(hdr + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i, sh := range sheets {
		n := i + 1
		name := sh.Name
		if name == "" {
			name = fmt.Sprintf("Foglio%d", n)
		}
		fmt.Fprintf(&ct, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, n)
		fmt.Fprintf(&wbs, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEsc(name), n, n)
		fmt.Fprintf(&rels, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, n, n)
		var s strings.Builder
		s.WriteString(hdr + `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
		s.WriteString(`<sheetViews><sheetView workbookViewId="0"`)
		if i == 0 {
			s.WriteString(` tabSelected="1"`)
		}
		s.WriteString(`><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
		if len(sh.Widths) > 0 {
			s.WriteString("<cols>")
			for c, wd := range sh.Widths {
				if wd > 0 {
					fmt.Fprintf(&s, `<col min="%d" max="%d" width="%d" customWidth="1"/>`, c+1, c+1, wd+2)
				}
			}
			s.WriteString("</cols>")
		}
		s.WriteString("<sheetData>")
		ncol := 0
		for r, row := range sh.Rows {
			ncol = max(ncol, len(row))
			fmt.Fprintf(&s, `<row r="%d">`, r+1)
			for c, cell := range row {
				ref := colName(c) + strconv.Itoa(r+1)
				st := ""
				if cell.Style != stNormal {
					st = fmt.Sprintf(` s="%d"`, cell.Style)
				}
				if cell.IsNum {
					fmt.Fprintf(&s, `<c r="%s"%s><v>%s</v></c>`, ref, st, strconv.FormatFloat(cell.N, 'f', -1, 64))
				} else if cell.S != "" {
					fmt.Fprintf(&s, `<c r="%s"%s t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, ref, st, xmlEsc(cell.S))
				} else if st != "" {
					fmt.Fprintf(&s, `<c r="%s"%s/>`, ref, st)
				}
			}
			s.WriteString("</row>")
		}
		s.WriteString("</sheetData>")
		if len(sh.Rows) > 1 && ncol > 0 {
			fmt.Fprintf(&s, `<autoFilter ref="A1:%s%d"/>`, colName(ncol-1), len(sh.Rows))
		}
		s.WriteString("</worksheet>")
		sheetXML = append(sheetXML, s.String())
	}
	ct.WriteString(`</Types>`)
	wbs.WriteString(`</sheets></workbook>`)
	rels.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, len(sheets)+1))
	rels.WriteString(`</Relationships>`)
	styles := hdr + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
		`<fonts count="2"><font><sz val="10"/><name val="Arial"/></font><font><b/><sz val="10"/><name val="Arial"/></font></fonts>` +
		`<fills count="5"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFC0C0C0"/><bgColor indexed="64"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFFFFF99"/><bgColor indexed="64"/></patternFill></fill>` +
		`<fill><patternFill patternType="solid"><fgColor rgb="FFFF99CC"/><bgColor indexed="64"/></patternFill></fill></fills>` +
		`<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>` +
		`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
		`<cellXfs count="4"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
		`<xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/>` +
		`<xf numFmtId="0" fontId="0" fillId="3" borderId="0" xfId="0" applyFill="1"/>` +
		`<xf numFmtId="0" fontId="0" fillId="4" borderId="0" xfId="0" applyFill="1"/></cellXfs>` +
		`<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`
	files := []struct{ n, c string }{
		{"[Content_Types].xml", ct.String()},
		{"_rels/.rels", hdr + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", wbs.String()},
		{"xl/_rels/workbook.xml.rels", rels.String()},
		{"xl/styles.xml", styles},
	}
	for i, x := range sheetXML {
		files = append(files, struct{ n, c string }{fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), x})
	}
	for _, f := range files {
		if err := add(f.n, f.c); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
