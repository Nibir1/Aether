// internal/display/themes.go
//
// Defines display themes for the Markdown renderer. Themes influence
// headings, separators, quote markers, metadata formatting, and
// overall stylistic accents. Rendering output remains valid Markdown
// regardless of theme; themes only change the visual flavor.

package display

// Theme defines the structural and visual markers for Markdown output.
// All fields are pure-text fragments or prefixes; no ANSI escapes are
// added here (that will be a future stage if we introduce TTY color).
type Theme struct {
	// HeadingPrefix determines the Markdown prefix used for section headings.
	// Examples:
	//   "# "   → big heading
	//   "## "  → medium heading
	HeadingPrefix string

	// MetadataPrefix controls how metadata key/value lines appear.
	// Example: "- " or "• " or "> " etc.
	MetadataPrefix string

	// SectionDivider is printed between major document sections.
	SectionDivider string

	// ParagraphSpacing controls vertical space between paragraphs.
	ParagraphSpacing string
}

// Built-in theme definitions.

var Minimalist = Theme{
	HeadingPrefix:    "## ",
	MetadataPrefix:   "- ",
	SectionDivider:   "\n---\n\n",
	ParagraphSpacing: "\n\n",
}

var GitHubDark = Theme{
	HeadingPrefix:    "## ",
	MetadataPrefix:   "- ",
	SectionDivider:   "\n---\n\n",
	ParagraphSpacing: "\n\n",
}

var SolarizedLight = Theme{
	HeadingPrefix:    "### ",
	MetadataPrefix:   "* ",
	SectionDivider:   "\n---\n\n",
	ParagraphSpacing: "\n\n",
}

var SolarizedDark = Theme{
	HeadingPrefix:    "### ",
	MetadataPrefix:   "* ",
	SectionDivider:   "\n---\n\n",
	ParagraphSpacing: "\n\n",
}

var Gruvbox = Theme{
	HeadingPrefix:    "## ",
	MetadataPrefix:   "- ",
	SectionDivider:   "\n---\n\n",
	ParagraphSpacing: "\n\n",
}

var Monokai = Theme{
	HeadingPrefix:    "## ",
	MetadataPrefix:   "- ",
	SectionDivider:   "\n---\n\n",
	ParagraphSpacing: "\n\n",
}

// DefaultTheme is used if the caller does not specify a theme.
var DefaultTheme = Minimalist
