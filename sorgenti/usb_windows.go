//go:build windows

// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Ricerca dei dispositivi HID come hid-windows.c di dmrconfig (licenza BSD a 3
// clausole, Copyright (C) 2018 Serge Vakulenko KK6ABQ), riscritta in Go.

package main

// Radio collegate via USB come dispositivo HID (RD-5R, DM-1801) o in modalità DFU
// (DM-1701, vedi sotto). Le radio HID: Windows le
// gestisce con il proprio driver HID, senza installare nulla. Il dispositivo si
// apre con openPortFunc("HID:15A2:0073", 0): ogni Write invia un rapporto di 42
// byte, ReadFull riceve il rapporto di risposta.

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modHid                              = windows.NewLazySystemDLL("hid.dll")
	procHidDGetAttributes               = modHid.NewProc("HidD_GetAttributes")
	procSetupDiEnumDeviceInterfaces     = modSetupapi.NewProc("SetupDiEnumDeviceInterfaces")
	procSetupDiGetDeviceInterfaceDetail = modSetupapi.NewProc("SetupDiGetDeviceInterfaceDetailW")
)

func init() {
	serial := openPortFunc
	openPortFunc = func(name string, baud int) (Port, error) {
		if strings.HasPrefix(name, "HID:") {
			return openHIDPort(name)
		}
		if strings.HasPrefix(name, "DFU:") {
			return openDFUPort(name)
		}
		return serial(name, baud)
	}
}

func usbIDs(name string) (uint16, uint16, error) {
	f := strings.Split(name, ":")
	if len(f) != 3 {
		return 0, 0, fmt.Errorf("dispositivo USB %q non valido", name)
	}
	v, err1 := strconv.ParseUint(f[1], 16, 16)
	p, err2 := strconv.ParseUint(f[2], 16, 16)
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("dispositivo USB %q non valido", name)
	}
	return uint16(v), uint16(p), nil
}

type spDeviceInterfaceData struct {
	cbSize   uint32
	guid     windows.GUID
	flags    uint32
	reserved uintptr
}

// usbDevicePaths: percorsi dei dispositivi presenti con l'interfaccia guid
func usbDevicePaths(guid windows.GUID) []string {
	const digcfPresent, digcfDeviceInterface = 0x2, 0x10
	h, _, _ := procSetupDiGetClassDevsW.Call(uintptr(unsafe.Pointer(&guid)), 0, 0, digcfPresent|digcfDeviceInterface)
	if h == 0 || h == ^uintptr(0) {
		return nil
	}
	defer procSetupDiDestroyDeviceInfoList.Call(h)
	var out []string
	for i := 0; i < 256; i++ {
		var d spDeviceInterfaceData
		d.cbSize = uint32(unsafe.Sizeof(d))
		r, _, _ := procSetupDiEnumDeviceInterfaces.Call(h, 0, uintptr(unsafe.Pointer(&guid)), uintptr(i), uintptr(unsafe.Pointer(&d)))
		if r == 0 {
			break
		}
		var need uint32
		procSetupDiGetDeviceInterfaceDetail.Call(h, uintptr(unsafe.Pointer(&d)), 0, 0, uintptr(unsafe.Pointer(&need)), 0)
		if need < 8 || need > 4096 {
			continue
		}
		buf := make([]byte, need)
		// SP_DEVICE_INTERFACE_DETAIL_DATA_W: cbSize 8 (64 bit) o 6 (32 bit), poi il percorso
		cb := uint32(8)
		if unsafe.Sizeof(uintptr(0)) == 4 {
			cb = 6
		}
		*(*uint32)(unsafe.Pointer(&buf[0])) = cb
		r, _, _ = procSetupDiGetDeviceInterfaceDetail.Call(h, uintptr(unsafe.Pointer(&d)), uintptr(unsafe.Pointer(&buf[0])), uintptr(need), 0, 0)
		if r == 0 {
			continue
		}
		u := unsafe.Slice((*uint16)(unsafe.Pointer(&buf[4])), (len(buf)-4)/2)
		out = append(out, windows.UTF16ToString(u))
	}
	return out
}

type hidPort struct {
	h    windows.Handle
	mu   sync.Mutex
	rbuf []byte
}

func openHIDPort(name string) (Port, error) {
	vid, pid, err := usbIDs(name)
	if err != nil {
		return nil, err
	}
	guid := windows.GUID{Data1: 0x4d1e55b2, Data2: 0xf16f, Data3: 0x11cf, Data4: [8]byte{0x88, 0xcb, 0x00, 0x11, 0x11, 0x00, 0x00, 0x30}}
	for _, path := range usbDevicePaths(guid) {
		p, err := windows.UTF16PtrFromString(path)
		if err != nil {
			continue
		}
		h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
			nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
		if err != nil {
			continue
		}
		var attr struct {
			size                uint32
			vendorID, productID uint16
			version             uint16
		}
		attr.size = uint32(unsafe.Sizeof(attr))
		r, _, _ := procHidDGetAttributes.Call(uintptr(h), uintptr(unsafe.Pointer(&attr)))
		if r != 0 && attr.vendorID == vid && attr.productID == pid {
			return &hidPort{h: h}, nil
		}
		windows.CloseHandle(h)
	}
	return nil, fmt.Errorf("radio non trovata sulla porta USB (dispositivo %04X:%04X): collega il cavo USB e accendi la radio", vid, pid)
}

// overlapped: lettura e scrittura con tempo massimo
func (p *hidPort) io(write bool, b []byte, timeout time.Duration) (int, error) {
	ev, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(ev)
	ov := windows.Overlapped{HEvent: ev}
	var n uint32
	if write {
		err = windows.WriteFile(p.h, b, &n, &ov)
	} else {
		err = windows.ReadFile(p.h, b, &n, &ov)
	}
	if err != nil && err != windows.ERROR_IO_PENDING {
		return 0, err
	}
	ms := uint32(timeout / time.Millisecond)
	if ev, _ := windows.WaitForSingleObject(ev, ms); ev == uint32(windows.WAIT_TIMEOUT) {
		windows.CancelIoEx(p.h, &ov)
		return 0, errTimeout
	}
	if err := windows.GetOverlappedResult(p.h, &ov, &n, true); err != nil {
		return 0, err
	}
	return int(n), nil
}

func (p *hidPort) Write(b []byte) error {
	_, err := p.io(true, b, 2*time.Second)
	if err != nil {
		return fmt.Errorf("invio alla radio USB non riuscito: %v", err)
	}
	return nil
}

func (p *hidPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for len(p.rbuf) < len(b) {
		if ctx.Err() != nil {
			return errCancelled
		}
		left := time.Until(deadline)
		if left <= 0 {
			return errTimeout
		}
		// un rapporto di ingresso: 42 byte (identificativo 0 e 41 di dati),
		// come il buffer di hid-windows.c
		buf := make([]byte, 42)
		n, err := p.io(false, buf, left)
		if err != nil {
			return err
		}
		p.rbuf = append(p.rbuf, buf[:n]...)
	}
	copy(b, p.rbuf)
	p.rbuf = p.rbuf[len(b):]
	return nil
}

func (p *hidPort) Purge() error      { p.rbuf = nil; return nil }
func (p *hidPort) SetBaud(int) error { return nil }
func (p *hidPort) Close() error      { return windows.CloseHandle(p.h) }

// ---------------------------------------------------------------- DFU (driver STTub30)

// Radio in modalità DFU (DM-1701): come dfu-windows.c di dmrconfig, le richieste
// di controllo passano al driver «STTub30» di STMicroelectronics (installato dal
// pacchetto del costruttore) con DeviceIoControl(PU_VENDOR_REQUEST). Il Port
// riceve le richieste nel formato di drv_tyt.go (8 byte + dati in uscita) e
// restituisce l'esito (4 byte) e i dati in entrata.
type dfuPort struct {
	h    windows.Handle
	rbuf []byte
	wbuf []byte
}

// CTL_CODE(FILE_DEVICE_UNKNOWN, 0x0805, METHOD_BUFFERED, FILE_ANY_ACCESS)
const puVendorRequest = 0x22<<16 | 0x0805<<2

func openDFUPort(name string) (Port, error) {
	vid, pid, err := usbIDs(name)
	if err != nil {
		return nil, err
	}
	if vid != 0x0483 || pid != 0xDF11 {
		return nil, fmt.Errorf("dispositivo DFU %04X:%04X non gestito", vid, pid)
	}
	// GUID dell'interfaccia del driver STTub30 (dfu-windows.c)
	guid := windows.GUID{Data1: 0x3fe809ab, Data2: 0xfb91, Data3: 0x4cb5, Data4: [8]byte{0xa6, 0x43, 0x69, 0x67, 0x0d, 0x52, 0x36, 0x6e}}
	paths := usbDevicePaths(guid)
	if len(paths) == 0 {
		return nil, fmt.Errorf("radio non trovata sulla porta USB (modalità DFU, dispositivo 0483:DF11): collega il cavo USB, accendi la radio " +
			"e installa il driver USB del costruttore (cartella «DM-1701 USB Driver» del pacchetto «DM-1701 CPS»)")
	}
	p, err := windows.UTF16PtrFromString(paths[0])
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire la radio USB (DFU): %v. Chiudi il programma del costruttore se è aperto", err)
	}
	return &dfuPort{h: h}, nil
}

// CNTRPIPE_RQ del driver: USHORT Function, ULONG Direction, UCHAR Request,
// USHORT Value, USHORT Index, ULONG Length (allineamento naturale: 20 byte)
func cntrpipeRq(in bool, req byte, value, index uint16, length int) []byte {
	b := make([]byte, 20)
	b[0], b[1] = 0x1B, 0x00 // URB_FUNCTION_CLASS_INTERFACE
	if in {
		b[4] = 1
	}
	b[8] = req
	b[10], b[11] = byte(value), byte(value>>8)
	b[12], b[13] = byte(index), byte(index>>8)
	b[16], b[17], b[18], b[19] = byte(length), byte(length>>8), byte(length>>16), byte(length>>24)
	return b
}

func (p *dfuPort) Write(b []byte) error {
	p.wbuf = append(p.wbuf, b...)
	for len(p.wbuf) >= 8 {
		h := p.wbuf
		in := h[0]&0x80 != 0
		n := int(h[6]) | int(h[7])<<8
		need := 8
		if !in {
			need += n
		}
		if len(p.wbuf) < need {
			return nil
		}
		req, value, index := h[1], uint16(h[2])|uint16(h[3])<<8, uint16(h[4])|uint16(h[5])<<8
		var res int32
		var data []byte
		var got uint32
		if in {
			rq := cntrpipeRq(true, req, value, index, n)
			data = make([]byte, n)
			var out *byte
			if n > 0 {
				out = &data[0]
			}
			if err := windows.DeviceIoControl(p.h, puVendorRequest, &rq[0], uint32(len(rq)), out, uint32(n), &got, nil); err != nil || int(got) != n {
				res = -1
			} else {
				res = int32(n)
			}
		} else {
			rq := append(cntrpipeRq(false, req, value, index, n), h[8:8+n]...)
			if err := windows.DeviceIoControl(p.h, puVendorRequest, &rq[0], uint32(len(rq)), nil, 0, &got, nil); err != nil {
				res = -1
			} else {
				res = int32(n)
			}
		}
		r := []byte{byte(res), byte(res >> 8), byte(res >> 16), byte(res >> 24)}
		p.rbuf = append(p.rbuf, r...)
		if in && res > 0 {
			p.rbuf = append(p.rbuf, data...)
		}
		p.wbuf = p.wbuf[need:]
	}
	return nil
}

func (p *dfuPort) ReadFull(ctx context.Context, b []byte, timeout time.Duration) error {
	if len(p.rbuf) < len(b) {
		return errTimeout
	}
	copy(b, p.rbuf)
	p.rbuf = p.rbuf[len(b):]
	return nil
}

func (p *dfuPort) Purge() error      { p.rbuf, p.wbuf = nil, nil; return nil }
func (p *dfuPort) SetBaud(int) error { return nil }
func (p *dfuPort) Close() error      { return windows.CloseHandle(p.h) }
