// IU0VYT Flash — canali, ponti e firmware delle radio Baofeng e Bajeton
// Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
// SPDX-License-Identifier: GPL-3.0-or-later
// Software libero: GNU General Public License versione 3 o successiva, SENZA
// ALCUNA GARANZIA. Testo della licenza in LICENSE.txt; autori e licenze di terzi
// in AUTORI_E_LICENZE.txt.

// Sostituto minimo di System.Windows.Forms per caricare, su .NET 8 e Linux, le
// classi dei programmi del costruttore che derivano da Form (se ne usano solo i
// campi statici, nessuna finestra viene creata).
namespace System.Windows.Forms
{
	public class Control { }
	public class ScrollableControl : Control { }
	public class ContainerControl : ScrollableControl { }
	public class Form : ContainerControl { }
}
