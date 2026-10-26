// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE dei programmi del costruttore «T17-T18 PMR446
// PRO» e «T17-T18 FRS» (classe AnaCps.Data.DataProtocol, caricata direttamente
// dall'eseguibile, senza decompilarla) collegato alla radio simulata.
//   t17_harness <eseguibile> <host:porta> read
//   t17_harness <eseguibile> <host:porta> write <modifiche>
// read: stampa i canali letti, uno per riga: n|RX|TX|toneRX|toneTX|potenza|banda|scan|occupato|salto
// write: legge la radio, applica le modifiche (righe nello stesso formato; «n|» =
// canale vuoto) e scrive come il pulsante «Scrivi» del programma.
using System;
using System.IO;
using System.IO.Ports;
using System.Reflection;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;
	static Assembly asm;

	static object Proto(SerialPort sp, object data, string dir)
	{
		var pT = asm.GetType("AnaCps.Data.DataProtocol", true);
		var dT = asm.GetType("AnaCps.Data.OPT_DIR", true);
		return pT.GetConstructors(All)[0].Invoke(new object[] { sp, data, Enum.Parse(dT, dir) });
	}

	static string Run(object p)
	{
		var r = p.GetType().GetMethod("DoIt", All).Invoke(p, null);
		return r.ToString();
	}

	static object F(object o, string name) => o.GetType().GetField(name, All).GetValue(o);
	static object P(object o, string name) => o.GetType().GetProperty(name, All).GetValue(o);
	static void SetP(object o, string name, object v) => o.GetType().GetProperty(name, All).SetValue(o, v);

	static int Main(string[] a)
	{
		SerialPort.Endpoint = a[1];
		asm = Assembly.LoadFrom(a[0]);
		var sp = new SerialPort();
		sp.PortName = "COM9";
		sp.BaudRate = 9600;
		sp.DataBits = 8;
		sp.StopBits = StopBits.One;
		sp.Parity = Parity.None;
		sp.Open();
		var dataT = asm.GetType("AnaCps.Data.DataSetApp", true);
		// lettura (FormMain: DataProtocol(sp, AppData, DIR_READ), poi AppData = bufForData)
		var rp = Proto(sp, Activator.CreateInstance(dataT), "DIR_READ");
		var res = Run(rp);
		var data = F(rp, "bufForData");
		if (res != "SUCCESS")
		{
			Console.Error.WriteLine("lettura: " + res);
			return 2;
		}
		if (a[2] == "write")
		{
			System.Threading.Thread.Sleep(300);
			var chs = (Array)F(data, "dataChannel");
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
				SetP(c, "RxCtsDcs", f[3]); SetP(c, "TxCtsDcs", f[4]);
				SetP(c, "PowerMode", int.Parse(f[5])); SetP(c, "BandWide", int.Parse(f[6]));
				SetP(c, "ScanAdd", int.Parse(f[7])); SetP(c, "BusyLock", int.Parse(f[8]));
				SetP(c, "Jmpfreq", int.Parse(f[9]));
			}
			var wp = Proto(sp, data, "DIR_WRITE");
			res = Run(wp);
			if (res != "SUCCESS")
			{
				Console.Error.WriteLine("scrittura: " + res);
				return 3;
			}
		}
		else
		{
			var chs = (Array)F(data, "dataChannel");
			for (int i = 0; i < chs.Length; i++)
			{
				var c = chs.GetValue(i);
				Console.WriteLine(string.Join("|", i, P(c, "RxFreq"), P(c, "TxFreq"), P(c, "RxCtsDcs"), P(c, "TxCtsDcs"),
					P(c, "PowerMode"), P(c, "BandWide"), P(c, "ScanAdd"), P(c, "BusyLock"), P(c, "Jmpfreq")));
			}
		}
		System.Threading.Thread.Sleep(300);
		sp.Close();
		SerialPort.Shutdown();
		Environment.Exit(0);
		return 0;
	}
}
