# IU0VYT Flash 4.46.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Halaman keluaran di GitHub (dalam bahasa Inggeris), dengan muat turun: [V.4.46.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.46.0)

IU0VYT Flash V.4.46.0 untuk Microsoft Windows 10/11

### Pembetulan
- **Radtel RT-880G**: tidak lagi ditunjukkan sebagai radio GMRS / FRS. Ia ialah pemancar-penerima radio amatur VHF/UHF (dengan GPS dan APRS), seperti RT-880 dan RT-860; amaran «Radio dijual sebagai GMRS» tidak lagi muncul sebelum penulisan.
- **Pengesahan selepas penulisan** pada Radtel RT-880G: ia gagal kerana radio, yang pada akhir penulisan menyimpan data dan dimulakan semula, dihubungi terlalu awal dan terlalu kerap. Kini program menunggu 10 saat tanpa mengganggunya dan mencuba semula setiap 15 saat, sehingga 90 saat. Masa yang sama digunakan untuk RT-880, RT-860 dan RT-817A / RT-871A, yang menggunakan pemacu yang sama.
- Semakan lebih ketat pada nilai yang diterima daripada AR8200, Kenwood dan TK-180 (laporan CodeQL).

### Nota
- Pengesahan telah diuji pada RT-880G sebenar; RT-880, RT-860 dan RT-817A / RT-871A belum diuji pada radio sebenar. Radio lain mengekalkan masa yang sebelumnya.
- Tiada radio baharu dalam versi ini.
- Program bebas, tidak bergabung dengan pengeluar radio.
- Rangkaian ujian penuh lulus: bahagian umum; 915 perbandingan struktur dan 380 bacaan / penulisan yang sama dengan CHIRP.
