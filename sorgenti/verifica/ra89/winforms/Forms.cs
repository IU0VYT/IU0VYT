// IU0VYT Flash — canali, ponti e firmware delle radio ricetrasmittenti
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Sostituto di System.Windows.Forms per far girare, su .NET 8 e Linux, i metodi
// del programma di aggiornamento della Retevis RA89 («TH8600-UpDAta Soft»): i
// controlli conservano solo i valori (testo, scelta, abilitazione), nessuna
// finestra viene creata. MessageBox.Show scrive il messaggio sullo standard output
// e lo rende disponibile al programma di prova (Messages).
using System.Collections;
using System.Collections.Generic;

namespace System.Windows.Forms
{
	public interface IWin32Window { IntPtr Handle { get; } }
	public enum DialogResult { None, OK, Cancel, Abort, Retry, Ignore, Yes, No }
	public enum FormStartPosition { Manual, CenterScreen, WindowsDefaultLocation, WindowsDefaultBounds, CenterParent }
	public enum AutoScaleMode { None, Font, Dpi, Inherit }
	public enum ComboBoxStyle { Simple, DropDown, DropDownList }
	public class FormClosingEventArgs : EventArgs { public bool Cancel { get; set; } }
	public delegate void FormClosingEventHandler(object sender, FormClosingEventArgs e);

	public class Control : IWin32Window
	{
		public IntPtr Handle => IntPtr.Zero;
		public virtual string Text { get; set; } = "";
		public bool Enabled { get; set; } = true;
		public bool Visible { get; set; } = true;
		public string Name { get; set; } = "";
		public bool AutoSize { get; set; }
		public int TabIndex { get; set; }
		public event EventHandler Click, DoubleClick, Enter, TextChanged;
		public object Invoke(Delegate d)
		{
			if (d is EventHandler eh) { eh(this, EventArgs.Empty); return null; }
			return d.DynamicInvoke();
		}
		public object Invoke(Delegate d, params object[] a) => d.DynamicInvoke(a);
		public void SuspendLayout() { }
		public void ResumeLayout(bool b) { }
		public void PerformLayout() { }
	}
	public class ScrollableControl : Control { }
	public class ContainerControl : ScrollableControl { public AutoScaleMode AutoScaleMode { get; set; } }
	public class Form : ContainerControl
	{
		public FormStartPosition StartPosition { get; set; }
		public event EventHandler Load;
		public event FormClosingEventHandler FormClosing;
		public void Close() { }
		protected virtual void Dispose(bool disposing) { }
		public void Dispose() { Dispose(true); }
	}
	public class ButtonBase : Control { public bool UseVisualStyleBackColor { get; set; } }
	public class Button : ButtonBase { }
	public class CheckBox : ButtonBase { public bool Checked { get; set; } }
	public class RadioButton : ButtonBase
	{
		public bool Checked { get; set; }
		public bool TabStop { get; set; }
		public event EventHandler CheckedChanged;
	}
	public class Label : Control { }
	public class TextBox : Control { }
	public class GroupBox : Control { public bool TabStop { get; set; } }
	public class ProgressBar : Control { public int Value { get; set; } }
	public class ListControl : Control
	{
		public bool FormattingEnabled { get; set; }
		public virtual int SelectedIndex { get; set; } = -1;
	}
	public class ComboBox : ListControl
	{
		public class ObjectCollection : IEnumerable
		{
			internal readonly List<object> l = new List<object>();
			public int Add(object o) { l.Add(o); return l.Count - 1; }
			public void Clear() { l.Clear(); }
			public int Count => l.Count;
			public object this[int i] { get => l[i]; set => l[i] = value; }
			public IEnumerator GetEnumerator() => l.GetEnumerator();
		}
		public ObjectCollection Items { get; } = new ObjectCollection();
		public ComboBoxStyle DropDownStyle { get; set; }
		public object SelectedItem => SelectedIndex >= 0 && SelectedIndex < Items.Count ? Items[SelectedIndex] : null;
		// come in Windows Forms: cambiare la scelta (anche dal programma) chiama i gestori
		int sel = -1;
		public override int SelectedIndex
		{
			get => sel;
			set { if (value != sel) { sel = value; SelectedIndexChanged?.Invoke(this, EventArgs.Empty); } }
		}
		public event EventHandler SelectedIndexChanged;
	}
	public class CommonDialog { public DialogResult ShowDialog() => DialogResult.OK; }
	public class FileDialog : CommonDialog
	{
		public string FileName { get; set; } = "";
		public string Filter { get; set; }
		public string InitialDirectory { get; set; }
		public bool CheckFileExists { get; set; }
		public bool CheckPathExists { get; set; }
		public bool ShowHelp { get; set; }
		public event System.ComponentModel.CancelEventHandler FileOk;
	}
	public class OpenFileDialog : FileDialog { public bool Multiselect { get; set; } public bool ReadOnlyChecked { get; set; } }
	public class SaveFileDialog : FileDialog { }
	public static class MessageBox
	{
		public static readonly List<string> Messages = new List<string>();
		public static DialogResult Show(string text)
		{
			lock (Messages) { Messages.Add(text); Threading.Monitor.PulseAll(Messages); }
			Console.WriteLine("MESSAGGIO: " + text);
			return DialogResult.OK;
		}
		public static DialogResult Show(string text, string caption) => Show(text);
	}
	public static class Application
	{
		public static void EnableVisualStyles() { }
		public static void SetCompatibleTextRenderingDefault(bool b) { }
		public static void Run(Form f) { }
	}
}
