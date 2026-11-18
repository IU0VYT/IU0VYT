# IU0VYT Flash 4.4.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Halaman keluaran di GitHub (dalam bahasa Inggeris), dengan muat turun: [V.4.4.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.4.0)

IU0VYT Flash V.4.4.0 untuk Microsoft Windows 10/11

Versi ini juga merangkumi perkara baharu 4.3.0, 4.2.0, 4.1.0 dan 4.0.0, yang tidak mempunyai keluaran stabil sendiri.

## Baharu dalam 4.4.0

### Dibetulkan: kotak «Pengenalpastian Radio Sedang Berjalan» tanpa radio
- Kotak itu boleh muncul walaupun dengan kabel tanpa radio: gangguan talian dan gema arahan yang dihantar kelihatan seperti radio yang menjawab.
- Kini ia muncul **hanya untuk kabel yang disambungkan ke radio yang benar-benar menjawab**. Dengan dua kabel dan satu radio (COM6 tanpa radio, COM7 dengan radio, atau sebaliknya) ia muncul hanya untuk port radio.
- Tanpa radio, pusingan percubaan juga lebih pantas.

### Bar atas, Utama, tab Kenal pasti
- Butang itu bernama **«🔎 Kenal pasti Radio»** dan berada di kanan, di sebelah «Kemas kini».
- Dengan banyak tab terbuka, butang mengecil sendiri, hingga tinggal kanta pembesar sahaja.
- Utama: penerangan kotak daripada 4.2 kembali. Di bawah, hak cipta, laman web dan kod sumber kekal kelihatan, walaupun pada 1280×720.
- Tab Kenal pasti: penerangan ringkas dan jelas.

### Perisian tegar Retevis (43 model dengan kemas kini perisian tegar)
- **Retevis RA79 V3** (dan UV-K5 V3 / UV-K1, pemuat but 7.x) dengan fail .bin Retevis, contohnya «RA79 V3.bin». Perisian tegar hanya dihantar jika pemuat but radio daripada keluarga yang sama.
- **Retevis RA89** dengan fail .bin Retevis, contohnya «Eng_RA89_Upadat_20250524_B91.38.bin». Ia menggunakan protokol program Retevis «RA89-update-software», yang dijalankan dengan radio simulasi. Arahan yang dihantar sama bait demi bait.

### Lain-lain
- Ujian dengan radio simulasi tidak lagi menghalang ujian daripada dikompil pada Windows.

Galeri: 386 model daripada 27 jenama (348 dengan saluran, 302 dengan tetapan, 43 dengan kemas kini perisian tegar).

## Baharu dalam 4.3.0

### Pengenalpastian automatik: lebih pantas, lebih banyak radio, lebih banyak port
- Satu pusingan penuh percubaan tanpa radio mengambil kira-kira **50 saat** (sebelum ini 135).
- Program **sentiasa mendengar semua port kabel**, setiap satu secara berasingan dan semuanya serentak:
  - ia juga mengenali radio yang disambungkan dalam keadaan mati dan dihidupkan kemudian;
  - ia mengenali **beberapa radio pada kabel berbeza** pada masa yang sama, menunjukkan datanya dan membuka semua tabnya.
- **Radio yang telah dikenali** (cth. K67 Pro, Bajeton 8300, BJ-9000) dicuba dahulu, kemudian yang lain.
- Jika tab radio sudah terbuka, notis «Radio sudah dikenal pasti!! Tab radio sudah terbuka» muncul (dengan beberapa radio «Radio sudah dikenal pasti!! Tab radio sudah terbuka»).
- Tab Kenal pasti: port mengikut urutan (COM6, COM7… COM10) dan ditulis seperti «COM6 — USB-SERIAL CH340»; «Kenal pasti Radio Baharu Secara Automatik» dalam huruf tebal.

### Kemas kini perisian tegar: 8 radio lagi (41 kesemuanya)
- **Radtel RT-470, RT-470L, RT-470X, RT-495, RT-630** dengan fail .kdhx Radtel (pemuat but «JJCC Bootloader Pro»). RT-470: perisian tegar 1.x hanya untuk papan lama, 2.x hanya untuk yang baharu. Program menyatakannya dan meminta pengesahan bertulis.
- **TIDRADIO TD-H3, TD-H8 dan Radtel RT-730** dengan fail .BIN pengeluar: mulakan kemas kini dan hidupkan radio sambil menekan PTT.

### Lain-lain
- **Bajeton BJ-9000**: seperti BJ-8300, kesemua 960 saluran dalam satu senarai tanpa zon («Saluran: semua (960)»).
- **Utama** lebih padat: di bawah, versi, hak cipta, laman web dan kod sumber sentiasa kelihatan.
- **Galeri**: garis tebal memisahkan model yang serasi daripada yang tidak serasi.
- **LEGGIMI**: dibuang satu bahagian panjang yang berulang secara tidak sengaja dalam 4.2.0.

Galeri: 386 model daripada 27 jenama (348 dengan saluran, 302 dengan tetapan, 41 dengan kemas kini perisian tegar).

## Baharu dalam 4.2.0

- Pengenalpastian automatik dibetulkan: kotak «Pengenalpastian Radio Sedang Berjalan» hanya muncul apabila radio yang disambungkan menjawab (tidak lagi semasa dimulakan dengan kabel sahaja) dan hilang apabila radio dikenali atau tidak ditemui. Selepas dikenali, data dan tab radio dibuka dan carian berhenti, tanpa mematikan ciri itu.
- Butang baharu «🔎 Kenal pasti Radio Baharu Secara Automatik» dalam bar atas, antara «Maklumat» dan «Kemas kini»: ia mencari radio baharu dengan peraturan yang sama dan berwarna kelabu jika pengenalpastian automatik dimatikan.
- 35 radio baharu (36 versi) dan 8 jenama lagi: keluarga TYT TH-UV88 (TH-UV88, TH-UV98, Retevis RT85, RA89, P2, P62, Explorer QRZ-1), keluarga Radtel RT-490 (dengan JC-8629, FB-8629, Jianpai 8800 Plus, Boristone 8RS, Abbree AR-869, HamGeek HG-590), 19 radio keluarga Retevis RT21 (dengan Abbree AR-63) dan Quansheng TK11.
- Kemas kini perisian tegar juga untuk Radtel RT-590 / RT-590 Plus dan RT-600 / RT-600 Plus dengan fail .bin Radtel.
- Dibetulkan beberapa kecacatan tersembunyi dalam membaca nama saluran dengan huruf beraksen atau Cina.
- Galeri 386 model daripada 27 jenama: 348 dengan saluran dan pengulang, 302 dengan tetapan, 33 dengan kemas kini perisian tegar.

## Baharu dalam 4.1.0

- Pengenalpastian automatik dibetulkan: ia berfungsi dalam mana-mana tab dan, apabila anda menanggalkan radio, mematikannya atau menukarnya, carian bermula semula sendiri dan mengenali radio yang disambungkan selepas itu (walaupun radio yang sama atau dihidupkan selepas kabel). Sebelum ini, selepas pengenalan pertama, ia tidak lagi mencari radio lain.
- Semasa carian automatik, kotak «Pengenalpastian Radio Sedang Berjalan. Sila tunggu.» muncul, dengan butang «Henti».
- 14 radio baharu dan 3 jenama lagi: Radtel RT-470, RT-470L, RT-470X, RT-470X Plus, RT-630, RT-495 dan RT-730, Hiroyasu HI-8811, Talkpod A36plus dan A36plus 8 W, Anysecu UV-A37, Abbree AR-730, TIDRADIO TD-H8 (generasi ke-2 dan ke-3) dan TD-H3, dalam versi asas, HAM dan GMRS.
- Kemas kini perisian tegar Quansheng UV-K5, UV-K5(8), UV-K6 dan UV-5R Plus dengan fail .bin Quansheng, dengan semakan bahawa perisian tegar itu untuk jenis radio yang disambungkan.
- Skrin Utama dan laman web: «PENGURUSAN SALURAN - FREKUENSI - PENGULANG - PERISIAN TEGAR» dan nama semua jenama dalam bingkai gaya permainan video tahun 80-an.
- Galeri 349 model daripada 19 jenama: 311 dengan saluran dan pengulang, 265 dengan tetapan, 31 dengan kemas kini perisian tegar.

## Baharu dalam 4.0.0

- Program pelbagai jenama: 335 model daripada 16 jenama dalam galeri, 297 dengan saluran dan frekuensi melalui kabel dan 251 dengan tetapan radio.
- Radio baharu: Quansheng UV-K5, UV-K6, UV-5R Plus dan UV-K1; 31 model Retevis (H777, RT22, RT22S, RB18, RB19, RB29, RT47, RT68, RB85, RA79, RA685…); Radtel RT-900, RT-900 BT, RT-920 dan T18; Radioddity GM-30, GM-30 Plus, GM-30 Pro, GS-10B, GA-2S dan GA-510; TIDRADIO TD-H6, MaxTalker, Abbree, Intek, Maverick, BTECH, WLN, Zastone, Luiton dan TID.
- Pengenalpastian automatik (tab «Kenal pasti»): sambungkan kabel dan radio dikenali sendiri, dengan tabnya terbuka dan saluran dibaca. Pilihan diingati dan dimatikan dengan butang yang sama.
- Galeri model dengan menu jenama dan satu warna bagi setiap jenama.
- «Baca dari radio» dan «Tulis ke radio» berwarna biru dalam tab radio.
- Utama, Maklumat, panduan, laman web dan dokumentasi dikemas kini untuk program pelbagai jenama.
