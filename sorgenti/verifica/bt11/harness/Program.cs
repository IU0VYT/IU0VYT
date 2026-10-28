// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE dei programmi Bajeton «BT-11_CPS» (BJ-8300 /
// BJ-7800) e «BJ-9000_CPS» (BT-RT950PRO_CPS.exe), caricati così come sono dai loro
// eseguibili (stessa struttura: classi KDH.RadioData e KDH.RWDataOperation), per il
// confronto con IU0VYT Flash:
//   bt11_harness <exe> read  <host:porta> <uscita.dat>   legge la radio simulata (RWDataOperation READ) e salva il file
//   bt11_harness <exe> write <host:porta> <file.dat>     apre il file e lo scrive nella radio simulata (WRITE)
//   bt11_harness <exe> dump  <file.dat>                  contenuto del file, campo per campo
//   bt11_harness <exe> fun   <file.dat> [<modifiche> <uscita.dat>]
//        impostazioni (KDH.FunConfigData): le stampa; con le modifiche (Proprietà=valore,
//        una per riga) le cambia e salva il file con RadioData.SaveToFile
//   bt11_harness <exe> new   <uscita.dat>                 file nuovo (new RadioData(), come «Nuovo»)
using System;
using System.Collections;
using System.Globalization;
using System.IO;
using System.IO.Ports;
using System.Reflection;
using System.Text;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;
	static Assembly asm;

	static byte[] Clear32Bit(byte[] b)
	{
		int pe = BitConverter.ToInt32(b, 0x3C);
		bool pe32plus = BitConverter.ToUInt16(b, pe + 24) == 0x20B;
		int dd = pe + 24 + (pe32plus ? 112 : 96);
		int cliRva = BitConverter.ToInt32(b, dd + 14 * 8);
		int nsec = BitConverter.ToUInt16(b, pe + 6);
		int sec = pe + 24 + BitConverter.ToUInt16(b, pe + 20);
		for (int i = 0; i < nsec; i++, sec += 40)
		{
			int va = BitConverter.ToInt32(b, sec + 12), vs = BitConverter.ToInt32(b, sec + 8), raw = BitConverter.ToInt32(b, sec + 20);
			if (cliRva >= va && cliRva < va + vs)
			{
				int f = raw + cliRva - va + 16;
				uint flags = BitConverter.ToUInt32(b, f) & ~2u;
				BitConverter.GetBytes(flags).CopyTo(b, f);
				break;
			}
		}
		return b;
	}

	static object Load(string path)
	{
		var rdT = asm.GetType("KDH.RadioData", true);
		using (var s = File.OpenRead(path))
			return rdT.GetMethod("CreatObjFromFile", All).Invoke(null, new object[] { s });
	}

	static void Save(object rd, string path)
	{
		using (var s = new FileStream(path, FileMode.Create))
			rd.GetType().GetMethod("SaveToFile", All).Invoke(rd, new object[] { s });
	}

	static void Dump(string path, object o, StringBuilder sb)
	{
		if (o == null) { sb.Append(path).Append("=<null>\n"); return; }
		var t = o.GetType();
		if (t.IsPrimitive || o is string || t.IsEnum)
		{
			sb.Append(path).Append('=').Append(Convert.ToString(o, CultureInfo.InvariantCulture)).Append('\n');
			return;
		}
		if (o is IList l)
		{
			for (int i = 0; i < l.Count; i++) Dump(path + "[" + i + "]", l[i], sb);
			return;
		}
		for (var tt = t; tt != null && tt != typeof(object); tt = tt.BaseType)
			foreach (var f in tt.GetFields(BindingFlags.Instance | BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.DeclaredOnly))
				if (!f.IsNotSerialized) Dump(path + "." + f.Name, f.GetValue(o), sb);
	}

	static int Main(string[] a)
	{
		Encoding.RegisterProvider(CodePagesEncodingProvider.Instance);
		try { asm = Assembly.Load(Clear32Bit(File.ReadAllBytes(a[0]))); }
		catch (Exception e) { Console.Error.WriteLine(e); return 2; }
		// i file contengono il nome dell'assembly «BT-11_CPS»: lo si risolve con quello caricato
		AppDomain.CurrentDomain.AssemblyResolve += (s, e) => new AssemblyName(e.Name).Name == asm.GetName().Name ? asm : null;
		var rdT = asm.GetType("KDH.RadioData", true);
		switch (a[1])
		{
			case "new":
				Save(Activator.CreateInstance(rdT), a[2]);
				return 0;
			case "dump":
			{
				var sb = new StringBuilder();
				Dump("radio", Load(a[2]), sb);
				Console.Write(sb.ToString());
				return 0;
			}
			case "fun":
			{
				var rd = Load(a[2]);
				var fc = rdT.GetField("funConfigData", All).GetValue(rd);
				var fT = fc.GetType();
				if (a.Length > 3)
				{
					foreach (var line in File.ReadAllLines(a[3]))
					{
						if (line.Trim() == "") continue;
						int k = line.IndexOf('=');
						fT.GetProperty(line.Substring(0, k)).SetValue(fc, int.Parse(line.Substring(k + 1), CultureInfo.InvariantCulture));
					}
					Save(rd, a[4]);
					return 0;
				}
				foreach (var p in fT.GetProperties(BindingFlags.Public | BindingFlags.Instance))
					Console.WriteLine(p.Name + "=" + p.GetValue(fc));
				return 0;
			}
			case "read":
			case "write":
			{
				SerialPort.Endpoint = a[2];
				// FormProgressBar: porta aperta prima di avviare RWDataOperation
				var sp = new SerialPort();
				sp.PortName = "COM9";
				sp.BaudRate = 9600;
				sp.DataBits = 8;
				sp.StopBits = StopBits.One;
				sp.Parity = Parity.None;
				sp.Open();
				object rd = a[1] == "read" ? Activator.CreateInstance(rdT) : Load(a[3]);
				var opT = asm.GetType("KDH.OPERATION_TYPE", true);
				var rwT = asm.GetType("KDH.RWDataOperation", true);
				var rw = Activator.CreateInstance(rwT, new object[] { sp, rd, Enum.Parse(opT, a[1] == "read" ? "READ" : "WRITE") });
				// DoIt: bool (BT-11_CPS) oppure OPERATION_RESULT (BJ-9000_CPS: COMMOK = riuscita)
				var res = rwT.GetMethod("DoIt", All).Invoke(rw, null);
				bool ok = res is bool b ? b : res.ToString() == "COMMOK";
				SerialPort.Shutdown();
				if (!ok) { Console.Error.WriteLine("DoIt: operazione non riuscita"); Environment.Exit(3); }
				if (a[1] == "read") Save(rd, a[3]);
				Environment.Exit(0);
				return 0;
			}
		}
		Console.Error.WriteLine("comando sconosciuto");
		return 1;
	}
}
