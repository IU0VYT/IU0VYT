// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE del programma del costruttore «DR-1801A6»
// (DR-1801A6.exe e Auctus.NetStream.dll, caricati così come sono):
//   dr1801_harness <exe> decode <file>                  canali, zone e liste lette da DataLayout
//   dr1801_harness <exe> edit <file> <modifiche> <uscita> modifiche sugli oggetti del programma, poi DataLayout.W(true)
//   dr1801_harness <exe> read <host:porta> <uscita>      lettura via cavo (PipeLineCpsDataReadTask)
//   dr1801_harness <exe> write <host:porta> <file>       scrittura via cavo (PipeLineCpsDataWriteTask, DataLayout.W())
// Modifiche, una per riga:
//   del|n                       canale n (da 0) cancellato come il programma (tolto da zone,
//                               liste di scansione e roaming), senza rinumerare gli altri
//   set|n|tipo|potenza|rx|tx|banda|tipoRX|invRX|valRX|tipoTX|invTX|valTX|p1|p2|nome
//                               canale n: nuovo come ModuleDataChannel.CreateNew, poi i campi
//                               come la finestra del canale; un canale nuovo che non è in nessuna
//                               zona va nella prima zona con posto (regola di IU0VYT Flash)
using System;
using System.Collections;
using System.IO;
using System.IO.Ports;
using System.Reflection;
using System.Text;
using System.Threading;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;
	static Assembly asm;
	static object app;
	static Type appT;

	// proprietà o campo
	static object P(object o, string n)
	{
		var p = o.GetType().GetProperty(n, All);
		return p != null ? p.GetValue(o) : o.GetType().GetField(n, All).GetValue(o);
	}
	static void SetP(object o, string n, object v)
	{
		var p = o.GetType().GetProperty(n, All);
		p.SetValue(o, Convert.ChangeType(v, p.PropertyType));
	}
	static object F(object o, string n) => o.GetType().GetField(n, All).GetValue(o);

	static object Layout(byte[] data)
	{
		var dlT = asm.GetType("Auctus.BF1801.Cps.Data.DataLayout", true);
		var dl = Activator.CreateInstance(dlT);
		dlT.GetMethod("R").Invoke(dl, new object[] { data });
		return dl;
	}

	static string Join(IEnumerable l)
	{
		var sb = new StringBuilder();
		foreach (var x in l) { if (sb.Length > 0) sb.Append(','); sb.Append(x); }
		return sb.ToString();
	}

	static void Decode(object dl)
	{
		var ch = P(dl, "Ch");
		foreach (var c in (IList)P(ch, "ChannelList"))
		{
			if (c == null || !(bool)P(c, "Available")) continue;
			Console.WriteLine(string.Join("|", "ch", P(c, "Index"), P(c, "Type"), P(c, "Power"), P(c, "RxFreq"), P(c, "TxFreq"), P(c, "AnalogBandWidth"),
				P(c, "AnalogRxCtdcsType"), P(c, "AnalogRxDcsInvert"), P(c, "AnalogRxCtdcs"), P(c, "AnalogTxCtdcsType"), P(c, "AnalogTxDcsInvert"),
				P(c, "AnalogTxCtdcs"), P(c, "AnalogTxCtcPara1"), P(c, "AnalogTxCtcPara2"), P(c, "Name")));
		}
		var zn = P(dl, "Zn");
		foreach (var z in (IList)F(zn, "ZoneInfoList"))
			Console.WriteLine(string.Join("|", "zone", P(z, "Index"), P(z, "Name"), Join((IEnumerable)P(z, "ChannelList"))));
		foreach (var s in (IList)F(P(dl, "Sl"), "List"))
			Console.WriteLine(string.Join("|", "scan", P(s, "Index"), Join((IEnumerable)P(s, "Channels"))));
		foreach (var r in (IList)F(P(dl, "Roaming"), "List"))
			Console.WriteLine(string.Join("|", "roam", P(r, "Index"), Join((IEnumerable)P(r, "Channels"))));
	}

	static void Edit(object dl, string[] lines)
	{
		var ch = P(dl, "Ch");
		var list = (IList)P(ch, "ChannelList");
		var cdT = asm.GetType("Auctus.BF1801.Cps.Data.ChannelData", true);
		var zn = P(dl, "Zn");
		var zones = (IList)F(zn, "ZoneInfoList");
		int maxZone = Convert.ToInt32(P(zn, "MAX_ZONE_CHANNEL"));
		foreach (var line in lines)
		{
			if (line.Trim() == "") continue;
			var f = line.Split('|');
			int n = int.Parse(f[1]);
			int pos = -1;
			for (int i = 0; i < list.Count; i++)
			{
				var c = list[i];
				if (c != null && (bool)P(c, "Available") && Convert.ToInt32(P(c, "Index")) == n) pos = i;
			}
			if (f[0] == "del")
			{
				// come frm_ChannelModule (Remove + elenco vuoto), ma nella stessa posizione
				var e = Activator.CreateInstance(cdT);
				SetP(e, "Index", ushort.MaxValue);
				list[pos] = e;
				foreach (var z in zones)
				{
					var cl = (IList)P(z, "ChannelList");
					cl.Remove((ushort)n);
					SetP(z, "ChannelNum", cl.Count);
				}
				foreach (var s in (IList)F(P(dl, "Sl"), "List")) ((IList)P(s, "Channels")).Remove((ushort)n);
				foreach (var r in (IList)F(P(dl, "Roaming"), "List")) ((IList)P(r, "Channels")).Remove((short)n);
				continue;
			}
			object cd;
			bool isNew = pos < 0;
			if (isNew)
			{
				// ModuleDataChannel.CreateNew con l'indice voluto
				cd = Activator.CreateInstance(cdT);
				SetP(cd, "Index", n);
				SetP(cd, "Type", byte.Parse(f[2]));
				SetP(cd, "Available", true);
				SetP(cd, "Name", "Channel " + (n + 1));
				if (byte.Parse(f[2]) == 1) SetP(cd, "AnalogBandWidth", 2);
				SetP(cd, "Power", 1);
				list[n] = cd;
			}
			else cd = list[pos];
			SetP(cd, "Type", byte.Parse(f[2]));
			SetP(cd, "Power", byte.Parse(f[3]));
			SetP(cd, "RxFreq", uint.Parse(f[4]));
			SetP(cd, "TxFreq", uint.Parse(f[5]));
			if (f[2] == "1")
			{
				SetP(cd, "AnalogBandWidth", byte.Parse(f[6]));
				SetP(cd, "AnalogRxCtdcsType", byte.Parse(f[7])); SetP(cd, "AnalogRxDcsInvert", byte.Parse(f[8])); SetP(cd, "AnalogRxCtdcs", ushort.Parse(f[9]));
				SetP(cd, "AnalogTxCtdcsType", byte.Parse(f[10])); SetP(cd, "AnalogTxDcsInvert", byte.Parse(f[11])); SetP(cd, "AnalogTxCtdcs", ushort.Parse(f[12]));
				SetP(cd, "AnalogTxCtcPara1", ushort.Parse(f[13])); SetP(cd, "AnalogTxCtcPara2", ushort.Parse(f[14]));
			}
			SetP(cd, "Name", f[15]);
			bool inZone = false;
			foreach (var z in zones) inZone |= ((IList)P(z, "ChannelList")).Contains((ushort)n);
			if (!inZone)
			{
				foreach (var z in zones)
				{
					var cl = (IList)P(z, "ChannelList");
					if (cl.Count < maxZone) { z.GetType().GetMethod("AppendChannel", All).Invoke(z, new object[] { (ushort)n }); inZone = true; break; }
				}
			}
		}
	}

	static object PipeLine()
	{
		return appT.GetField("_pipeLine", All).GetValue(app);
	}

	static object RunTask(object task)
	{
		var done = new ManualResetEvent(false);
		object args = null;
		var ev = task.GetType().GetEvent("Finished", All);
		var argT = ev.EventHandlerType.GetGenericArguments()[0];
		EventHandler<EventArgs> dummy = null;
		var h = Delegate.CreateDelegate(ev.EventHandlerType, new Holder(done, a => args = a), typeof(Holder).GetMethod("On"));
		ev.AddEventHandler(task, h);
		var ctxT = asm.GetType("Auctus.BF1801.Cps.Core.PipeLine.PipeLineContext", true);
		var ctx = ctxT.GetConstructors(All)[0].Invoke(new object[] { PipeLine() });
		task.GetType().GetMethod("Execute", All).Invoke(task, new object[] { ctx });
		if (!done.WaitOne(240000)) throw new Exception("tempo scaduto");
		_ = dummy;
		return args;
	}

	public class Holder
	{
		readonly ManualResetEvent d; readonly Action<object> a;
		public Holder(ManualResetEvent d, Action<object> a) { this.d = d; this.a = a; }
		public void On(object s, EventArgs e) { a(e); d.Set(); }
	}

	static int Main(string[] a)
	{
		Encoding.RegisterProvider(CodePagesEncodingProvider.Instance);
		asm = Assembly.LoadFrom(a[0]);
		appT = asm.GetType("Auctus.BF1801.Cps.Core.App", true);
		app = appT.GetField("Instance", All).GetValue(null);
		var msT = asm.GetType("Auctus.BF1801.Cps.Core.DefaultModuleService", true);
		appT.GetProperty("ModuleService", All).SetValue(app, msT.GetConstructors(All)[0].Invoke(new object[] { null }));
		switch (a[1])
		{
			case "decode":
				Decode(Layout(File.ReadAllBytes(a[2])));
				return 0;
			case "edit":
			{
				var dl = Layout(File.ReadAllBytes(a[2]));
				Edit(dl, File.ReadAllLines(a[3]));
				File.WriteAllBytes(a[4], (byte[])dl.GetType().GetMethod("W").Invoke(dl, new object[] { true }));
				return 0;
			}
			case "read":
			case "write":
			{
				SerialPort.Endpoint = a[2];
				appT.GetMethod("TryBuildPipeLine", All).Invoke(app, new object[] { "COM9", 9600, true });
				object task;
				if (a[1] == "read")
					task = asm.GetType("Auctus.BF1801.Cps.Core.PipeLine.PipeLineCpsDataReadTask", true).GetConstructors(All)[0].Invoke(new object[] { "" });
				else
				{
					var dl = Layout(File.ReadAllBytes(a[3]));
					var body = (byte[])dl.GetType().GetMethod("W").Invoke(dl, new object[] { false });
					task = asm.GetType("Auctus.BF1801.Cps.Core.PipeLine.PipeLineCpsDataWriteTask", true).GetConstructors(All)[0].Invoke(new object[] { "", (ushort)1, body });
				}
				var r = RunTask(task);
				bool ok = (bool)P(r, "Succes");
				Console.Error.WriteLine("esito: " + ok + " " + P(r, "FailureMsg"));
				if (a[1] == "read" && ok)
				{
					var data = (IList)P(task, "DataReaded");
					var b = new byte[data.Count];
					data.CopyTo(b, 0);
					File.WriteAllBytes(a[3], b);
				}
				SerialPort.Shutdown();
				Environment.Exit(ok ? 0 : 4);
				return 0;
			}
		}
		return 1;
	}
}
