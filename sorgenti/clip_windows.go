// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

package main

// Testo negli appunti di Windows (3.9: indirizzo dell'autore copiato senza
// passare dalla pagina).

import (
	"runtime"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32Clip        = windows.NewLazySystemDLL("user32.dll")
	modKernel32Clip      = windows.NewLazySystemDLL("kernel32.dll")
	procOpenClipboard    = modUser32Clip.NewProc("OpenClipboard")
	procCloseClipboard   = modUser32Clip.NewProc("CloseClipboard")
	procEmptyClipboard   = modUser32Clip.NewProc("EmptyClipboard")
	procSetClipboardData = modUser32Clip.NewProc("SetClipboardData")
	procGlobalAlloc      = modKernel32Clip.NewProc("GlobalAlloc")
	procGlobalFree       = modKernel32Clip.NewProc("GlobalFree")
	procGlobalLock       = modKernel32Clip.NewProc("GlobalLock")
	procGlobalUnlock     = modKernel32Clip.NewProc("GlobalUnlock")
	procRtlMoveMemory    = modKernel32Clip.NewProc("RtlMoveMemory")
)

func init() {
	copyText = func(s string) bool {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		const gmemMoveable, cfUnicodeText = 0x0002, 13
		if r, _, _ := procOpenClipboard.Call(0); r == 0 {
			return false
		}
		defer procCloseClipboard.Call()
		procEmptyClipboard.Call()
		u := append(utf16.Encode([]rune(s)), 0)
		h, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(len(u)*2))
		if h == 0 {
			return false
		}
		p, _, _ := procGlobalLock.Call(h)
		if p == 0 {
			procGlobalFree.Call(h)
			return false
		}
		procRtlMoveMemory.Call(p, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)*2))
		procGlobalUnlock.Call(h)
		if r, _, _ := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
			procGlobalFree.Call(h)
			return false
		}
		return true
	}
}
