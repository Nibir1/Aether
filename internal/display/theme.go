// internal/display/theme.go
//
// Theme definitions for Aether's Display subsystem.
// Updated in Stage 18 to include full table styling support.

package display

//
// ───────────────────────────────────────────────────────────────────────
//                           COLOR MODE
// ───────────────────────────────────────────────────────────────────────
//

type ColorMode int

const (
	ColorModeAuto ColorMode = iota
	ColorModeAlways
	ColorModeNever
)

//
// ───────────────────────────────────────────────────────────────────────
//                           HEADING STYLES
// ───────────────────────────────────────────────────────────────────────
//

// HeadingStyle describes how a given heading level should be rendered.
type HeadingStyle struct {
	Prefix        string
	Underline     bool
	UnderlineRune rune
	Uppercase     bool
}

//
// ───────────────────────────────────────────────────────────────────────
//                          INLINE EMPHASIS STYLE
// ───────────────────────────────────────────────────────────────────────
//

type EmphasisStyle struct {
	Strong string
	Em     string
	Code   string
}

//
// ───────────────────────────────────────────────────────────────────────
//                          TABLE STYLE (NEW)
// ───────────────────────────────────────────────────────────────────────
//

// TableStyle controls optional styling for a table row.
type TableStyle struct {
	Bold     bool // apply strong/bold styling
	Faint    bool // apply faint/metadata styling
	Colorize bool // colorize (theme-aware)
}

//
// ───────────────────────────────────────────────────────────────────────
//                           THEME (EXTENDED)
// ───────────────────────────────────────────────────────────────────────
//

// Theme defines how Aether should render Markdown, tables, and text.
type Theme struct {
	Name     string
	Color    ColorMode
	MaxWidth int

	Indent    string
	Bullet    string
	CodeFence string

	HeadingStyles map[int]HeadingStyle
	Emphasis      EmphasisStyle

	ShowSectionRoles bool

	// ─── Table rendering extensions ────────────────────────────────
	TablePadding     int
	TableHeaderStyle TableStyle
	TableBodyStyle   TableStyle
	AsciiOnly        bool
}

//
// ───────────────────────────────────────────────────────────────────────
//                      BUILT-IN THEMES (UPDATED)
// ───────────────────────────────────────────────────────────────────────
//

// DefaultTheme — Markdown-first, general purpose.
func DefaultTheme() Theme {
	return sanitizeTheme(Theme{
		Name:      "default",
		Color:     ColorModeAuto,
		MaxWidth:  0,
		Indent:    "  ",
		Bullet:    "-",
		CodeFence: "```",

		HeadingStyles: map[int]HeadingStyle{
			1: {Prefix: "# "},
			2: {Prefix: "## "},
			3: {Prefix: "### "},
		},

		Emphasis: EmphasisStyle{
			Strong: "**",
			Em:     "_",
			Code:   "`",
		},

		ShowSectionRoles: false,

		TablePadding: 2,
		TableHeaderStyle: TableStyle{
			Bold:     true,
			Colorize: false,
		},
		TableBodyStyle: TableStyle{
			Bold:     false,
			Faint:    false,
			Colorize: false,
		},
		AsciiOnly: false,
	})
}

// DarkTheme — same as default but allows colorization.
func DarkTheme() Theme {
	t := DefaultTheme()
	t.Name = "dark"
	t.Color = ColorModeAlways
	return sanitizeTheme(t)
}

// MinimalTheme — very compact, no ANSI, no bold.
func MinimalTheme() Theme {
	return sanitizeTheme(Theme{
		Name:      "minimal",
		Color:     ColorModeNever,
		MaxWidth:  0,
		Indent:    "  ",
		Bullet:    "-",
		CodeFence: "```",

		HeadingStyles: map[int]HeadingStyle{
			1: {Underline: true, UnderlineRune: '='},
			2: {Underline: true, UnderlineRune: '-'},
		},

		Emphasis: EmphasisStyle{
			Strong: "",
			Em:     "",
			Code:   "`",
		},

		ShowSectionRoles: false,

		TablePadding: 1,
		TableHeaderStyle: TableStyle{
			Bold: false,
		},
		TableBodyStyle: TableStyle{
			Bold: false,
		},
		AsciiOnly: true,
	})
}

// PaperTheme — printable output, limited width, clear headers.
func PaperTheme() Theme {
	return sanitizeTheme(Theme{
		Name:      "paper",
		Color:     ColorModeNever,
		MaxWidth:  80,
		Indent:    "    ",
		Bullet:    "•",
		CodeFence: "```",

		HeadingStyles: map[int]HeadingStyle{
			1: {Underline: true, UnderlineRune: '='},
			2: {Underline: true, UnderlineRune: '-'},
			3: {Prefix: "• "},
		},

		Emphasis: EmphasisStyle{
			Strong: "**",
			Em:     "_",
			Code:   "`",
		},

		ShowSectionRoles: true,

		TablePadding: 2,
		TableHeaderStyle: TableStyle{
			Bold: true,
		},
		TableBodyStyle: TableStyle{
			Bold: false,
		},
		AsciiOnly: true,
	})
}

//
// ───────────────────────────────────────────────────────────────────────
//                             HELPERS
// ───────────────────────────────────────────────────────────────────────
//

func sanitizeTheme(t Theme) Theme {
	if t.Indent == "" {
		t.Indent = "  "
	}
	if t.Bullet == "" {
		t.Bullet = "-"
	}
	if t.CodeFence == "" {
		t.CodeFence = "```"
	}
	if t.HeadingStyles == nil {
		t.HeadingStyles = map[int]HeadingStyle{
			1: {Prefix: "# "},
		}
	}
	if t.TablePadding <= 0 {
		t.TablePadding = 2
	}
	return t
}

// HeadingForLevel resolves heading style for a given level.
func (t Theme) HeadingForLevel(level int) HeadingStyle {
	if level <= 0 {
		level = 1
	}
	if hs, ok := t.HeadingStyles[level]; ok {
		return hs
	}
	// nearest fallback
	for l := level - 1; l >= 1; l-- {
		if hs, ok := t.HeadingStyles[l]; ok {
			return hs
		}
	}
	return HeadingStyle{Prefix: "# "}
}

// EffectiveWidth resolves usable width.
func (t Theme) EffectiveWidth(fallback int) int {
	if t.MaxWidth > 0 {
		return t.MaxWidth
	}
	if fallback > 0 {
		return fallback
	}
	return 80
}
