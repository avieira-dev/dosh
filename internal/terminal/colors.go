package terminal

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	RuleFg         = "\033[38;5;238m"
	GutterFg       = "\033[38;5;240m"
	GutterActiveFg = "\033[38;5;252m"
	TitleFg        = "\033[38;5;255m"
	StatusFg       = "\033[38;5;244m"
	StatusKeyFg    = "\033[38;5;250m"
	CurrentLineBg  = "\033[48;5;235m"
	MatchBg        = "\033[48;5;250m"
	MatchFg        = "\033[38;5;233m"
	SelectionBg    = "\033[48;5;238m"
	SelectionFg    = "\033[38;5;255m"
)
