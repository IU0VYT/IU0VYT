// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.
// Parti derivate da CHIRP (https://chirpmyradio.com, GNU GPL), driver:
// retevis_ha1g.py, retevis_ha2.py (4.9).
// Modifiche: riscritte in Go e adattate a IU0VYT Flash da Daniele "IU0VYT" Bonifazi,
// novembre 2026 (elenco delle modifiche per versione in CHANGELOG.txt).

package main

// 4.9: mappe della memoria di retevis_ha1g.py e retevis_ha2.py di CHIRP (copiate così come sono)

const ha1gDefs = `
struct radioinfo {
    u8 modelnumber[32];
    u8 hardwareversion[2];
    u8 serialno[16];
    ul32 freqmin;
    ul32 freqmax;
    u8 rev_1;
    u8 saleterr;
    u8 radiomode;
    u8 rev[7];
};

struct dataver {
    u8 dataver[2];
    u8 softver[4];
    u8 rev[4];
};

struct zoneinfo {
    ul16 zonenum;
    ul16 zoneindex[16];
    struct {
        char name[14];
        ul16 chnum;
        ul16 chindex[64];
    } zones[16];
};

struct scaninfo {
    ul16 scannum;
    ul16 scanindex[16];
    struct  {
       char name[14];
       u8 scantxch:4,
          scancondition:4;
       u8 hangtime:4,
          talkback:1,
          scanstatus:3;
       u8 chnum;
       ul16 specifych;
       ul16 PriorityCh1;
       ul16 PriorityCh2;
       u8 scanmode;
       ul16 chindex[100];
    } scans[16];
};

struct vfoscaninfo {
    ul16 vfoscannum;
    ul16 vfoscanindex[3];
    struct {
       u8 scantxch:4,
          scancondition:4;
       u8 hangtime:4,
          talkback:1,
          startcondition:1,
          rev_1:2;
       u8 scanmode;
       ul32 vhffreq_start;
       ul32 vhffreq_end;
       ul32 uhffreq_start;
       ul32 uhffreq_end;
       u8 rev;
    } vfoscans[3];
};

struct alarminfo {
    ul16 alarmnum;
    ul16 alarmindex[8];
    struct {
        char name[14];
        u8 alarmtype:4,
           alarmmode:4;
        ul16 jumpch;
        u8 localalarm:1,
           txbackground:1,
           ctcmode:2,
           rev_1:4;
        u8 alarmtime:4,
           alarmcycle:4;
        u8 mictime:4,
           txinterval:4;
        u8 alarmid[8];
        u8 alarmstatus;
        u8 rev_3[1];
    } alarms[8];
};

struct dtmfinfo {
    ul16 dtmfnum;
    ul16 dtmfindex[4];
    struct {
        u8 autoresettime:4,
           codedelaytime:4;
        u8 stunmode:2,
           showani:1,
           sidetone:1,
           pttidtype:2,
           rev_1:2;
        char callid[10];
        char stunid[10];
        char revive[10];
        char bot[16];
        char eot[16];
        char rev_2[8];
    } dtmfcomm;
    struct {
        char name[14];
        u8 codelen:4,
           signaling:4;
        u8 groupcode:4,
           intermediatecode:4;
        char fastcall1[16];
        char fastcall2[16];
        char fastcall3[16];
        char fastcall4[16];
        char fastcall5[16];
        char fastcall6[16];
        char fastcall7[16];
        char fastcall8[16];
        char fastcall9[16];
        char fastcall10[16];
        u8 rev_1:6,
           EncodingEnable:1,
           DecodingEnable:1;
        u8 dtmfstatus;
        u8 rev_2[12];
    } dtmfs[4];
};
`

const ha1gFormat = `

#seekto 0x0E;
struct radioinfo radioinfo;

#seekto 0x52;
struct dataver dataver;

#seekto 0x5C;
struct {
  u8 readpwd[8];
  u8 writepwd[8];
  u8 poweronpwd[8];
  u8 squelch:4,
     rev_1:4;
  u8 powersavingdelaytime:4,
     powersavingmode:4;
  u8 tk1_short;
  u8 tk1_long;
  u8 sk1_short;
  u8 sk1_long;
  u8 sk2_short;
  u8 sk2_long;
  u8 longtime;
  u8 chknoblock:1,
     sidekeylock:1,
     panellock:1,
     rev_2:3,
     autoormanuallock:1,
     keylockswitch:1;
  u8 wxscan:1,
     agc:1,
     rtcreport:1,
     displaybatterylevel:1,
     batterlevelreport:1,
     lowbattery:1,
     stunmode:2;
  u8 voiceprompts:1,
     touchtone:1,
     txpermittone:1,
     channelmode:1,
     screenlight:1,
     rogerbeep:1,
     rev_3:2;
  ul16 poweron_type_1:4,
       homepoweronzone_1:12;
  ul16 homepoweronch_1;
  u8 relaytaildelay;
  u8 radioid[10];
  u8 radioname[16];
  u8 micmain;
  u8 agcmain:4,
     agcthreshold:4;
  u8 scenemode;
  u8 autooff;
  u8 menuouttime;
  u8 voxthreshold:4,
     voxdelaytime:4;
  u8 rev_4:1,
     homeselect:3,
     homeindex:2,
     chdisplay:2;
  ul16 poweron_type_2:4,
     homepoweronzone_2:12;
  ul16 homepoweronch_2;
  ul16 homepoweronzone_3;
  ul16 homepoweronch_3;
  ul16 freqstep;
  u8 language;
  u8 backlightbrightness:4,
     backlighttime:4;
  u8 wxch;
  u8 calltone;
  u8 rev_5:5,
     homechtype_3:1,
     homechtype_2:1,
     homechtype_1:1;
  u8 spreadspectrummode;
  u8 salezone;
  u8 tailsoundeliminationsfre:7,
     tailsoundeliminationswitch:1;
  u8 singletone;
  u8 scanlist;
  u8 rev_6[6];
} settings;

#seekto 0xc0;
struct zoneinfo zoneinfo;

#seekto 0x0D42;
struct  {
    ul16 chnum;
    ul16 chindex[1027];
} channeldata;

#seekto 0x154a;
struct  {
  char alias[14];
  u8 chmode:4,
     chpro:4;
  u8 fixedpower:1,
     fixedbandwidth:1,
     tailelimination:2,
     power:2,
     bandwidth:2;
  ul32 rxfreq;
  ul32 txfreq;
  u16 rxctcvaluetype:2,
      rxctctypecode:1,
      rev_1:1,
      rxctc:12;
  u8 rxsqlmode;
  u16 txctcvaluetype:2,
      txctctypecode:1,
      rev_2:1,
      txctc:12;
  u8 totpermissions:2,
     tottime:6;
  u8 vox:1,
     companding:1,
     scramble:1,
     offlineorreversal:2,
     rev_3:3;
  u8 voxthreshold:4,
     voxdelaytime:4;
  u8 autoscan:1,
     scanlist:7;
  u8 alarmlist;
  u8 rev_4:2,
     pttidtype:2,
     dtmfsignalinglist:4;
  ul16 dtmfcallid;
  u8 reserve[3];
} channels[1027];

#seekto 0xb5c2;
struct scaninfo scans;

#seekto 0xc3e4;
struct vfoscaninfo vfoscans;

#seekto 0xc444;
struct alarminfo alarms;

#seekto 0xc848;
struct dtmfinfo dtmfinfos;

`

const ha2Format = `
#seekto 0x0E;
struct radioinfo radioinfo;
#seekto 0x52;
struct dataver dataver;
#seekto 0x5C;
struct {
  u8 readpwd[8];
  u8 writepwd[8];
  u8 poweronpwd[8];
  u8 squelch;
  u8 powersavingdelaytime:4,
     powersavingmode:4;
  u8 tk1_short;
  u8 tk1_long;
  u8 sk1_short;
  u8 sk1_long;
  u8 sk2_short;
  u8 sk2_long;
  u8 rev_2;
  u8 chknoblock:1,
     sidekeylock:1,
     panellock:1,
     rev_3:3,
     autoormanuallock:1,
     keylockswitch:1;
  u8 rev_4:5,
     lowbattery:1,
     stunmode:2;
  u8 voiceprompts:1,
     touchtone:1,
     txpermittone:1,
     rev_5:1,
     screenlight:1,
     DualWatch:1,
     rogerbeep:2;
  ul16 poweron_type_1:4,
       homepoweronzone_1:12;
  ul16 homepoweronch_1;
  u8 rev_6[27];
  u8 micgain;
  u8 rev_7[2];
  u8 autooff;
  u8 menuouttime;
  u8 voxthreshold:4,
     voxdelaytime:4;
  u8 rev_8:1,
     homeselect:3,
     homeindex:2,
     chdisplay:2;
  ul16 poweron_type_2:4,
     homepoweronzone_2:12;
  ul16 homepoweronch_2;
  u8 rev_9[4];
  u8 freqstep;
  u8 rev_10[2];
  u8 backlightbrightness:4,
     backlighttime:4;
  u8 wxch;
  u8 calltone;
  u8 rev_11:5,
     homechtype_3:1,
     homechtype_2:1,
     homechtype_1:1;
  u8 spreadspectrummode;
  u8 salezone;
  u8 tailsoundeliminationsfre:7,
     tailsoundeliminationswitch:1;
  u8 singletone;
  u8 rev_12:3,
     btswitch:1,
     rev_13:1,
     aviationfreqrx:1,
     fmchtype:1,
     aliasdisplay:1;
  u8 btinmic:1,
     btinspk:1,
     rev_14:6;
  u8 scanlist;
  u8 rev_15[2];
  u16 radioid;
  u8 rev_16[12];
  u8 pttidaliastxmode;
  u8 batterydisplaymode;
  u8 rev_17[28];
  u8 fmchannel;
  u8 rev_18;
  u8 aliasnum;
  u8 timezone;
  u8 rev_19[6];
} settings;

#seekto 0xf4;
struct zoneinfo zoneinfo;

struct channel {
    char alias[14];
    u8 chmode:4,
       chpro:4;
    u8 fixedpower:1,
       fixedbandwidth:1,
       rev_20:2,
       power:2,
       bandwidth:2;
    ul32 rxfreq;
    ul32 txfreq;
    u16 rxctcvaluetype:2,
        rxctctypecode:1,
        rev_1:1,
        rxctc:12;
    u8 rxsqlmode;
    u16 txctcvaluetype:2,
        txctctypecode:1,
        rev_2:1,
        txctc:12;
    u8 totpermissions:2,
       tottime:6;
    u8 vox:1,
       offlineorreversal:2,
       companding:2,
       scramble:1,
       rev_3:2;
    u8 scrambleFreq;
    u8 freqdiff;
    u8 autoscan:1,
       scanlist:7;
    u8 alarmlist;
    u8 rev_4:2,
       pttidtype:2,
       dtmfsignalinglist:4;
    ul32 difffreq;
};

#seekto 0x0a16;
struct {
    ul16 chnum;
    ul16 chindex[1024];
    struct channel channels[1024];
} channeldata;

#seekto 0xb218;
struct {
    ul16 chnum;
    ul16 chindex[3];
    struct channel vfochannels[3];
} vfochanneldata;

#seekto 0xb298;
struct scaninfo scans;

#seekto 0xc0ba;
struct vfoscaninfo vfoscans;

#seekto 0xc0fe;
struct alarminfo alarms;

#seekto 0xc200;
struct {
    ul16 fmnum;
    ul16 fmindex[64];
    struct {
        char name[14];
        ul32 freq;
        u8 bandwidth;
        u8 rev_1[3];
    } fms[64];
} fmdata;

#seekto 0xc802;
struct dtmfinfo dtmfinfos;

struct callsign {
   char callsign[8];
   u8 ssid;
};

#seekto 0xcc6e;
struct {
   u8 rev_1:4,
      anaptttype:2,
      aprsswitch:2;
   u8 smartbeaconing:2,
      rev_2:2,
      beaconmode:2,
      positionorgps:1,
      beaconreceivetone:1;
   u8 rev_3[3];
   u8 digi_path:4,
      aprsaudio:1,
      rev_4:3;
   struct callsign call;
   struct callsign digi_path_p1[2];
   struct callsign digi_path_p2[2];
   struct callsign digi_path_p3[2];
   struct callsign digi_path_p4[2];
   struct callsign digi_path_f1[8];
   struct callsign digi_path_f2[8];
   char fixsymbol[2];
   u8 rev_5[5];
   struct {
      char degrees[2];
      char minutes[2];
      char dot;
      char seconds[2];
      u8 direction;
   } latitude;
   struct {
      char degrees[3];
      char minutes[2];
      char dot;
      char seconds[2];
      u8 direction;
   } longitude;
   u8 rev_6;
   u8 tx_delay;
   u8 autointerval;
   u8 rev_7;
   u8 screenhold;
   u8 rev_8[4];
   u8 ana_ch;
   u8 rev_9[9];
   struct {
      u8 lowspeed;
      u8 highspeed;
      u8 slowrate;
      u8 fastrate;
      u8 turnangle;
      u8 turnslope;
      u8 turntime;
      u8 rev_10;
   } smartbeacons[4];
   u8 rev_11[64];
   struct {
      ul32 freq;
      u8 band;
   } ana_chs[8];
   u8 rev_12[188];
   char commenttext[64];
} aprsinfo;

#seekto 0x11812;
struct {
   u8 gnssswitch:1,
      pttuploadmode:2,
      rev_1:5;
   u8 rev_2[4];
   u8 txtype:4,
      autotxtime:4;
   char replaytext[60];
   u8 rev_3[10];
} gnssinfo;

#seekto 0x1185e;
struct {
    ul16 aliasnum;
    ul16 aliasindex[30];
   struct {
      char alias[8];
   } alias[30];
} aliaslist;
`
