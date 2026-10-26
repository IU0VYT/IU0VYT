// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE del programma del costruttore «BF-Ts» della
// BF-T9 (classe KDH.BF.BF480.WriFreq, caricata direttamente da BF-Tseries.exe,
// senza decompilarla) collegato alla radio simulata.
//   t9_harness <eseguibile> <host:porta> read
//   t9_harness <eseguibile> <host:porta> write <modifiche>
// read: stampa i canali letti, uno per riga: n|RX|TX|toneRX|toneTX|potenza|banda|scan|occupato|beat
// (valori del programma: potenza 0 bassa / 1 alta, banda 0 stretta / 1 larga,
// scan 1 = aggiunto); write: legge la radio, applica le modifiche (stesso
// formato; «n|» = canale vuoto) e scrive come il pulsante «Scrivi» del programma.
using System;
using System.IO;
using System.IO.Ports;
using System.Reflection;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;
	static Assembly asm;

	static object F(object o, string name) => o.GetType().GetField(name, All).GetValue(o);
	static object P(object o, string name) => o.GetType().GetProperty(name, All).GetValue(o);
	static void SetP(object o, string name, object v) => o.GetType().GetProperty(name, All).SetValue(o, v);

	static object Wri(SerialPort sp, object data, string op)
	{
		var wT = asm.GetType("KDH.BF.BF480.WriFreq", true);
		var oT = asm.GetType("KDH.BF.BF480.OPERATION_TYPE", true);
		return wT.GetConstructors(All)[0].Invoke(new object[] { sp, data, "English", Enum.Parse(oT, op) });
	}

	static bool Run(object w) => (bool)w.GetType().GetMethod("DoIt", All).Invoke(w, null);

	static SerialPort Open()
	{
		var sp = (SerialPort)Activator.CreateInstance(asm.GetType("KDH.BF.BF480.MySerialPort", true), true);
		sp.PortName = "COM9";
		sp.BaudRate = 9600;
		sp.DataBits = 8;
		sp.StopBits = StopBits.One;
		sp.Parity = Parity.None;
		sp.Open();
		return sp;
	}

	static int Main(string[] a)
	{
		SerialPort.Endpoint = a[1];
		asm = Assembly.LoadFrom(a[0]);
		var dataT = asm.GetType("KDH.BF.BF480.ApplicationDataFormat", true);
		// lettura (FormMain: WriFreq(sP, theApplicationData, lang, READ), poi i dati in BufForAppData)
		var rw = Wri(Open(), Activator.CreateInstance(dataT), "READ");
		if (!Run(rw))
		{
			Console.Error.WriteLine("lettura non riuscita");
			return 2;
		}
		var data = F(rw, "BufForAppData");
		var chs = (Array)F(data, "CHInfoData");
		if (a[2] == "write")
		{
			System.Threading.Thread.Sleep(300);
			foreach (var line in File.ReadAllLines(a[3]))
			{
				if (line.Trim() == "") continue;
				var f = line.Split('|');
				var c = chs.GetValue(int.Parse(f[0]));
				if (f.Length < 3)
				{
					SetP(c, "RxFreq", ""); SetP(c, "TxFreq", "");
					continue;
				}
				SetP(c, "RxFreq", f[1]); SetP(c, "TxFreq", f[2]);
				SetP(c, "CTS_DCS_Decoder", f[3]); SetP(c, "CTS_DCS_Encoder", f[4]);
				SetP(c, "TxPower", int.Parse(f[5])); SetP(c, "BandWidth", int.Parse(f[6]));
				SetP(c, "ScanAdd", int.Parse(f[7])); SetP(c, "BusyLock", int.Parse(f[8]));
				SetP(c, "BeatShift", int.Parse(f[9]));
			}
			var ww = Wri(Open(), data, "WRITE");
			if (!Run(ww))
			{
				Console.Error.WriteLine("scrittura non riuscita");
				return 3;
			}
		}
		else
		{
			for (int i = 0; i < chs.Length; i++)
			{
				var c = chs.GetValue(i);
				Console.WriteLine(string.Join("|", i, P(c, "RxFreq"), P(c, "TxFreq"), P(c, "CTS_DCS_Decoder"), P(c, "CTS_DCS_Encoder"),
					P(c, "TxPower"), P(c, "BandWidth"), P(c, "ScanAdd"), P(c, "BusyLock"), P(c, "BeatShift")));
			}
		}
		System.Threading.Thread.Sleep(300);
		SerialPort.Shutdown();
		Environment.Exit(0);
		return 0;
	}
}
