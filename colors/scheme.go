package colors

import "github.com/charmbracelet/lipgloss"

var (
	MainBg                 = lipgloss.AdaptiveColor{Light: "63", Dark: "63"}
	MainText               = lipgloss.AdaptiveColor{Light: "228", Dark: "228"}
	PromptText             = lipgloss.AdaptiveColor{Light: "27", Dark: "39"}
	DocumentText           = lipgloss.AdaptiveColor{Light: "234", Dark: "252"}
	DocumentTextDim        = lipgloss.AdaptiveColor{Light: "240", Dark: "249"}
	DocumentTextBright     = lipgloss.AdaptiveColor{Light: "60", Dark: "230"}
	HelperText             = lipgloss.AdaptiveColor{Light: "244", Dark: "243"}
	HelperTextDim          = lipgloss.AdaptiveColor{Light: "249", Dark: "240"}
	HelperHighlightText    = lipgloss.AdaptiveColor{Light: "", Dark: "69"}
	HelperHighlightTextDim = lipgloss.AdaptiveColor{Light: "", Dark: "75"}
	ErrorText              = lipgloss.AdaptiveColor{Light: "160", Dark: "169"}
	WarningText            = lipgloss.AdaptiveColor{Light: "208", Dark: "202"}
	MsgText                = lipgloss.AdaptiveColor{Light: "35", Dark: "41"}
	SelectionText          = lipgloss.AdaptiveColor{Light: "206", Dark: "207"} // peach
	BorderFg               = lipgloss.AdaptiveColor{Light: "63", Dark: "63"}
	BorderDimFg            = lipgloss.AdaptiveColor{Light: "249", Dark: "240"}

	Green = lipgloss.AdaptiveColor{Light: "29", Dark: "41"}
	Cyan  = lipgloss.AdaptiveColor{Light: "73", Dark: "122"}
	Skin  = lipgloss.AdaptiveColor{Light: "173", Dark: "222"}
	Pale  = lipgloss.AdaptiveColor{Light: "102", Dark: "187"}
)
