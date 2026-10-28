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
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
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

// ---------------------------------------------------------------- DFU (driver STTub30 o WinUSB)

// Radio in modalità DFU (DM-1701): le richieste di controllo USB passano
//   - al driver «STTub30» di STMicroelectronics (installato dal pacchetto del
//     costruttore) con DeviceIoControl(PU_VENDOR_REQUEST), come dfu-windows.c di
//     dmrconfig;
//   - oppure (3.8) al driver generico di Windows «WinUSB» (installabile con
//     Zadig), con WinUsb_ControlTransfer, come dfu-libusb.c di dmrconfig con
//     libusb su Windows.
//
// Il Port (dfuPort, dfuport.go) riceve le richieste nel formato di drv_tyt.go
// (8 byte + dati in uscita) e restituisce l'esito (4 byte) e i dati in entrata.

// CTL_CODE(FILE_DEVICE_UNKNOWN, 0x0805, METHOD_BUFFERED, FILE_ANY_ACCESS)
const puVendorRequest = 0x22<<16 | 0x0805<<2

var (
	modWinusb                 = windows.NewLazySystemDLL("winusb.dll")
	procWinUsbInitialize      = modWinusb.NewProc("WinUsb_Initialize")
	procWinUsbFree            = modWinusb.NewProc("WinUsb_Free")
	procWinUsbControlTransfer = modWinusb.NewProc("WinUsb_ControlTransfer")
	procWinUsbSetPipePolicy   = modWinusb.NewProc("WinUsb_SetPipePolicy")
	// GUID_DEVINTERFACE_USB_DEVICE: tutti i dispositivi USB presenti
	guidUSBDevice = windows.GUID{Data1: 0xA5DCBF10, Data2: 0x6530, Data3: 0x11D2, Data4: [8]byte{0x90, 0x1F, 0x00, 0xC0, 0x4F, 0xB9, 0x51, 0xED}}
)

const dfuNoDriver = "la radio è collegata ma Windows non ha un driver adatto per la modalità DFU (dispositivo 0483:DF11). " +
	"Installa il driver USB del costruttore (pulsante «Installa il driver USB…», cartella «DM-1701 USB Driver» del pacchetto «DM-1701 CPS») " +
	"oppure il driver WinUSB con Zadig (https://zadig.akeo.ie: Options → List All Devices, scegli «STM32 BOOTLOADER» o 0483 DF11, driver WinUSB, Install Driver)"

var procSetupDiOpenDevRegKey = modSetupapi.NewProc("SetupDiOpenDevRegKey")

// usbPresent: percorsi dei dispositivi USB presenti con questi VID e PID: quelli
// dell'interfaccia generica dei dispositivi USB e quelli delle interfacce
// dichiarate dal driver nel registro («DeviceInterfaceGUIDs», scritte per
// WinUSB da Zadig / libwdi), come fa libusb su Windows.
func usbPresent(vid, pid uint16) []string {
	key := strings.ToLower(fmt.Sprintf("vid_%04x&pid_%04x", vid, pid))
	var out []string
	seen := map[string]bool{}
	add := func(paths []string) {
		for _, p := range paths {
			if l := strings.ToLower(p); strings.Contains(l, key) && !seen[l] {
				seen[l] = true
				out = append(out, p)
			}
		}
	}
	add(usbDevicePaths(guidUSBDevice))
	for _, g := range usbDriverInterfaceGUIDs(key) {
		add(usbDevicePaths(g))
	}
	return out
}

// usbDriverInterfaceGUIDs: GUID delle interfacce scritti nel registro per i
// dispositivi USB presenti con l'identificativo key («vid_0483&pid_df11»)
func usbDriverInterfaceGUIDs(key string) []windows.GUID {
	const digcfPresent, digcfAllClasses = 0x2, 0x4
	enum, _ := windows.UTF16PtrFromString("USB")
	h, _, _ := procSetupDiGetClassDevsW.Call(0, uintptr(unsafe.Pointer(enum)), 0, digcfPresent|digcfAllClasses)
	if h == 0 || h == ^uintptr(0) {
		return nil
	}
	defer procSetupDiDestroyDeviceInfoList.Call(h)
	var out []windows.GUID
	for i := 0; i < 512; i++ {
		var d spDevinfoData
		d.cbSize = uint32(unsafe.Sizeof(d))
		if r, _, _ := procSetupDiEnumDeviceInfo.Call(h, uintptr(i), uintptr(unsafe.Pointer(&d))); r == 0 {
			break
		}
		buf := make([]uint16, 1024)
		const spdrpHardwareID = 0x1
		if r, _, _ := procSetupDiGetDeviceRegistryPropW.Call(h, uintptr(unsafe.Pointer(&d)), spdrpHardwareID, 0,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)*2), 0); r == 0 {
			continue
		}
		if !strings.Contains(strings.ToLower(windows.UTF16ToString(buf)), key) {
			continue
		}
		// chiave del dispositivo (DICS_FLAG_GLOBAL, DIREG_DEV, KEY_READ)
		k, _, _ := procSetupDiOpenDevRegKey.Call(h, uintptr(unsafe.Pointer(&d)), 1, 0, 1, windows.KEY_READ)
		if k == 0 || k == ^uintptr(0) {
			continue
		}
		rk := registryKey(windows.Handle(k))
		for _, name := range []string{"DeviceInterfaceGUIDs", "DeviceInterfaceGUID"} {
			for _, gs := range rk.strings(name) {
				if g, err := windows.GUIDFromString(strings.TrimSpace(gs)); err == nil {
					out = append(out, g)
				}
			}
		}
		windows.RegCloseKey(windows.Handle(k))
	}
	return out
}

type registryKey windows.Handle

// strings: valore REG_SZ o REG_MULTI_SZ come elenco di stringhe
func (k registryKey) strings(name string) []string {
	n, _ := windows.UTF16PtrFromString(name)
	var typ, size uint32
	if windows.RegQueryValueEx(windows.Handle(k), n, nil, &typ, nil, &size) != nil || size < 2 || size > 8192 {
		return nil
	}
	buf := make([]uint16, size/2+1)
	if windows.RegQueryValueEx(windows.Handle(k), n, nil, &typ, (*byte)(unsafe.Pointer(&buf[0])), &size) != nil {
		return nil
	}
	var out []string
	for _, part := range strings.Split(string(utf16.Decode(buf[:size/2])), "\x00") {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func openDFUPort(name string) (Port, error) {
	vid, pid, err := usbIDs(name)
	if err != nil {
		return nil, err
	}
	if vid != 0x0483 || pid != 0xDF11 {
		return nil, fmt.Errorf("dispositivo DFU %04X:%04X non gestito", vid, pid)
	}
	// 1) driver del costruttore (STTub30): GUID dell'interfaccia (dfu-windows.c)
	guid := windows.GUID{Data1: 0x3fe809ab, Data2: 0xfb91, Data3: 0x4cb5, Data4: [8]byte{0xa6, 0x43, 0x69, 0x67, 0x0d, 0x52, 0x36, 0x6e}}
	if paths := usbDevicePaths(guid); len(paths) > 0 {
		return openSTTubPort(paths[0])
	}
	// 2) WinUSB
	paths := usbPresent(vid, pid)
	if len(paths) == 0 {
		return nil, fmt.Errorf("radio non trovata sulla porta USB (modalità DFU, dispositivo 0483:DF11): collega il cavo USB, accendi la radio " +
			"e installa il driver USB del costruttore (cartella «DM-1701 USB Driver» del pacchetto «DM-1701 CPS») oppure WinUSB con Zadig")
	}
	var last error
	for _, p := range paths {
		port, err := openWinUSBPort(p)
		if err == nil {
			return port, nil
		}
		last = err
	}
	return nil, last
}

func openSTTubPort(path string) (Port, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire la radio USB (DFU): %v. Chiudi il programma del costruttore se è aperto", err)
	}
	port := &dfuPort{close: func() error { return windows.CloseHandle(h) }}
	port.xfer = func(in bool, rt, req byte, value, index uint16, out []byte, n int) (int32, []byte) {
		var got uint32
		if in {
			rq := cntrpipeRq(true, req, value, index, n)
			data := make([]byte, n)
			var o *byte
			if n > 0 {
				o = &data[0]
			}
			if err := windows.DeviceIoControl(h, puVendorRequest, &rq[0], uint32(len(rq)), o, uint32(n), &got, nil); err != nil || int(got) != n {
				return -1, nil
			}
			return int32(n), data
		}
		rq := append(cntrpipeRq(false, req, value, index, len(out)), out...)
		if err := windows.DeviceIoControl(h, puVendorRequest, &rq[0], uint32(len(rq)), nil, 0, &got, nil); err != nil {
			return -1, nil
		}
		return int32(len(out)), nil
	}
	return port, nil
}

// openWinUSBPort: dispositivo con il driver WinUSB (WinUsb_Initialize fallisce
// se il driver è un altro o manca)
func openWinUSBPort(path string) (Port, error) {
	if err := modWinusb.Load(); err != nil {
		return nil, fmt.Errorf("winusb.dll non disponibile: %v", err)
	}
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil,
		windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OVERLAPPED, 0)
	if err != nil {
		return nil, errors.New(dfuNoDriver)
	}
	var wu uintptr
	if r, _, _ := procWinUsbInitialize.Call(uintptr(h), uintptr(unsafe.Pointer(&wu))); r == 0 {
		windows.CloseHandle(h)
		return nil, errors.New(dfuNoDriver)
	}
	// tempo massimo delle richieste di controllo (PIPE_TRANSFER_TIMEOUT, ms)
	tmo := uint32(5000)
	procWinUsbSetPipePolicy.Call(wu, 0, 0x03, 4, uintptr(unsafe.Pointer(&tmo)))
	port := &dfuPort{close: func() error {
		procWinUsbFree.Call(wu)
		return windows.CloseHandle(h)
	}}
	port.xfer = func(in bool, rt, req byte, value, index uint16, out []byte, n int) (int32, []byte) {
		// WINUSB_SETUP_PACKET (8 byte, passato per valore)
		var sp [8]byte
		sp[0], sp[1] = rt, req
		binary.LittleEndian.PutUint16(sp[2:], value)
		binary.LittleEndian.PutUint16(sp[4:], index)
		buf := out
		if in {
			buf = make([]byte, n)
		}
		binary.LittleEndian.PutUint16(sp[6:], uint16(len(buf)))
		var bp *byte
		if len(buf) > 0 {
			bp = &buf[0]
		}
		var got uint32
		r, _, _ := procWinUsbControlTransfer.Call(wu, uintptr(binary.LittleEndian.Uint64(sp[:])), uintptr(unsafe.Pointer(bp)),
			uintptr(len(buf)), uintptr(unsafe.Pointer(&got)), 0)
		if r == 0 {
			return -1, nil
		}
		if in {
			return int32(got), buf[:got]
		}
		return int32(got), nil
	}
	return port, nil
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
