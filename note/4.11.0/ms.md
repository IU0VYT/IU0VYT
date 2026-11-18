# IU0VYT Flash 4.11.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
<td align="center"><a href="it.md"><img src="../../docs/img/flags/it.svg" width="36" height="24" alt=""><br><sub>Italiano</sub></a></td>
<td align="center"><a href="fr.md"><img src="../../docs/img/flags/fr.svg" width="36" height="24" alt=""><br><sub>Français</sub></a></td>
<td align="center"><a href="de.md"><img src="../../docs/img/flags/de.svg" width="36" height="24" alt=""><br><sub>Deutsch</sub></a></td>
<td align="center"><a href="es.md"><img src="../../docs/img/flags/es.svg" width="36" height="24" alt=""><br><sub>Español</sub></a></td>
<td align="center"><a href="pt.md"><img src="../../docs/img/flags/pt.svg" width="36" height="24" alt=""><br><sub>Português</sub></a></td>
<td align="center"><a href="ru.md"><img src="../../docs/img/flags/ru.svg" width="36" height="24" alt=""><br><sub>Русский</sub></a></td>
<td align="center"><a href="zh.md"><img src="../../docs/img/flags/zh.svg" width="36" height="24" alt=""><br><sub>中文</sub></a></td>
<td align="center"><a href="ja.md"><img src="../../docs/img/flags/ja.svg" width="36" height="24" alt=""><br><sub>日本語</sub></a></td>
<td align="center"><a href="ko.md"><img src="../../docs/img/flags/ko.svg" width="36" height="24" alt=""><br><sub>한국어</sub></a></td>
<td align="center"><a href="ar.md"><img src="../../docs/img/flags/ar.svg" width="36" height="24" alt=""><br><sub>العربية</sub></a></td>
<td align="center"><img src="../../docs/img/flags/ms.svg" width="36" height="24" alt=""><br><sub><b>Bahasa Melayu</b></sub></td>
<td align="center"><a href="hi.md"><img src="../../docs/img/flags/hi.svg" width="36" height="24" alt=""><br><sub>हिन्दी</sub></a></td>
</tr>
</table>
</div>

Halaman keluaran di GitHub (dalam bahasa Inggeris), dengan muat turun: [V.4.11.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.11.0)

IU0VYT Flash V.4.11.0 untuk Microsoft Windows 10/11

### Nama zon
- Sebagai ganti «Zon 1», «Zon 2»… setiap zon boleh mempunyai nama, seperti dalam program pengeluar: butang «✎ Nama zon…» atau klik dua kali pada zon. Bar zon dan baris bawah menunjukkan «nombor · nama».
- Radio: Radtel RT-950 Pro, Bajeton BJ-9000 dan Hiroyasu HI-UV850 Pro (perisian tegar hingga 0.23 dan dari 0.24), Radtel RT-950, Radtel RT-920, Radtel RT-910 / Bajeton BJ-7700 tanpa Bluetooth; zon radio DMR Baofeng RD-5R, DM-1801 dan DM-32UV.
- Nama dibaca dan ditulis melalui kabel bersama saluran dan disimpan dalam fail .img. Panjang dan aksara seperti dalam program pengeluar (12 aksara, termasuk Cina, pada RT-950 Pro; 10 pada RT-920; 16 pada DMR).
- Keselamatan: kawasan nama hanya ditulis jika nama berubah, selepas dibaca semula daripada radio; jika radio menyimpan data di situ yang bukan nama, tiada apa-apa ditulis.

### Edit pukal
- «✎ Edit pukal…» (juga melalui klik kanan) mengubah serentak kuasa, lebar jalur, imbasan, nada TX, nada RX, anjakan dan pemancaran bagi saluran yang ditanda (atau dipilih). Item yang dibiarkan pada «jangan ubah» kekal seperti asal; Ctrl+Z membatalkan semuanya.

### Dua bahasa lagi (11 kesemuanya)
- Bahasa Korea, dengan bendera Korea Selatan.
- Bahasa Arab, dengan bendera Emiriah Arab Bersatu dan halaman dari kanan ke kiri (frekuensi dan nombor kekal dari kiri ke kanan).
- Semua fail maklumat dalam folder `lingue` turut diterjemahkan ke bahasa Korea dan Arab serta dikemas kini ke 4.11 dalam semua bahasa.

### Lain-lain
- Laman web: bahasa yang tersedia dinyatakan; senarai fail perisian tegar dibuang.
- `go test` lulus pada Windows (termasuk i18n_test.go); ujian baharu menyemak bahawa LEGGIMI.txt, AUTORI_E_LICENZE.txt dan lesen adalah sama dalam folder utama dan dalam «sorgenti».

Format baharu diperoleh daripada kod program pengeluar dan pemacu CHIRP serta diuji dengan radio simulasi; ia belum diuji dengan radio sebenar: sentiasa simpan salinan sandaran.

Galeri: 489 model daripada 44 jenama (451 dengan saluran, 396 dengan tetapan, 57 dengan kemas kini perisian tegar).
