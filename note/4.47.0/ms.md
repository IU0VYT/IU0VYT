# IU0VYT Flash 4.47.0

<div align="center">
<table>
<tr>
<td align="center"><a href="https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0"><img src="../../docs/img/flags/en.svg" width="36" height="24" alt=""><br><sub>English</sub></a></td>
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

Halaman keluaran di GitHub (dalam bahasa Inggeris), dengan muat turun: [V.4.47.0](https://github.com/IU0VYT/IU0VYT/releases/tag/V.4.47.0)

IU0VYT Flash V.4.47.0 untuk Microsoft Windows 10/11

### Baharu
- **Retevis, 24 model baharu yang diperoleh daripada program pengeluar**: B3B, B63B, RT22P, RT622P, RB666, RB666P (memori saluran yang sama dengan versi baharu H777); RB45, RB645, RB47; B3H, B63H, RB647, RT68H, RT668H, NR10, NR610, RB66S, RB666S, RT68B, RT668B, RB669 (keluarga T18); RB65, RB66, RB66P (satu frekuensi dan satu nada bagi setiap saluran). Sambungan, pembacaan dan penulisan ditulis semula dari awal dan disemak dengan fail data kilang program-program itu. RB66 dan RB66P, serta RB669, RT68H dan RT668H, menjawab dengan kod yang sama: kad yang betul mesti dipilih sendiri.
- **Tetapan bagi 24 model Retevis**: 241 item kesemuanya (squelch, had masa pemancaran, VOX, imbasan, panduan suara, penjimatan bateri, bip, kekunci sisi…). Hanya item yang dipaparkan oleh program model itu ditawarkan; nilai dalam 24 fail kilang semuanya pilihan yang dibenarkan.
- **Kata laluan pengaturcaraan**: pada 24 model Retevis dan pada versi baharu Retevis H777 ia tidak menghalang pembacaan atau penulisan, seperti dalam CHIRP (program pengeluar pula berhenti jika kata laluan bukan yang ditaip oleh pengguna). Ia ialah satu item tetapan, dipaparkan dengan jelas: ia boleh dilihat, ditukar dan dibuang. Panjang dan aksara adalah seperti program pengeluar: 6 aksara (3 bagi RB45, RB645, RB47 dan RB66; 4 bagi RB66P), digit dan huruf (digit sahaja bagi RB65 dan RB66). Bagi B3H, B63H, RB45, RB645, RB66 dan RB666, yang programnya tidak memaparkan kotak kata laluan, satu amaran mengingatkannya.
- **Tetapan untuk radio yang sebelum ini hanya mempunyai saluran**: keluarga AnyTone (AT-D868UV, AT-D878UV, AT-D878UV II Plus, BTECH DMR-6X2 dan DMR-6X2 PRO: 27 item, daripada huraian memori qdmr dan dmrconfig); TYT MD-UV380, MD-UV390, MD-2017 dan MD-9600, Retevis RT3S dan RT82 (item Baofeng DM-1701); Radtel RT-880 dan RT-950 Pro dengan perisian tegar nicFW880 / nicFW950 (17 item, daripada dokumentasi pengarangnya); TID TD-UV68 (item TD-H8); Radtel RT-890 dengan perisian tegar M7OCM (53 item).
- **Versi percubaan**: salinan dengan nombor yang lebih tinggi daripada versi stabil terkini yang diterbitkan memaparkan bar amaran («Versi percubaan (ujian beta)…») dan, hanya salinan itu, menawarkan versi percubaan yang berikutnya, yang sentiasa dimuat turun dari GitHub. Pengguna versi stabil tidak melihat versi percubaan.
- Galeri: 697 model daripada 54 jenama (647 dengan saluran, 596 dengan tetapan).

### Nota
- Tiada satu pun daripada 24 radio Retevis diuji pada peranti sebenar, sama ada saluran atau tetapan: baca sahaja dahulu, kemudian satu tetapan yang tidak berbahaya. Jika radio menjawab secara tidak dijangka, pembacaan berhenti dengan mesej dan tiada apa-apa ditulis ke radio.
- Dengan kata laluan pengaturcaraan yang sudah ada, program tetap cuba membaca dan menulis, seperti CHIRP: ini belum diuji pada radio sebenar.
- RB66 dan RB66P, serta RB669, RT68H dan RT668H, menjawab dengan kod yang sama: kad yang betul mesti dipilih sendiri.
- Program bebas, tidak bergabung dengan pengeluar radio.
- Rangkaian ujian penuh lulus: bahagian umum; 915 perbandingan struktur dan 380 bacaan / penulisan yang sama dengan CHIRP.
