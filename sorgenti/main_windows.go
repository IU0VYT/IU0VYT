//go:build windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Versione Windows: finestra del programma con WebView2 (già presente in
// Windows 11), finestre native per apri/salva, avviso alla chiusura se ci
// sono modifiche non salvate. Se WebView2 non è disponibile, l'interfaccia
// si apre nel browser predefinito.

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
)

func init() { runtime.LockOSThread() }

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	pMsgBox        = user32.NewProc("MessageBoxW")
	pDpiCtx        = user32.NewProc("SetProcessDpiAwarenessContext")
	pDPIAware      = user32.NewProc("SetProcessDPIAware")
	pSetWndLong    = user32.NewProc("SetWindowLongPtrW")
	pCallWndProc   = user32.NewProc("CallWindowProcW")
	pFindWindow    = user32.NewProc("FindWindowW")
	pSetForeground = user32.NewProc("SetForegroundWindow")
	pShowWindow    = user32.NewProc("ShowWindow")
	pSetWindowText = user32.NewProc("SetWindowTextW")
	pOpen          = comdlg32.NewProc("GetOpenFileNameW")
	pSave          = comdlg32.NewProc("GetSaveFileNameW")
	pCreateMutex   = kernel32.NewProc("CreateMutexW")
	pShellExec     = shell32.NewProc("ShellExecuteW")
)

const (
	mbOK         = 0x0
	mbYesNo      = 0x4
	mbIconError  = 0x10
	mbIconWarn   = 0x30
	mbIconInfo   = 0x40
	mbDefButton2 = 0x100
	idYes        = 6

	ofnOverwrite = 0x2
	ofnHideRO    = 0x4
	ofnNoChDir   = 0x8
	ofnPathMust  = 0x800
	ofnFileMust  = 0x1000
	ofnExplorer  = 0x80000

	wmClose    = 0x0010
	swMaximize = 3
	swRestore  = 9
)

const winTitle = "IU0VYT Flash"

var fileDialogNative = true

type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

// filtro con separatori NUL: "Descrizione|*.ext|..."
func winFilter(s string) *uint16 {
	r := utf16.Encode([]rune(strings.ReplaceAll(s, "|", "\x00") + "\x00\x00"))
	return &r[0]
}

var mainHwnd uintptr

func msgBox(text string, flags uintptr) int {
	r, _, _ := pMsgBox.Call(mainHwnd, uintptr(unsafe.Pointer(u16(text))), uintptr(unsafe.Pointer(u16(winTitle))), flags)
	return int(r)
}

// nativeDialog va eseguita sul thread della finestra.
func nativeDialog(save bool, title, filt, initial, defExt string) string {
	buf := make([]uint16, 32768)
	dir := ""
	if initial != "" {
		if strings.HasSuffix(initial, `\`) || strings.HasSuffix(initial, "/") {
			dir = initial
		} else {
			copy(buf, utf16.Encode([]rune(initial)))
			dir = filepath.Dir(initial)
		}
	}
	ofn := openFileName{
		hwndOwner:   mainHwnd,
		lpstrFilter: winFilter(filt),
		lpstrFile:   &buf[0],
		nMaxFile:    uint32(len(buf)),
		lpstrTitle:  u16(title),
		flags:       ofnExplorer | ofnNoChDir | ofnPathMust | ofnHideRO,
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	if defExt != "" {
		ofn.lpstrDefExt = u16(defExt)
	}
	if dir != "" {
		ofn.lpstrInitialDir = u16(dir)
	}
	p := pOpen
	if save {
		ofn.flags |= ofnOverwrite
		p = pSave
	} else {
		ofn.flags |= ofnFileMust
	}
	r, _, _ := p.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func setDPIAware() {
	// PER_MONITOR_AWARE_V2 = -4
	ctx := -4
	if pDpiCtx.Find() == nil {
		if r, _, _ := pDpiCtx.Call(uintptr(ctx)); r != 0 {
			return
		}
	}
	if pDPIAware.Find() == nil {
		_, _, _ = pDPIAware.Call()
	}
}

func singleInstance() bool {
	_, _, e := pCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(u16("Local\\GestioneCanaliRadio_K67_Bajeton"))))
	if e == syscall.Errno(183) { // ERROR_ALREADY_EXISTS
		if h, _, _ := pFindWindow.Call(uintptr(unsafe.Pointer(u16("webview"))), uintptr(unsafe.Pointer(u16(winTitle)))); h != 0 {
			_, _, _ = pShowWindow.Call(h, swRestore)
			_, _, _ = pSetForeground.Call(h)
		} else {
			msgBox("Il programma è già aperto.", mbOK|mbIconInfo)
		}
		return false
	}
	return true
}

func openBrowser(url string) {
	_, _, _ = pShellExec.Call(0, uintptr(unsafe.Pointer(u16("open"))), uintptr(unsafe.Pointer(u16(url))), 0, 0, 1)
}

func init() {
	// collegamenti nel browser predefinito; file mostrati in Esplora risorse
	openExternal = func(target string, selectFile bool) bool {
		if selectFile {
			args := "/select,\"" + target + "\""
			r, _, _ := pShellExec.Call(0, uintptr(unsafe.Pointer(u16("open"))), uintptr(unsafe.Pointer(u16("explorer.exe"))),
				uintptr(unsafe.Pointer(u16(args))), 0, 1)
			return r > 32
		}
		r, _, _ := pShellExec.Call(0, uintptr(unsafe.Pointer(u16("open"))), uintptr(unsafe.Pointer(u16(target))), 0, 0, 1)
		return r > 32
	}
}

var (
	origWndProc uintptr
	theApp      *App
)

const dirtyQuestion = "Ci sono modifiche ai canali non ancora salvate.\n\nVuoi chiudere il programma senza salvarle?"

func wndProc(hwnd, msg, wp, lp uintptr) uintptr {
	if msg == wmClose && theApp != nil && theApp.IsDirty() {
		if msgBox(dirtyQuestion, mbYesNo|mbIconWarn|mbDefButton2) != idYes {
			return 0
		}
	}
	r, _, _ := pCallWndProc.Call(origWndProc, hwnd, msg, wp, lp)
	return r
}

func main() {
	setDPIAware()
	if !singleInstance() {
		return
	}
	dir := appDataDir()
	st, err := openStore(dir)
	if err != nil {
		msgBox("Impossibile creare la cartella dei dati:\n"+dir+"\n\n"+err.Error(), mbOK|mbIconError)
		return
	}
	app := newApp(st)
	if t := os.Getenv("GESTIONE_CANALI_TOKEN"); t != "" { // solo per i test automatici
		app.token = t
	}
	theApp = app
	ln, url, err := app.listen()
	if err != nil {
		msgBox("Impossibile avviare il programma:\n"+err.Error(), mbOK|mbIconError)
		return
	}
	srv := &http.Server{Handler: app.handler()}
	go func() { _ = srv.Serve(ln) }()

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		DataPath:  filepath.Join(dir, "WebView2"),
		WindowOptions: webview2.WindowOptions{
			Title:  winTitle,
			Width:  1280,
			Height: 800,
			IconId: 1,
			Center: true,
		},
	})
	if w == nil {
		// Senza WebView2: interfaccia nel browser predefinito.
		// La finestra vuota creata prima dell'errore viene nascosta (e resa irriconoscibile
		// per il controllo "programma già aperto").
		if h, _, _ := pFindWindow.Call(uintptr(unsafe.Pointer(u16("webview"))), uintptr(unsafe.Pointer(u16(winTitle)))); h != 0 {
			_, _, _ = pShowWindow.Call(h, 0)
			_, _, _ = pSetWindowText.Call(h, uintptr(unsafe.Pointer(u16(""))))
		}
		fileDialog = func(save bool, title, filt, initial, defExt string) (string, error) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			return nativeDialog(save, title, filt, initial, defExt), nil
		}
		openBrowser(url)
		for {
			msgBox("Il programma è aperto nel browser (il componente WebView2 di Windows non è disponibile).\n\n"+
				"Premi OK quando hai finito, per chiudere il programma.", mbOK|mbIconInfo)
			if !app.IsDirty() || msgBox(dirtyQuestion, mbYesNo|mbIconWarn|mbDefButton2) == idYes {
				break
			}
		}
		_ = srv.Close()
		return
	}
	mainHwnd = uintptr(w.Window())
	fileDialog = func(save bool, title, filt, initial, defExt string) (string, error) {
		ch := make(chan string, 1)
		w.Dispatch(func() { ch <- nativeDialog(save, title, filt, initial, defExt) })
		return <-ch, nil
	}
	quitApp = func() { w.Dispatch(func() { w.Terminate() }) }
	gwlpWndProc := -4
	origWndProc, _, _ = pSetWndLong.Call(mainHwnd, uintptr(gwlpWndProc), syscall.NewCallback(wndProc))
	w.SetSize(900, 600, webview2.HintMin)
	_, _, _ = pShowWindow.Call(mainHwnd, swMaximize)
	w.Navigate(url)
	w.Run()
	_ = srv.Close()
	os.Exit(0)
}
