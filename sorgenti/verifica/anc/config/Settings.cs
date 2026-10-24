// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Sostituto minimo delle impostazioni (.NET Framework ApplicationSettingsBase)
// usate dal programma originale (Properties.Settings: BaudRate, dirpath).
using System;
using System.Collections.Generic;
using System.Reflection;

namespace System.Configuration
{
	public abstract class SettingsBase
	{
		public static readonly Dictionary<string, object> Values = new Dictionary<string, object>();
		public static SettingsBase Synchronized(SettingsBase b) => b;
		public virtual object this[string name]
		{
			get
			{
				lock (Values)
				{
					if (Values.TryGetValue(name, out var v)) return v;
				}
				var p = GetType().GetProperty(name, BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance);
				var a = p?.GetCustomAttribute<DefaultSettingValueAttribute>();
				object d = a == null ? null : (p.PropertyType == typeof(string) ? a.Value : Convert.ChangeType(a.Value, p.PropertyType));
				if (d == null && p != null && p.PropertyType.IsValueType) d = Activator.CreateInstance(p.PropertyType);
				return d;
			}
			set { lock (Values) Values[name] = value; }
		}
		public virtual void Save() { }
	}

	public abstract class ApplicationSettingsBase : SettingsBase
	{
		protected ApplicationSettingsBase() { }
		public override object this[string propertyName] { get => base[propertyName]; set => base[propertyName] = value; }
		public override void Save() { }
		public void Reload() { }
	}

	[AttributeUsage(AttributeTargets.Property)] public sealed class UserScopedSettingAttribute : Attribute { }
	[AttributeUsage(AttributeTargets.Property)] public sealed class ApplicationScopedSettingAttribute : Attribute { }
	[AttributeUsage(AttributeTargets.Property)]
	public sealed class DefaultSettingValueAttribute : Attribute
	{
		public DefaultSettingValueAttribute(string value) { Value = value; }
		public string Value { get; }
	}
}
