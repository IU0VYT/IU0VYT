// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Sostituto di System.IO.Ports.SerialPort per il confronto con il codice
// originale del programma di aggiornamento: la "porta" è un collegamento TCP
// con la radio simulata (test Go TestDiffAnc). Ogni apertura e chiusura della
// porta e ogni cambio di velocità sono comunicati al simulatore, che simula una
// UART: i byte scambiati a velocità diverse arrivano rovinati.
//   PC → simulatore:  'O' baud(4) | 'C' | 'W' n(4) dati
//   simulatore → PC:  'D' n(4) dati
// L'evento DataReceived (usato da alcuni programmi, es. «DR-1801A6») viene
// chiamato dal filo di lettura dopo ogni blocco ricevuto.
using System;
using System.Collections.Generic;
using System.Net.Sockets;
using System.Threading;

namespace System.IO.Ports
{
	public enum Parity { None, Odd, Even, Mark, Space }
	public enum StopBits { None, One, Two, OnePointFive }
	public enum Handshake { None, XOnXOff, RequestToSend, RequestToSendXOnXOff }
	public enum SerialData { Chars = 1, Eof = 2 }
	public enum SerialError { TXFull = 256, RXOver = 1, Overrun = 2, RXParity = 4, Frame = 8 }
	public class SerialDataReceivedEventArgs : EventArgs
	{
		public SerialDataReceivedEventArgs(SerialData t) { EventType = t; }
		public SerialData EventType { get; }
	}
	public class SerialErrorReceivedEventArgs : EventArgs
	{
		public SerialError EventType { get; set; }
	}
	public delegate void SerialDataReceivedEventHandler(object sender, SerialDataReceivedEventArgs e);
	public delegate void SerialErrorReceivedEventHandler(object sender, SerialErrorReceivedEventArgs e);

	public class SerialPort : IDisposable
	{
		public static string Endpoint = "127.0.0.1:0";
		static TcpClient c;
		static NetworkStream s;
		static readonly object wlock = new object();
		readonly List<byte> inbuf = new List<byte>();
		bool open;
		static SerialPort current;

		public string PortName { get; set; }
		public int BaudRate { get; set; } = 9600;
		public int DataBits { get; set; }
		public Parity Parity { get; set; }
		public StopBits StopBits { get; set; }
		public Handshake Handshake { get; set; }
		public bool DtrEnable { get; set; }
		public bool RtsEnable { get; set; }
		public int ReadTimeout { get; set; } = -1;
		public int WriteTimeout { get; set; } = -1;
		public bool IsOpen => open;
		public int ReceivedBytesThreshold { get; set; } = 1;
		// eventi come la porta vera: chiamati dal filo di lettura, uno alla volta
		public event SerialDataReceivedEventHandler DataReceived;
#pragma warning disable 67
		public event SerialErrorReceivedEventHandler ErrorReceived;
#pragma warning restore 67
		public static string[] GetPortNames() => new[] { "COM9" };

		public SerialPort() { }
		public SerialPort(string portName, int baudRate) { PortName = portName; BaudRate = baudRate; }

		public void DiscardInBuffer() { lock (inbuf) inbuf.Clear(); }
		public void DiscardOutBuffer() { }

		static void Connect()
		{
			if (c != null) return;
			var p = Endpoint.Split(':');
			c = new TcpClient(p[0], int.Parse(p[1])) { NoDelay = true };
			s = c.GetStream();
			var t = new Thread(Reader) { IsBackground = true };
			t.Start();
		}

		static void Reader()
		{
			var h = new byte[5];
			try
			{
				while (true)
				{
					ReadExact(h, 5);
					int n = BitConverter.ToInt32(h, 1);
					var d = new byte[n];
					ReadExact(d, n);
					var cur = current;
					if (cur == null) continue;
					lock (cur.inbuf)
					{
						if (cur.open) cur.inbuf.AddRange(d);
						Monitor.PulseAll(cur.inbuf);
					}
					var ev = cur.DataReceived;
					if (ev != null && cur.open)
					{
						try { ev(cur, new SerialDataReceivedEventArgs(SerialData.Chars)); } catch (Exception) { }
					}
				}
			}
			catch (Exception) { }
		}

		static void ReadExact(byte[] b, int n)
		{
			int got = 0;
			while (got < n)
			{
				int r = s.Read(b, got, n - got);
				if (r <= 0) throw new IOException("fine");
				got += r;
			}
		}

		static void Send(byte kind, byte[] data, int off, int n)
		{
			lock (wlock)
			{
				var f = new byte[5 + n];
				f[0] = kind;
				BitConverter.GetBytes(n).CopyTo(f, 1);
				if (n > 0) Array.Copy(data, off, f, 5, n);
				s.Write(f, 0, f.Length);
				s.Flush();
			}
		}

		public void Open()
		{
			if (open) throw new InvalidOperationException("porta già aperta");
			Connect();
			lock (inbuf) { inbuf.Clear(); open = true; }
			current = this;
			Send((byte)'O', BitConverter.GetBytes(BaudRate), 0, 4);
		}

		public void Close()
		{
			if (!open) return;
			lock (inbuf) { open = false; inbuf.Clear(); }
			Send((byte)'C', new byte[0], 0, 0);
		}

		public void Dispose() { Close(); }

		public int BytesToRead { get { lock (inbuf) return inbuf.Count; } }

		public void Write(byte[] b, int off, int n)
		{
			if (!open) throw new InvalidOperationException("porta chiusa");
			Send((byte)'W', b, off, n);
		}

		public int Read(byte[] b, int off, int n)
		{
			lock (inbuf)
			{
				var limit = ReadTimeout < 0 ? DateTime.MaxValue : DateTime.UtcNow.AddMilliseconds(ReadTimeout);
				while (inbuf.Count == 0)
				{
					if (!open) throw new InvalidOperationException("porta chiusa");
					if (DateTime.UtcNow >= limit) throw new TimeoutException("Timeout dell'operazione");
					Monitor.Wait(inbuf, 100);
				}
				int k = Math.Min(n, inbuf.Count);
				inbuf.CopyTo(0, b, off, k);
				inbuf.RemoveRange(0, k);
				return k;
			}
		}

		public int ReadByte()
		{
			var b = new byte[1];
			Read(b, 0, 1);
			return b[0];
		}

		public static void Shutdown() { try { s?.Dispose(); c?.Close(); } catch (Exception) { } }
	}
}
