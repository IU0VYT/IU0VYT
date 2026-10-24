// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE del programma di aggiornamento Baofeng
// "BF_Upgrade_Tool_for_TFT_330L_Series.exe" (升级工具_V1.0, classe
// KDH_Bootloader.BootHelper, caricata direttamente dal suo eseguibile, senza
// decompilarla) collegato alla radio simulata.
//   bf_harness <BF_Upgrade_Tool…exe> <firmware.BF> <host:porta> <auto|manual> [velocità]
// "auto" = pulsante 自动升级 (flagPTTPress = false); "manual" = 手动升级 (true).
// Serve anche per gli altri programmi con la stessa classe BootHelper:
// "Baofeng Bootloader Pro" (UV-25 Plus: auto 9600, manual 115200),
// "BJ9000_EnUPDATE" (BJ-9000: 115200) e "Bootloader_BaudRate38400" (UV-9G:
// auto 9600, manual 38400). La velocità di apertura della porta è quella dei
// pulsanti del programma (FormMain); se non indicata 115200.
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
		// FormMain.btnDownload_Click / btn_AutoUpdate_Click
		var sp = new SerialPort();
		sp.PortName = "COM9";
		sp.BaudRate = a.Length > 4 ? int.Parse(a[4]) : 115200;
		sp.DataBits = 8;
		sp.StopBits = StopBits.One;
		sp.Parity = Parity.None;
		sp.DtrEnable = true;
		sp.RtsEnable = true;
		sp.Open();
		// FormMain.TaskBooting
		var bhT = asm.GetType("KDH_Bootloader.BootHelper", true);
		var bh = bhT.GetConstructors(All)[0].Invoke(new object[] { sp, a[1], a[3] == "manual" });
		MethodInfo boot = null;
		foreach (var m in bhT.GetMethods(All))
			if (m.Name == "BootLoading" && m.GetParameters().Length == 2) boot = m;
		boot.Invoke(bh, new object[] { sp, a[1] });
		var msg = (string)bhT.GetField("stateMsg", All).GetValue(bh);
		Console.WriteLine(msg.Replace("\r\n", " | "));
		sp.Close();
		SerialPort.Shutdown();
		Environment.Exit(msg.Contains("下载完成") || msg.Contains("Download Completed") ? 0 : 2);
		return 0;
	}
}
