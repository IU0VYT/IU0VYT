# IU0VYT Flash 4.9.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Halaman keluaran di GitHub (dalam bahasa Inggeris), dengan muat turun: [V.4.9.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.9.0)

IU0VYT Flash V.4.9.0 untuk Microsoft Windows 10/11

Versi ini juga merangkumi perkara baharu 4.8.0, 4.7.0 dan 4.6.0, yang diterbitkan hanya sebagai versi ujian.

## Baharu dalam 4.9.0

### Dibetulkan: Kenal pasti dengan Lanchonlh HG-UV98 dan Quansheng TG-UV2+
- Semasa «Kenal pasti», port sentiasa berkomunikasi dalam 8N1. Kini ia menggunakan pariti genap (HG-UV98) dan dua bit henti (TG-UV2+), seperti bacaan dari tab radio. Kedua-dua radio kini dikenali.

### Dibetulkan: pariti HG-UV98
- Dengan pariti genap, port Windows kini turut menyemak ralat pariti. Sebelum ini semakan dihidupkan lalu terus dimatikan.

### Quansheng UV-K5 / UV-K6 / UV-5R Plus dengan perisian tegar bebas
- Tab baharu untuk perisian tegar egzumer, F4HWN dan LOSEHU/LTS/UVE.
- F4HWN menggunakan pemacu F4HWN oleh armel (di luar CHIRP): 8 tahap kuasa, kunci pemancaran, 3 senarai imbasan dan tetapan perisian tegar.
- Setiap tab membaca dan menulis perisian tegarnya sendiri sahaja. Selepas kemas kini perisian tegar, program menunjukkan tab yang perlu digunakan.

### Radio baharu
- Wouxun (20 model): KG-UVD1P (KG-UV2D, KG-UV3D), KG-UV6D, KG-816, KG-818, KG-805G, KG-UV8D, KG-UV8D Plus, KG-UV8E, KG-UV9D Plus, KG-UV9PX, KG-UV9GX, KG-UV9K, KG-UV9G Pro, KG-935G, KG-935G Plus, KG-935H, KG-UV8H, KG-XS20G, KG-XS20G Plus, KG-XS20H
- TYT TH-UV8000D/E, UVL-15W, TH-350 (AS)
- Abbree AR-518, Ruyage UV58Plus, Radtel RT-880G, Radtel RT-890 (perisian tegar Radtel dan perisian tegar bebas M7OCM)
- Retevis C2, HA1G, HA1UV, HA2, RB15, RB615, RB17P, RB28, RB628, RB48, RB648, RT23, RT26, RT76P, RT87, RT16, RB27B, RB27V
- BTECH GMRS-V1, GMRS-V2, MURS-V1, MURS-V2, FRS-A1
- AnyTone TERMN-8R, OBLTR-8R
- Puxing PX-888K, QYT KT-8R, Rugged RH5R-V2, Baojie BJ-UV55, MTC UV-5R-3, Radioddity MU-5, Zartek ZA-758, TIDRADIO TD-M11, TDXone TD-Q8A

Bagi radio daripada pemacu CHIRP, saluran dan tetapan disemak dengan ujian automatik berbanding CHIRP. Radio GMRS, FRS dan MURS adalah untuk Amerika Syarikat: di Itali pemancaran pada frekuensi itu tidak dibenarkan.

## Baharu dalam 4.8.0

### Kemas kini perisian tegar .BTU (Radtel RT-910 dan RT-900 tanpa Bluetooth, Bajeton BJ-7700 tanpa Bluetooth)
- Perisian tegar .BTU dikemas kini dengan protokol program «Radtel Bootload Tools». Pemuat but adalah sama dengan Baofeng UV-5RM, dengan salam pembukaan «PROGRAMBT80U».
- Mod automatik, dengan radio dihidupkan seperti biasa. Mod manual: radio dimatikan, dihidupkan sambil menekan kedua-dua kekunci sisi, sehingga paparan menunjukkan «UPDATE».
- Pemuat but tidak menyemak model. Oleh itu program mengenal pasti radio dan versi perkakasan daripada nama fail, menunjukkan amaran dan sentiasa meminta anda menaip model.
- Dalam galeri, RT-910 (tanpa Bluetooth, .BTU) dan RT-910B (dengan Bluetooth, .anc) kini mempunyai dua kad. Begitu juga BJ-7700 dan BJ-7700 tanpa Bluetooth.

### Dibetulkan: Radtel RT-900 tanpa Bluetooth dengan perisian tegar 512 saluran
- Hingga V1.09P radio mempunyai 512 saluran; dari V1.14P ia mempunyai 999.
- Kini saluran 513-999 hanya ditulis jika anda mengubahnya. Terlebih dahulu program menyemak dalam radio bahawa kawasan itu kosong atau berisi saluran.

### Radio baharu (pemacu CHIRP)
- TYT TH-UV3R, TH-UV3R-25, TH-UVF1, TH-UVF8D
- KYD NC-630A, IP-620
- KSUN M6
- Puxing PX-777, PX-2R
- Boblov X3 Plus
- Lanchonlh HG-UV98
- Radioddity R2
- Retevis RT24, RT24V, H777S, RT1, H777 V4 (versi baharu H777), RT21H
- Quansheng TG-UV2+
- Feidaxin FD-268A, FD-268B, FD-288A, FD-288B, FD-150A, FD-160A, FD-450A, FD-460A, FD-460UH

Bagi semuanya, saluran dan tetapan disemak dengan ujian automatik berbanding CHIRP.

## Baharu dalam 4.7.0

### Dibetulkan: pengenalpastian automatik tersekat atau sangat perlahan
- Radio yang dihidupkan dalam mod kemas kini perisian tegar menghantar mesej berkala dengan sendirinya: sebelum ini ia seolah-olah menjawab setiap model dan carian mengambil beberapa minit. Kini program mendengar port sebelum mencuba dan tidak tertipu: carian selesai dalam masa kurang seminit dan tab Kenal pasti menerangkan apa yang berlaku.
- Antara muka tidak lagi tergantung semasa carian bermula; radio yang paling biasa dicuba dahulu.
- Dibetulkan notis palsu «Tab radio sudah terbuka» bagi tab yang telah ditutup sebelumnya.

### Radio yang diurus dengan cara yang sama
- Tetingkap pilihan dan tab Kenal pasti menyenaraikan semua radio yang mungkin (cth. Radtel RT-950 Pro, Bajeton BJ-9000, Hiroyasu HI-UV850 Pro), dengan menu untuk memilih dan petunjuk ringkas tentang tempat melihat versi perisian tegar pada radio.

### Dibetulkan: Radtel RT-950 Pro / Bajeton BJ-9000 dengan perisian tegar hingga 0.23
- Bit «pemancaran dibenarkan» bagi saluran kini sentiasa ditulis, dengan kedua-dua versi (seperti program Radtel hingga v1.2.x; v1.3.3 mengabaikannya). Dengan perisian tegar lama, saluran yang diubah dalam tab 990 saluran mungkin tidak memancar. Bit itu juga dipulihkan dalam saluran 1-960 yang tidak memilikinya.
- Versi sebenar Bajeton BJ-9000 disahkan: perisian tegar V0.22, V0.27 dan V0.29 (sama dengan Radtel), program v1.1.1 (960 saluran) dan v1.3.3 (990 saluran, iaitu program Radtel). Saluran dan tetapan turut disemak dengan kod program Bajeton.

### Kemas kini perisian tegar untuk radio lain
- Radtel RT-950 Pro, Radtel RT-950 dan Hiroyasu HI-UV850 Pro (fail .BTF, pemuat but sama dengan BJ-9000).
- Radtel RT-920 dan RT-910B (fail .anc, protokol sama dengan BJ-7800 dan BJ-7700).

### Radio baharu (pemacu CHIRP)
- Radtel RT-620, Radioddity GA-510 V2, TYT TH-350, Retevis RT18, Retevis RB627B (PMR446); dalam galeri juga Hiroyasu HI-UV850 Pro.
- Tetapan juga untuk SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5 dan Anysecu AC-580.

## Baharu dalam 4.6.0

### Dibetulkan: Radtel RT-950 Pro / Bajeton BJ-9000 dengan perisian tegar terkini
- Sejak perisian tegar **0.24** (Disember 2025) radio mempunyai **990 saluran dalam 10 zon 99** (atau satu senarai 990); hingga **0.23** ia mempunyai **960 dalam 15 zon 64** (atau satu senarai 960). Program kini mengurus kedua-dua versi.
- Versi dikenal pasti daripada data yang dibaca dari radio (radio tidak memaklumkannya): bacaan membuka saluran dalam tab yang betul. Jika data tidak mencukupi, satu nota menerangkan cara menyemak versi pada radio.
- Saluran 961-990 hanya ditulis jika berubah, selepas dibaca semula daripada radio; jika radio tidak memilikinya (perisian tegar hingga 0.23), tiada apa-apa ditulis.
- Saluran dan tetapan disemak dengan kod asal program Radtel v1.2.0 (960 saluran) dan v1.3.3 (990 saluran).

### Baharu: Radtel RT-950 (tanpa «Pro»)
- 960 saluran (15 zon 64) atau satu senarai, tetapan umum; disemak dengan kod asal program «BT-9000_CPS» v1.3.0.

### Dibetulkan: nama tepat radio
- Radio berbeza yang diurus dengan cara yang sama (cth. **Radtel RT-950 Pro** dan **Bajeton BJ-9000**, **Baofeng K67 Pro** dan **5RH Pro**, Radtel RT-910 dan Bajeton BJ-7700) kini dipanggil dengan nama sendiri dalam tab, bacaan, penulisan, pengenalpastian dan fail salinan.
- Pada pengenalpastian pertama salah satu radio ini, program bertanya radio yang mana lalu mengingatinya.

### Dibetulkan: amaran data belum disimpan
- Selepas bacaan, penulisan atau pemajuan, saluran berada dalam radio: semasa menutup program, amaran hanya muncul jika ada perubahan sebenar.

### Radio baharu (pemacu CHIRP)
- SenhaiX 8800, Radioddity GS-5B, Cignus XTR-5, Anysecu AC-580, TID TD-UV68.
- Radtel RT-920: 990 saluran dalam 10 zon 99, seperti pada radio.

Galeri: 489 model daripada 44 jenama (451 dengan saluran, 386 dengan tetapan, 52 dengan kemas kini perisian tegar). Jenama baharu: Wouxun, AnyTone, QYT, Rugged, Baojie, MTC, Zartek, Ruyage, TDXone.
