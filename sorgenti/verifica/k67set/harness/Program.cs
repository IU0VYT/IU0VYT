// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE del programma del costruttore «5RH PRO» (CPS.exe,
// caricato così com'è) per le impostazioni generali del K67 Pro / 5RH Pro:
//   k67set_harness <exe> read <file.xlc>                         valori letti da ConvertSetting
//   k67set_harness <exe> write <file.xlc> <modifiche> <uscita>   ConvertSetting, modifiche, SettingConvert
// Modifiche: una per riga, Proprietà=valore (proprietà di CPS_5RH.Data.DatGenSetInfo).
using System;
using System.IO;
using System.Reflection;
using System.Runtime.CompilerServices;
using System.Text;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;

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

	static int Main(string[] args)
	{
		Encoding.RegisterProvider(CodePagesEncodingProvider.Instance);
		Assembly asm;
		// l'eseguibile è marcato «solo 32 bit» (COR20 flag 32BITREQUIRED): su .NET 8 a
		// 64 bit si carica una copia in memoria con quel flag tolto (il codice non cambia)
		try { asm = Assembly.Load(Clear32Bit(File.ReadAllBytes(args[0]))); }
		catch (Exception e) { Console.Error.WriteLine("HR=0x" + e.HResult.ToString("X") + " " + e.Message); for (var x = e.InnerException; x != null; x = x.InnerException) Console.Error.WriteLine(x); return 2; }
		var dpT = asm.GetType("CPS_5RH.DataProtocol", true);
		var appT = asm.GetType("CPS_5RH.Data.DataApp", true);
		var data = File.ReadAllBytes(args[2]);
		var full = (byte[])dpT.GetField("FullData", All).GetValue(null);
		Array.Copy(data, full, Math.Min(data.Length, full.Length));
		var dp = RuntimeHelpers.GetUninitializedObject(dpT);
		var app = Activator.CreateInstance(appT);
		dpT.GetField("bufForData", All).SetValue(dp, app);
		dpT.GetMethod("ConvertSetting", All).Invoke(dp, null);
		var gs = appT.GetField("dataGenSetInfor", All).GetValue(app);
		var gsT = gs.GetType();
		if (args[1] == "read")
		{
			foreach (var p in gsT.GetProperties(BindingFlags.Public | BindingFlags.Instance))
			{
				if (p.CanRead && p.GetIndexParameters().Length == 0)
					Console.WriteLine(p.Name + "=" + p.GetValue(gs));
			}
			return 0;
		}
		foreach (var line in File.ReadAllLines(args[3]))
		{
			if (line.Trim() == "") continue;
			int k = line.IndexOf('=');
			var p = gsT.GetProperty(line.Substring(0, k));
			p.SetValue(gs, Convert.ChangeType(line.Substring(k + 1), p.PropertyType));
		}
		dpT.GetField("appData", All).SetValue(dp, app);
		dpT.GetMethod("SettingConvert", All).Invoke(dp, null);
		var outb = new byte[data.Length];
		Array.Copy(data, outb, data.Length);
		Array.Copy(full, outb, Math.Min(data.Length, full.Length));
		File.WriteAllBytes(args[4], outb);
		return 0;
	}
}
