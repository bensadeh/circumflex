package theme

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/BurntSushi/toml"
)

type Theme struct {
	App      AppColors      `toml:"app"`
	Headline HeadlineColors `toml:"headline"`
	Comment  CommentColors  `toml:"comment"`
	Meta     MetaColors     `toml:"meta"`
	Reader   ReaderColors   `toml:"reader"`
	Header   HeaderColors   `toml:"header"`
	Footer   FooterColors   `toml:"footer"`
	Help     HelpColors     `toml:"help"`
	Indent   IndentColors   `toml:"indent"`
}

type AppColors struct {
	Primary   string `toml:"primary"`
	Secondary string `toml:"secondary"`
	Tertiary  string `toml:"tertiary"`
}

type HeadlineColors struct {
	AskHN    string `toml:"ask_hn"`
	ShowHN   string `toml:"show_hn"`
	TellHN   string `toml:"tell_hn"`
	ThankHN  string `toml:"thank_hn"`
	LaunchHN string `toml:"launch_hn"`
	YCLabel  string `toml:"yc_label"`
	Year     string `toml:"year"`
	Audio    string `toml:"audio"`
	Video    string `toml:"video"`
	PDF      string `toml:"pdf"`
}

type CommentColors struct {
	URL          string `toml:"url"`
	Mention      string `toml:"mention"`
	Mod          string `toml:"mod"`
	Variable     string `toml:"variable"`
	Backtick     string `toml:"backtick"`
	OP           string `toml:"op"`
	GP           string `toml:"gp"`
	NewIndicator string `toml:"new_indicator"`
}

type MetaColors struct {
	Author      string `toml:"author"`
	Comments    string `toml:"comments"`
	Score       string `toml:"score"`
	ID          string `toml:"id"`
	NewComments string `toml:"new_comments"`
	URL         string `toml:"url"`
	ReaderMode  string `toml:"reader_mode"`
}

type ReaderColors struct {
	H1         string `toml:"h1"`
	H2         string `toml:"h2"`
	H3         string `toml:"h3"`
	H4         string `toml:"h4"`
	H5         string `toml:"h5"`
	H6         string `toml:"h6"`
	Image      string `toml:"image"`
	BBCImage   string `toml:"bbc_image"`
	BBCCaption string `toml:"bbc_caption"`
}

type HeaderColors struct {
	C         string `toml:"c"`
	L         string `toml:"l"`
	X         string `toml:"x"`
	Favorites string `toml:"favorites"`
}

type HelpColors struct {
	MainMenu       string `toml:"main_menu"`
	CommentSection string `toml:"comment_section"`
	Legend         string `toml:"legend"`
}

type FooterColors struct {
	ReadMode     string `toml:"read_mode"`
	NavigateMode string `toml:"navigate_mode"`
}

type IndentColors struct {
	Cycle []string `toml:"cycle"`
}

func Default() *Theme {
	return &Theme{
		App: AppColors{
			Primary:   "magenta",
			Secondary: "yellow",
			Tertiary:  "blue",
		},
		Headline: HeadlineColors{
			AskHN:    "blue",
			ShowHN:   "red",
			TellHN:   "magenta",
			ThankHN:  "cyan",
			LaunchHN: "green",
			YCLabel:  "yellow",
			Year:     "magenta",
			Audio:    "cyan",
			Video:    "cyan",
			PDF:      "cyan",
		},
		Comment: CommentColors{
			URL:          "blue",
			Mention:      "yellow",
			Mod:          "green",
			Variable:     "cyan",
			Backtick:     "magenta",
			OP:           "red",
			GP:           "magenta",
			NewIndicator: "cyan",
		},
		Meta: MetaColors{
			Author:      "red",
			Comments:    "magenta",
			Score:       "yellow",
			ID:          "green",
			NewComments: "cyan",
			URL:         "blue",
			ReaderMode:  "green",
		},
		Reader: ReaderColors{
			H1:         "blue",
			H2:         "red",
			H3:         "magenta",
			H4:         "yellow",
			H5:         "green",
			H6:         "white",
			Image:      "red",
			BBCImage:   "cyan",
			BBCCaption: "yellow",
		},
		Header: HeaderColors{
			C:         "magenta",
			L:         "yellow",
			X:         "blue",
			Favorites: "219",
		},
		Footer: FooterColors{
			ReadMode:     "magenta",
			NavigateMode: "yellow",
		},
		Help: HelpColors{
			MainMenu:       "magenta",
			CommentSection: "yellow",
			Legend:         "blue",
		},
		Indent: IndentColors{
			Cycle: []string{
				"red", "yellow", "green", "cyan", "blue", "magenta",
				"bright_red", "bright_yellow", "bright_green", "bright_cyan", "bright_blue", "bright_magenta",
			},
		},
	}
}

func configHeader() string {
	return "# circumflex theme configuration\n" +
		"#\n" +
		"# Color values can be:\n" +
		"#   - Named: \"red\", \"blue\", \"green\", \"yellow\", \"magenta\", \"cyan\", \"white\", \"black\"\n" +
		"#   - Bright: \"bright_red\", \"bright_blue\", \"bright_green\", etc.\n" +
		"#   - ANSI 256: \"219\", \"33\", \"196\"\n" +
		"#   - Hex: \"#ff5500\", \"#1a1a2e\"\n\n"
}

func WriteDefaultConfig() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(configDir, "circumflex")
	path := filepath.Join(dir, "theme.toml")

	if _, statErr := os.Stat(path); statErr == nil {
		return "", fmt.Errorf("config already exists at %s", path)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(configHeader()); err != nil {
		return "", err
	}

	if err := toml.NewEncoder(f).Encode(Default()); err != nil {
		return "", err
	}

	return path, nil
}

func Load() (*Theme, error) {
	t := Default()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return t, err
	}

	path := filepath.Join(configDir, "circumflex", "theme.toml")

	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return t, nil
	}

	_, err = toml.DecodeFile(path, t)
	if err != nil {
		return Default(), err
	}

	return t, nil
}

var namedColors = map[string]color.Color{
	"red":            lipgloss.Red,
	"blue":           lipgloss.Blue,
	"green":          lipgloss.Green,
	"yellow":         lipgloss.Yellow,
	"magenta":        lipgloss.Magenta,
	"cyan":           lipgloss.Cyan,
	"white":          lipgloss.White,
	"black":          lipgloss.Black,
	"bright_red":     lipgloss.BrightRed,
	"bright_blue":    lipgloss.BrightBlue,
	"bright_green":   lipgloss.BrightGreen,
	"bright_yellow":  lipgloss.BrightYellow,
	"bright_cyan":    lipgloss.BrightCyan,
	"bright_magenta": lipgloss.BrightMagenta,
	"bright_white":   lipgloss.BrightWhite,
	"bright_black":   lipgloss.BrightBlack,
}

func ParseColor(s string) color.Color {
	s = strings.TrimSpace(s)

	if c, ok := namedColors[s]; ok {
		return c
	}

	if strings.HasPrefix(s, "#") {
		return lipgloss.Color(s)
	}

	if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= 255 {
		return lipgloss.ANSIColor(n)
	}

	return lipgloss.NoColor{}
}
