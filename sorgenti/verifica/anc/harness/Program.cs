// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE del programma di aggiornamento "BinTe ANC
// Bootloader" (classe KDH_Bootloader.BootHelper, caricata direttamente dal suo
// eseguibile, senza decompilarla) collegato alla radio simulata.
//   anc_harness <BinTe ANC Bootloader.exe> <firmware.anc> <host:porta> <update|recovery> <0=57600|1=115200>
// "update" = pulsante «Update» (flagPTTPress = 1); "recovery" = pulsante
// «Click Here If Your Radio Can't Work» (flagPTTPress = 0).
using System;
using System.IO.Ports;
using System.Reflection;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;

	static int Main(string[] a)
	{
		SerialPort.Endpoint = a[2];
		var asm = Assembly.LoadFrom(a[0]);
		var settingsT = asm.GetType("KDH_Bootloader.Properties.Settings", true);
		var def = settingsT.GetProperty("Default", All).GetValue(null);
		settingsT.GetProperty("BaudRate", All).SetValue(def, int.Parse(a[4]));

		// FormMain.btnDownload_Click / btn_AutoUpdate_Click
		var sp = new SerialPort();
		sp.PortName = "COM9";
		sp.BaudRate = 9600;
		sp.DataBits = 8;
		sp.StopBits = StopBits.One;
		sp.Parity = Parity.None;
		sp.DtrEnable = true;
		sp.RtsEnable = true;
		sp.Open();
		// FormMain.TaskBooting
		var bhT = asm.GetType("KDH_Bootloader.BootHelper", true);
		var ctor = bhT.GetConstructors(All)[0];
		var pt = ctor.GetParameters()[2].ParameterType;
		object flag = a[3] == "update" ? (pt == typeof(bool) ? (object)true : 1) : (pt == typeof(bool) ? (object)false : 0);
		var bh = ctor.Invoke(new object[] { sp, a[1], flag });
		MethodInfo boot = null;
		foreach (var m in bhT.GetMethods(All))
			if (m.Name == "BootLoading" && m.GetParameters().Length == 2) boot = m;
		boot.Invoke(bh, new object[] { sp, a[1] });
		var msg = (string)bhT.GetField("stateMsg", All).GetValue(bh);
		Console.WriteLine(msg.Replace("\r\n", " | "));
		sp.Close();
		SerialPort.Shutdown();
		Environment.Exit(msg.Contains("Finish!") ? 0 : 2);
		return 0;
	}
}
