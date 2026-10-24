//go:build windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type winPort struct {
	h    windows.Handle
	name string
}

func init() {
	openPortFunc = openWinPort
	listPortsFunc = listWinPorts
}

func openWinPort(name string, baud int) (Port, error) {
	path := `\\.\` + name
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		if err == windows.ERROR_ACCESS_DENIED {
			return nil, fmt.Errorf("la porta %s è già in uso (chiudi il programma originale o altri programmi che la usano)", name)
		}
		if err == windows.ERROR_FILE_NOT_FOUND {
			return nil, fmt.Errorf("la porta %s non esiste: controlla che il cavo sia collegato", name)
		}
		return nil, fmt.Errorf("impossibile aprire %s: %v", name, err)
	}
	wp := &winPort{h: h, name: name}
	_ = windows.SetupComm(h, 8192, 8192)
	if err := wp.SetBaud(baud); err != nil {
		windows.CloseHandle(h)
		return nil, err
	}
	// lettura: ritorna subito se ci sono byte, altrimenti aspetta al massimo 50 ms
	to := windows.CommTimeouts{ReadIntervalTimeout: 0xFFFFFFFF, ReadTotalTimeoutMultiplier: 0xFFFFFFFF,
		ReadTotalTimeoutConstant: 50, WriteTotalTimeoutMultiplier: 0, WriteTotalTimeoutConstant: 10000}
	if err := windows.SetCommTimeouts(h, &to); err != nil {
		windows.CloseHandle(h)
		return nil, fmt.Errorf("impostazione della porta %s non riuscita: %v", name, err)
	}
	_ = windows.PurgeComm(h, windows.PURGE_RXCLEAR|windows.PURGE_TXCLEAR)
	return wp, nil
}

func (p *winPort) SetBaud(baud int) error {
	var dcb windows.DCB
	dcb.DCBlength = uint32(unsafe.Sizeof(dcb))
	if err := windows.GetCommState(p.h, &dcb); err != nil {
		return fmt.Errorf("lettura impostazioni della porta %s non riuscita: %v", p.name, err)
	}
	dcb.BaudRate = uint32(baud)
	dcb.ByteSize = 8
	dcb.Parity = windows.NOPARITY
	dcb.StopBits = windows.ONESTOPBIT
	// binario, DTR e RTS attivi, nessun controllo di flusso (come SerialPort .NET)
	dcb.Flags = 0x0001 | windows.DTR_CONTROL_ENABLE | windows.RTS_CONTROL_ENABLE
	if err := windows.SetCommState(p.h, &dcb); err != nil {
		return fmt.Errorf("impostazione della porta %s a %d baud non riuscita: %v", p.name, baud, err)
	}
	return nil
}

func (p *winPort) Write(b []byte) error {
	for len(b) > 0 {
		var n uint32
		if err := windows.WriteFile(p.h, b, &n, nil); err != nil {
			return fmt.Errorf("scrittura sulla porta %s non riuscita: %v", p.name, err)
		}
		if n == 0 {
			return fmt.Errorf("scrittura sulla porta %s interrotta", p.name)
		}
		b = b[n:]
	}
	return nil
}

func (p *winPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	got := 0
	for got < len(b) {
		if err := ctx.Err(); err != nil {
			return errCancelled
		}
		var n uint32
		if err := windows.ReadFile(p.h, b[got:], &n, nil); err != nil {
			return fmt.Errorf("lettura dalla porta %s non riuscita: %v", p.name, err)
		}
		got += int(n)
		if got < len(b) && time.Now().After(deadline) {
			return errTimeout
		}
	}
	return nil
}

func (p *winPort) Purge() error {
	return windows.PurgeComm(p.h, windows.PURGE_RXCLEAR)
}

// Close: come SerialPort.Close di .NET, prima attende che i byte già scritti
// siano trasmessi (importante per l'ultimo comando prima di chiudere e per i
// cambi di velocità dell'aggiornamento del firmware), poi chiude.
func (p *winPort) Close() error {
	_ = windows.FlushFileBuffers(p.h)
	_ = windows.PurgeComm(p.h, windows.PURGE_RXCLEAR|windows.PURGE_TXCLEAR)
	return windows.CloseHandle(p.h)
}

// listWinPorts: porte COM presenti (registro di sistema) con il nome del
// dispositivo, quando Windows lo fornisce.
func listWinPorts() []PortInfo {
	names := map[string]string{}
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DEVICEMAP\SERIALCOMM`, registry.QUERY_VALUE); err == nil {
		vals, _ := k.ReadValueNames(-1)
		for _, v := range vals {
			if s, _, err := k.GetStringValue(v); err == nil && strings.HasPrefix(strings.ToUpper(s), "COM") {
				names[strings.ToUpper(s)] = ""
			}
		}
		k.Close()
	}
	for port, desc := range friendlyPortNames() {
		if _, ok := names[port]; ok {
			names[port] = desc
		}
	}
	var out []PortInfo
	for n, d := range names {
		out = append(out, PortInfo{Name: n, Desc: d})
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.Atoi(strings.TrimPrefix(out[i].Name, "COM"))
		b, _ := strconv.Atoi(strings.TrimPrefix(out[j].Name, "COM"))
		return a < b
	})
	return out
}

var (
	modSetupapi                       = windows.NewLazySystemDLL("setupapi.dll")
	procSetupDiGetClassDevsW          = modSetupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo         = modSetupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceRegistryPropW = modSetupapi.NewProc("SetupDiGetDeviceRegistryPropertyW")
	procSetupDiDestroyDeviceInfoList  = modSetupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

type spDevinfoData struct {
	cbSize    uint32
	classGUID windows.GUID
	devInst   uint32
	reserved  uintptr
}

// friendlyPortNames: "COM6" -> "USB-SERIAL CH340" (dalla classe "Porte (COM e LPT)").
func friendlyPortNames() map[string]string {
	out := map[string]string{}
	guid := windows.GUID{Data1: 0x4d36e978, Data2: 0xe325, Data3: 0x11ce, Data4: [8]byte{0xbf, 0xc1, 0x08, 0x00, 0x2b, 0xe1, 0x03, 0x18}}
	const digcfPresent = 0x2
	h, _, _ := procSetupDiGetClassDevsW.Call(uintptr(unsafe.Pointer(&guid)), 0, 0, digcfPresent)
	if h == 0 || h == ^uintptr(0) {
		return out
	}
	defer procSetupDiDestroyDeviceInfoList.Call(h)
	for i := 0; i < 256; i++ {
		var d spDevinfoData
		d.cbSize = uint32(unsafe.Sizeof(d))
		r, _, _ := procSetupDiEnumDeviceInfo.Call(h, uintptr(i), uintptr(unsafe.Pointer(&d)))
		if r == 0 {
			break
		}
		buf := make([]uint16, 512)
		const spdrpFriendlyName = 0xC
		r, _, _ = procSetupDiGetDeviceRegistryPropW.Call(h, uintptr(unsafe.Pointer(&d)), spdrpFriendlyName, 0,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)*2), 0)
		if r == 0 {
			continue
		}
		fn := syscall.UTF16ToString(buf)
		// "USB-SERIAL CH340 (COM6)"
		if i := strings.LastIndex(fn, "(COM"); i >= 0 && strings.HasSuffix(fn, ")") {
			port := strings.ToUpper(fn[i+1 : len(fn)-1])
			out[port] = strings.TrimSpace(fn[:i])
		}
	}
	return out
}
