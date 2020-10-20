package model

import "strings"

func FormatShowAndTell(title string) string {
	title = strings.ReplaceAll(title, "Show HN:", reverse("Show HN:"))
	title = strings.ReplaceAll(title, "Ask HN:", reverse("Ask HN:"))
	title = strings.ReplaceAll(title, "Tell HN:", reverse("Tell HN:"))
	title = strings.ReplaceAll(title, "Launch HN:", reverse("Launch HN:"))
	return title
}

func reverse(text string) string {
	return "[::r]" + text + "[-:-:-]"
}

func FormatYCStartups(title string) string {
	title = strings.ReplaceAll(title, "(YC S05)", orange("(YC S05)"))
	title = strings.ReplaceAll(title, "(YC W05)", orange("(YC W05)"))
	title = strings.ReplaceAll(title, "(YC S06)", orange("(YC S06)"))
	title = strings.ReplaceAll(title, "(YC W06)", orange("(YC W06)"))
	title = strings.ReplaceAll(title, "(YC S07)", orange("(YC S07)"))
	title = strings.ReplaceAll(title, "(YC W07)", orange("(YC W07)"))
	title = strings.ReplaceAll(title, "(YC S08)", orange("(YC S08)"))
	title = strings.ReplaceAll(title, "(YC W08)", orange("(YC W08)"))
	title = strings.ReplaceAll(title, "(YC S09)", orange("(YC S09)"))
	title = strings.ReplaceAll(title, "(YC W09)", orange("(YC W09)"))
	title = strings.ReplaceAll(title, "(YC S10)", orange("(YC S10)"))
	title = strings.ReplaceAll(title, "(YC W10)", orange("(YC W10)"))
	title = strings.ReplaceAll(title, "(YC S11)", orange("(YC S11)"))
	title = strings.ReplaceAll(title, "(YC W11)", orange("(YC W11)"))
	title = strings.ReplaceAll(title, "(YC S12)", orange("(YC S12)"))
	title = strings.ReplaceAll(title, "(YC W12)", orange("(YC W12)"))
	title = strings.ReplaceAll(title, "(YC S13)", orange("(YC S13)"))
	title = strings.ReplaceAll(title, "(YC W13)", orange("(YC W13)"))
	title = strings.ReplaceAll(title, "(YC S14)", orange("(YC S14)"))
	title = strings.ReplaceAll(title, "(YC W14)", orange("(YC W14)"))
	title = strings.ReplaceAll(title, "(YC S15)", orange("(YC S15)"))
	title = strings.ReplaceAll(title, "(YC W15)", orange("(YC W15)"))
	title = strings.ReplaceAll(title, "(YC S16)", orange("(YC S16)"))
	title = strings.ReplaceAll(title, "(YC W16)", orange("(YC W16)"))
	title = strings.ReplaceAll(title, "(YC S17)", orange("(YC S17)"))
	title = strings.ReplaceAll(title, "(YC W17)", orange("(YC W17)"))
	title = strings.ReplaceAll(title, "(YC S18)", orange("(YC S18)"))
	title = strings.ReplaceAll(title, "(YC W18)", orange("(YC W18)"))
	title = strings.ReplaceAll(title, "(YC S19)", orange("(YC S19)"))
	title = strings.ReplaceAll(title, "(YC W19)", orange("(YC W19)"))
	title = strings.ReplaceAll(title, "(YC S20)", orange("(YC S20)"))
	title = strings.ReplaceAll(title, "(YC W20)", orange("(YC W20)"))
	return title
}

func orange(text string) string {
	return "[orange]" + text + "[-:-:-]"
}

