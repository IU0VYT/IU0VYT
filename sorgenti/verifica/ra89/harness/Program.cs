// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Fa girare il codice ORIGINALE dei programmi di aggiornamento della Retevis RA89
// («RA89-update-software.exe», «TH8600-UpDAta Soft» Ver 2.3 20230621) e della
// Retevis RT85 («RT85 10W_5W_UpData.exe», Ver 2.1 20200701): classe
// FlaUPDataProject.Form1, caricata direttamente dall'eseguibile, senza
// decompilarla, collegata alla radio simulata.
//   ra89_harness <programma.exe> <firmware> <host:porta> [modello] [velocità] [lines]
// modello: indice della scelta «radio» del programma (RA89: 0 UV88, 1 UV98,
// 2 RT45P, 3 RA89, quella proposta all'avvio; RT85: 0 5W, 1 10W, quella proposta);
// velocità: 57600 (proposta) o 115200.
// «lines» stampa le righe che il programma prepara dal file (WriteDataBuf),
// senza collegarsi alla radio.
// Il modulo non viene costruito (il costruttore disegna la finestra): i campi
// sono impostati con i valori del costruttore e di InitializeComponent (filo a due
// conduttori: radioButton4; textBox2 «20»), i gestori delle scelte «radio» e
// «velocità» sono collegati come in InitializeComponent (4.5: impostano la
// risposta attesa dalla radio e il ritardo), poi Form1_Load e il pulsante
// «Aggiorna» (EnterWriteModPro).
using System;
using System.IO.Ports;
using System.Reflection;
using System.Runtime.CompilerServices;
using System.Windows.Forms;

static class Program
{
	const BindingFlags All = BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance | BindingFlags.Static;
	static Type fT;
	static object f;

	static void Set(string n, object v)
	{
		var fi = fT.GetField(n, All);
		if (v != null && fi.FieldType.IsPrimitive && v.GetType() != fi.FieldType) v = Convert.ChangeType(v, fi.FieldType);
		fi.SetValue(f, v);
	}
	static object Get(string n) => fT.GetField(n, All).GetValue(f);
	static void Call(string n, params object[] a) => fT.GetMethod(n, All).Invoke(f, a);

	static int Main(string[] a)
	{
		var asm = Assembly.LoadFrom(a[0]);
		fT = asm.GetType("FlaUPDataProject.Form1", true);
		var c1 = asm.GetType("FlaUPDataProject.Class1", true);
		c1.GetField("English_En", All).SetValue(null, 1);
		f = RuntimeHelpers.GetUninitializedObject(fT);
		// costruttore (campi)
		Set("GbprogressBar", false);
		Set("ComUseAble", new bool[16]);
		Set("ConAckPass", new byte[] { 0x50, 0x41, 0x53, 0x53 });
		Set("ConHandAck", new byte[] { 0xFE, 0xFE, 0xEF, 0xEE, 0xE1, 0x56 });
		Set("ConAckExitCode", new byte[] { 0x55, 0, 0 });
		Set("ConAckIapMod", new byte[] { 0xAA, 0x01, 0xEF });
		Set("ConAckSucceed", new byte[] { 0xAA, 0x02, 0xEF });
		Set("ConAckFail", new byte[] { 0xAA, 0x03, 0xEF });
		Set("ConAckSetcodeAdd", new byte[] { 0xAA, 0x11, 0xEF });
		Set("ConAckCodeAddErro", new byte[] { 0xAA, 0x44, 0xEF });
		Set("ConAckExit", new byte[] { 0xAA, 0xFF, 0xEF });
		Set("GiFileName", "FlaTempFile.icf");
		Set("GiCurFileName", "FlaTempFile.icf");
		Set("GiUpDataFileName", "");
		Set("SendByte", 0);
		Set("GbCommDelay", 5);
		Set("RxBuffer", new byte[200]);
		Set("GiSendStr", "");
		if (fT.GetField("GiDataLeg", All) != null) Set("GiDataLeg", 128); // solo RA89
		Set("CommunicationStep", 0);
		Set("GbReadMod", true);
		Set("GbFirComEn", true);
		Set("HandCom", new[] { "aa01ef", "aa01ef", "aa01ef", "aa01ef", "aa01ef", "aa01ef" });
		// InitializeComponent: i controlli usati
		foreach (var fi in fT.GetFields(All))
		{
			if (fi.IsStatic || fi.GetValue(f) != null) continue;
			var t = fi.FieldType;
			if (t.Namespace == "System.Windows.Forms" && !t.IsAbstract && t.GetConstructor(Type.EmptyTypes) != null)
				fi.SetValue(f, Activator.CreateInstance(t));
		}
		var sp = new SerialPort();
		SerialPort.Endpoint = a[2];
		Set("serialPort1", sp);
		fT.GetMethod("serialPort1_DataReceived", All).CreateDelegate(typeof(SerialDataReceivedEventHandler), f);
		sp.DataReceived += (SerialDataReceivedEventHandler)fT.GetMethod("serialPort1_DataReceived", All).CreateDelegate(typeof(SerialDataReceivedEventHandler), f);
		((TextBox)Get("textBox2")).Text = "20";
		((RadioButton)Get("radioButton4")).Checked = true;
		var cb1 = (ComboBox)Get("comboBox1");
		var cb2 = (ComboBox)Get("comboBox2");
		var cb3 = (ComboBox)Get("comboBox3");
		// gestori collegati da InitializeComponent
		cb1.SelectedIndexChanged += (EventHandler)fT.GetMethod("comboBox1_SelectedIndexChanged", All).CreateDelegate(typeof(EventHandler), f);
		cb3.SelectedIndexChanged += (EventHandler)fT.GetMethod("comboBox3_SelectedIndexChanged", All).CreateDelegate(typeof(EventHandler), f);
		// Form1_Load (senza la ricerca delle porte)
		Set("OnlyOneLine", ((RadioButton)Get("radioButton3")).Checked);
		Call("SetRadioKindPro");
		Call("SetBaudRateKindPro");
		cb2.Items.Add("COM9");
		cb2.SelectedIndex = 0;
		if (a.Length > 3) cb1.SelectedIndex = int.Parse(a[3]);
		if (a.Length > 4) cb3.SelectedIndex = a[4] == "115200" ? 1 : 0;
		Call("TimerResPro");
		Set("RadioType", cb1.SelectedIndex);
		// OpenFile + FileOk
		if (fT.GetField("typekind", All) != null) Set("typekind", cb1.SelectedIndex);
		((OpenFileDialog)Get("openFileDialog1")).FileName = a[1];
		Call("openFileDialog1_FileOk", null, null);
		if (a.Length > 5 && a[5] == "lines")
		{
			if (fT.GetMethod("GetFileDataPro", All) != null) Call("GetFileDataPro"); // RA89 (.bin); RT85: già in FileOk
			Console.Error.WriteLine("RISPOSTA " + BitConverter.ToString((byte[])Get("ConAckIapMod")) + " RITARDO " + ((TextBox)Get("textBox2")).Text);
			foreach (var l in (string[])Get("WriteDataBuf")) Console.WriteLine("RIGA " + l);
			return 0;
		}
		// pulsante «Aggiorna»
		Call("EnterWriteModPro", null, EventArgs.Empty);
		lock (MessageBox.Messages)
		{
			var end = DateTime.UtcNow.AddSeconds(180);
			while (MessageBox.Messages.Count == 0 && DateTime.UtcNow < end)
				System.Threading.Monitor.Wait(MessageBox.Messages, 500);
		}
		System.Threading.Thread.Sleep(300);
		SerialPort.Shutdown();
		var ok = MessageBox.Messages.Count > 0 && MessageBox.Messages[0].StartsWith("Upgrade complete");
		Environment.Exit(ok ? 0 : 2);
		return 0;
	}
}
