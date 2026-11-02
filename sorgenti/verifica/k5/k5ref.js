// IU0VYT Flash — confronto dell'aggiornamento del firmware della UV-K5 con uvtools
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi - GNU GPL v3 o successiva
//
// Uso: node k5ref.js <cartella js di uvtools> <firmware.bin>
// Con le funzioni di uvtools (https://github.com/egzumer/uvtools: fwpack.js,
// qsSerial.js, tool_patcher.js) decifra il file e prepara i pacchetti che
// uvtools invia alla radio; stampa in JSON versione, firmware e pacchetti (hex).
// Usato da TestK5FirmwareUvtools (variabile K5REF_JS).
const fs = require('fs');
const path = require('path');
const [dir, fwPath] = process.argv.slice(2);
const read = f => fs.readFileSync(path.join(dir, f), 'utf8');
const pick = (src, re) => { const m = src.match(re); if (!m) throw new Error('funzione non trovata: ' + re); return m[0]; };
const fp = read('fwpack.js'), qs = read('qsSerial.js'), tp = read('tool_patcher.js');
const code = [
  'let rawVersion = null; const log = () => {};',
  pick(fp, /const Crc16Tab = \[[\s\S]*?\];/),
  pick(fp, /function crc16_ccitt\(data\) \{[\s\S]*?\n\}/),
  pick(fp, /function crc16_ccitt_le\(data\) \{[\s\S]*?\n\}/),
  pick(fp, /function firmware_xor\(fwcontent\) \{[\s\S]*?\n\}/),
  pick(fp, /function unpack\(encoded_firmware\) \{[\s\S]*?\n\}/),
  pick(qs, /function xor\(data\) \{[\s\S]*?\n\}/),
  pick(qs, /function crc16xmodem\(data, crc = 0\) \{[\s\S]*?\n\}/),
  pick(qs, /function packetize\(data\) \{[\s\S]*?\n\}/),
  pick(tp, /function flash_generateCommand\(data, address, totalSize\) \{[\s\S]*?\n\}/),
  'return { unpack, packetize, flash_generateCommand, getVersion: () => rawVersion };',
].join('\n');
const u = new Function(code)();
const hex = b => Buffer.from(b).toString('hex');
const raw = new Uint8Array(fs.readFileSync(fwPath));
const fw = u.unpack(raw);
const ver = u.getVersion();
// come flash_init: 0x30 0x05 lunghezza 0x00 + versione
const packets = [u.packetize(new Uint8Array([0x30, 0x5, ver.length, 0x0, ...ver]))];
for (let i = 0; i < fw.length; i += 0x100) packets.push(u.packetize(u.flash_generateCommand(fw.slice(i, i + 0x100), i, fw.length)));
console.log(JSON.stringify({ version: hex(ver), firmware: hex(fw), packets: packets.map(hex) }));
