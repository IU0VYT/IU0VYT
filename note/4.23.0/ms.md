# IU0VYT Flash 4.23.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Halaman keluaran di GitHub (dalam bahasa Inggeris), dengan muat turun: [V.4.23.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.23.0)

IU0VYT Flash V.4.23.0 untuk Microsoft Windows 10/11

### Baharu
- **Kenwood TK-481 dan TK-981** (900 MHz, saluran dibahagi kepada sistem) dan radio kenderaan **TK-780 dan TK-880**: saluran dan tetapan, dengan pemacu CHIRP.
- **Kenwood TK-2180 / TK-3180 / TK-7180 / TK-8180**: kini juga tetapan radio.
- **Radtel RT-13**: saluran (22, tanpa nama), diperoleh daripada program pengeluar.
- **Radtel RT-69 GPS**: saluran (999, dengan nama 12 aksara), diperoleh daripada program pengeluar; sebelum menulis, setiap saluran disemak berada dalam jalur radio.
- Lebih kukuh dengan fail rosak: dibetulkan dua ranap semasa membaca nada Baofeng UV-B5 / UV-B6 dan UV-3R.
- Penukaran: saluran yang tidak dapat disimpan oleh radio sasaran dilangkau dengan amaran dan tidak lagi menghentikan penukaran (perisian tegar IJV, nicFW, OpenGD77, UV-KX, R3b0rn, Radtel RT-480 / RT-68, RT-830 / RT-850).
- Galeri: 569 model daripada 47 jenama (515 dengan saluran, 433 dengan tetapan).

### Nota
- TK-481 / TK-981: hanya sistem yang sudah ada dalam radio digunakan; sistem sumber kongsi (trunking) tidak diubah.
- RT-13 dan RT-69 GPS: jika kata laluan ditetapkan pada radio, buang dahulu dengan program Radtel. Diperoleh daripada program pengeluar dan diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu.
- Program bebas, tidak bergabung dengan pengeluar radio.

---

Versi ini juga merangkumi perkara baharu versi 4.13.0 hingga 4.22.0, yang diterbitkan hanya sebagai versi ujian. Inilah ia, bermula daripada yang terkini.

## IU0VYT Flash 4.22.0

### Baharu
- **Radtel RT-480 dan RT-68**: saluran dan nama (256 saluran, bernombor 0 hingga 255, dengan nama 12 aksara), diperoleh daripada program pengeluar. Frekuensi, nama, kuasa, lebar jalur, imbasan dan nada boleh diubah; item saluran lain kekal seperti asal.
- **Radtel RT-830 dan RT-850**: saluran (128, tanpa nama), diperoleh daripada program pengeluar. Frekuensi, kuasa, lebar jalur, imbasan dan nada boleh diubah.
- **Radtel RT-20M**: dibuka dengan tab Ruyage UV58Plus (memori sama dengan iRadio UV-5118plus).

### Nota
- RT-480 / RT-68: hanya saluran dan nama ditulis; tetapan, VFO dan DTMF radio tidak disentuh. Frekuensi dalam langkah 1.25 kHz; setiap saluran mesti mempunyai frekuensi pancar.
- RT-830 / RT-850: tetapan dan VFO dikembalikan seperti asal; kata laluan radio tidak pernah ditulis. Jika kata laluan ditetapkan pada radio, buang dahulu dengan program Radtel.
- Semuanya diperoleh daripada program pengeluar dan diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu.

## IU0VYT Flash 4.21.0

### Baharu
- **Radtel RT-5D (DMR)**: saluran (16 zon 64, dengan nama 12 aksara), diperoleh daripada program pengeluar. Frekuensi, nama, kuasa, lebar jalur, imbasan dan nada boleh diubah; item saluran lain kekal seperti asal.
- **Kenwood TK-2180, TK-3180K, TK-3180K2, TK-3180E** dan radio kenderaan **TK-7180, TK-8180** (juga E): saluran, dengan pemacu CHIRP. Senarai mempunyai 250 kedudukan bagi setiap 32 zon pertama; hanya zon yang sudah ada dalam radio digunakan, 512 saluran kesemuanya.

### Nota
- RT-5D: bacaan mengambil seluruh memori dan penulisan menghantar semuanya kembali seperti dalam fail, sama seperti program Radtel. Sebelum menulis, buat bacaan baharu: fail lama akan mengembalikan kenalan dan tetapan masa itu. Diuji hanya dengan radio simulasi.
- TK-x180: saluran sahaja, tiada tetapan; zon tidak dicipta atau dipadam. Jika radio mempunyai zon bukan konvensional (trunking), bacaan ditolak.

## IU0VYT Flash 4.20.0

### Baharu
- **Radtel RT-4D dan RT-6D (DMR) dan RT-1000 Pro**: saluran (1024, dengan nama 16 aksara) dan zon (250), diperoleh daripada program pengeluar. Frekuensi, nama, kuasa, lebar jalur, nada dan, pada RT-4D dan RT-6D, imbasan boleh diubah; item saluran lain kekal seperti asal. Hanya saluran dan zon ditulis: kenalan, kumpulan, mesej, radio FM dan tetapan tidak disentuh.

### Nota
- Diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu dan laporkan sebarang masalah.
- «Kenal pasti» tidak membezakan radio ini daripada radio Radtel dan Bajeton lain dengan sambungan yang sama: tab mesti dipilih secara manual.
- RT-1000 Pro: pengecualian imbasan tidak diurus (kekal seperti pada radio).

## IU0VYT Flash 4.19.0

### Baharu
- **Radtel RT-817A / RT-871A**: saluran (1024, dengan nama 16 aksara), diperoleh daripada program pengeluar. Sambungan dan format saluran sama dengan RT-880G, tanpa zon. Hanya kawasan yang digunakan oleh program Radtel dibaca dan ditulis (tetapan, saluran, radio FM); logo tidak disentuh.

### Nota
- Diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu dan laporkan sebarang masalah.
- «Kenal pasti» tidak membezakan RT-817A / RT-871A daripada RT-880G: tab mesti dipilih secara manual.

## IU0VYT Flash 4.18.0

### Baharu
- **Kenwood TK-2140, TK-3140, TK-280, TK-380**: radio genggam profesional dengan saluran dibahagi kepada zon (kumpulan), saluran dan tetapan. Senarai menunjukkan 250 kedudukan memori; saluran baharu masuk ke zon saluran sebelumnya.
- **Quansheng UV-K5 / UV-K6 dengan perisian tegar R3b0rn oleh fagci**: saluran (298 dengan memori asal, hingga 1024 dengan memori lebih besar). Tetapan, VFO, pratetap dan pembetulan SSB tidak pernah ditulis.
- **Radtel RT-330 dan RT-400**: saluran, diperoleh daripada program pengeluar.

### Nota
- Pemacu CHIRP untuk radio genggam TK diisytiharkan eksperimen. Zon tidak dicipta atau dipadam daripada program.
- Perisian tegar R3b0rn, RT-330 dan RT-400 diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu dan laporkan sebarang masalah.
- Dengan perisian tegar R3b0rn, saluran tanpa nama mengambil frekuensi sebagai nama (bagi perisian tegar itu, saluran tanpa nama tidak wujud).
- Tab RT-400 menolak radio yang memperkenalkan diri sebagai RT-470, RT-630, RT-495 atau HI-8811: gunakannya hanya dengan RT-400.

## IU0VYT Flash 4.17.0

### Baharu
- **Kenwood TH-F6A, TH-F7E, TH-G71, TH-D7, TH-K2**: saluran dibaca dan ditulis satu demi satu, dengan radio dihidupkan seperti biasa (9600 baud). Semasa menulis, hanya saluran yang berubah dihantar.
- **Radio genggam profesional Kenwood TK**: saluran dan tetapan.
  - Siri 60: TK-260, TK-270, TK-272, TK-278, TK-360, TK-370, TK-372, TK-378.
  - Siri 60G: TK-260G, TK-270G, TK-272G, TK-278G, TK-360G, TK-370G, TK-372G, TK-378G, TK-388G.
- **Radtel RT-860**: gunakan tab «Radtel RT-880G» (sambungan dan memori saluran yang sama).

### Nota
- Kenwood TH-F6/F7, TH-G71, TH-D7 dan TH-K2 menggunakan arahan pemacu CHIRP dan diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu dan laporkan sebarang masalah.
- Pemacu CHIRP untuk radio genggam TK diisytiharkan eksperimen.
- RT-860 diperoleh daripada perbandingan dengan program pengeluar dan belum diuji pada radio sebenar.

## IU0VYT Flash 4.16.0

### Baharu
- **OpenGD77**: saluran radio DMR dengan perisian tegar bebas OpenGD77.
  - Radioddity GD-77 dan GD-77S, Baofeng DM-1801 dan RD-5R.
  - Baofeng DM-1701, TYT MD-UV380, MD-380, MD-9600 dan MD-2017, Retevis RT3S dan RT84.
  - Radio muncul sebagai port bersiri USB; hanya saluran dan zon yang berubah ditulis, tidak pernah tetapan, kenalan, kumpulan dan kalibrasi.
- **Quansheng UV-K5, perisian tegar bebas lain**:
  - **NUNU** oleh kamilsss655: saluran dan tetapan.
  - **UV-KX** oleh joaquim.org: saluran dan tetapan; kalibrasi tidak pernah ditulis.
  - Tab «OSFW» juga sesuai untuk perisian tegar asal yang diubah suai dengan jalur diperluas (uvmod-kitchen, koleksi Tunas1337).
- **Radtel RT-880** tanpa GPS dengan perisian tegar asal: gunakan tab «Radtel RT-880G».

### Nota
- OpenGD77 ditulis berdasarkan protokol yang diterangkan dalam program bebas qdmr dan diuji hanya dengan radio simulasi: buat bacaan sahaja dahulu, simpan salinan sandaran dan laporkan sebarang masalah.
- NUNU dan UV-KX dibandingkan bait demi bait dengan pemacu CHIRP oleh pengarangnya.

## IU0VYT Flash 4.15.0

### Baharu
- **D-STAR dan C4FM**: saluran melalui kabel untuk radio genggam digital amatur, dengan fail .img yang serasi dengan CHIRP. Lajur «Mod» menunjukkan «D-STAR» atau «C4FM».
  - **Icom** ID-31A/E, ID-51A/E, ID-51 Plus, ID-51 Plus2 (D-STAR) dan IC-E90 / IC-T90.
  - **Yaesu** VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D (salinan bermula dari radio: arahan ada dalam tetingkap kabel).
  - **Kenwood** (jenama baharu, 47 kesemuanya) TH-D72, TH-D74, TH-D75, dengan kabel USB.
  - Tetapan untuk IC-E90, VX-6, VX-3, FT-70D, FT1D, FT2D, FT3D, FT5D dan TH-D72.
- **Perisian tegar bebas**:
  - **nicFW 2.x** untuk TIDRADIO TD-H3: saluran dan tetapan utama.
  - **nicFW880 / nicFW950** untuk Radtel RT-880, RT-880G dan RT-950 Pro: saluran.
  - **F4HWN Fusion 5.x** untuk Quansheng UV-K1 dan UV-K5 V3: 1024 saluran dan tetapan.
  - Dengan nicFW, kalibrasi radio tidak pernah ditulis.
- **Kenwood dalam galeri**: juga radio genggam PMR446 dan DMR/NXDN, dalam senarai tidak serasi (hanya boleh diprogram dengan program KPG yang dikhaskan untuk pengedar).

### Pembetulan
- Tetapan: medan teks yang tidak pernah digunakan dibaca kosong dan bukan «nilai tidak dijangka»; nilai yang tidak berubah tidak ditulis semula.
- Puxing PX-888K: dibuang enam item tetapan yang boleh menulis data salah (kod BOT/EOT dan masa DTMF).
- Yaesu VX-3: ruang tidak lagi dibenarkan dalam kod DTMF (radio tidak memilikinya).

### Nota
- Pemacu ditulis dari awal berdasarkan pemacu CHIRP (GNU GPL), dokumentasi yang diterbitkan oleh nicsure (nicFW) dan pemacu F4HWN, serta dibandingkan bait demi bait menggunakan radio simulasi. Radio sebenar versi ini belum diuji: simpan salinan sandaran dan laporkan sebarang masalah.
- Bagi Icom, Yaesu dan Kenwood tiada kemas kini perisian tegar (ia dibuat dengan program pengeluar).

## IU0VYT Flash 4.14.0

### Baharu
- **Yaesu dan Icom**: dua jenama baharu (46 kesemuanya) dan 13 radio lagi, dengan saluran melalui kabel dan fail .img yang serasi dengan CHIRP.
  - **Yaesu** FT-4XR, FT-4XE, FT-4VR, FT-65R, FT-65E, FT-25R (kabel SCU-35) dan FT-60R/E (salinan dimulakan dari radio, «F8 CLONE»).
  - **Icom** IC-V86, IC-T10, IC-V80, IC-T70, IC-V82, IC-U82 (kabel OPC-478).
  - Tetapan untuk semua kecuali IC-T70. Tiada kemas kini perisian tegar untuk radio ini.
- **Dua bahasa lagi**: bahasa Melayu dan Hindi (हिन्दी), 13 kesemuanya: program, laman web, panduan dan fail maklumat, README dan dasar keselamatan.
- **Penukar dan «Majukan ke radio lain»** ke Yaesu dan Icom: nada yang tiada pada radio dibuang dengan amaran dan saluran yang tidak dapat disimpan dilangkau dengan amaran, dan tidak lagi menghentikan keseluruhan operasi.

### Pembetulan
- Terjemahan: lima perkataan tebal dalam panduan kekal dalam bahasa Itali dalam semua bahasa; dua ayat salah dalam bahasa Korea dan Arab.
- Wouxun KG-UV8D: nama medan dalaman diselaraskan dengan CHIRP semasa (tiada kesan pada saluran).

### Nota
- Pemacu ditulis dari awal berdasarkan pemacu CHIRP (GNU GPL) dan dibandingkan bait demi bait dengan CHIRP menggunakan radio simulasi. Radio Yaesu dan Icom sebenar belum diuji: simpan salinan sandaran dan laporkan sebarang masalah.
- Terjemahan bahasa Melayu dan Hindi adalah baharu: laporkan teks yang kurang jelas.

## IU0VYT Flash 4.13.1

### Baharu
- **Perisian tegar IJV** untuk Quansheng UV-K5, UV-K5 V3 dan UV-K1 (perisian tegar tidak rasmi oleh IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): lima tab, satu bagi setiap versi perisian tegar.
  - **V3** dengan 200 atau 999 saluran pada UV-K5 klasik (perkakasan V1).
  - **V4**, **V5 («FW5»)** dan **V6** pada UV-K5 V3 dan UV-K1: 500 saluran bank yang dipilih pada radio, dengan nama kumpulan dan bank.
  - Saluran dan tetapan utama (menu, paparan, VOX, kekunci sisi) dibaca dan ditulis. Sebelum menulis, program menyemak bahawa bank yang sama masih dipilih pada radio dan hanya menulis bahagian yang berubah; kalibrasi tidak pernah disentuh.
  - **V6 (baharu dalam 4.13.1)**: pautan data V6, penulisan mengikut sektor 4 KiB, setiap satu dibaca semula dan disemak. Jika radio menyatakan pautan data tiada, kemas kininya dengan FLASH V6 daripada RadioManager.
- **Kemas kini perisian tegar**: fail .bin dengan versi IJV ditanda, bersama tab yang perlu digunakan untuk saluran dan amaran untuk V5 dan V6.
- Terjemahan dalam kesemua 11 bahasa.

### Nota
- Protokol dan memori diperoleh daripada modul CHIRP yang diterbitkan oleh IJV dan daripada analisis RadioManager 4.3.2; kod ditulis dari awal, diuji dengan radio simulasi. Radio sebenar dengan IJV belum diuji: laporkan sebarang masalah.
- Modul, item, lesen dan perisian tegar IJV kekal milik RadioManager oleh IJV.

## IU0VYT Flash 4.13.0

### Baharu
- **Perisian tegar IJV** untuk Quansheng UV-K5, UV-K5 V3 dan UV-K1 (perisian tegar tidak rasmi oleh IU0IJV, [ijvradio.com](https://www.ijvradio.com/)): empat tab baharu, satu bagi setiap versi perisian tegar.
  - **V3** dengan 200 atau 999 saluran pada UV-K5 klasik (perkakasan V1).
  - **V4** dan **V5 («FW5»)** pada UV-K5 V3 dan UV-K1: 500 saluran bank yang dipilih pada radio, dengan nama kumpulan dan bank.
  - Saluran dan tetapan utama (menu, paparan, VOX, kekunci sisi) dibaca dan ditulis. Sebelum menulis, program menyemak bahawa bank yang sama masih dipilih pada radio dan hanya menulis blok yang berubah; kalibrasi tidak pernah disentuh.
  - **V6** dikenali: saluran dan tetapan V6 diurus dengan RadioManager oleh IJV.
- **Kemas kini perisian tegar**: fail .bin dengan versi IJV ditanda, bersama tab yang perlu digunakan untuk saluran dan amaran untuk V5 dan V6.
- Terjemahan dalam kesemua 11 bahasa.

### Nota
- Protokol dan memori diperoleh daripada modul CHIRP yang diterbitkan oleh IJV dan daripada analisis RadioManager 4.3.2; kod ditulis dari awal, diuji dengan radio simulasi. Radio sebenar dengan IJV belum diuji: laporkan sebarang masalah.
