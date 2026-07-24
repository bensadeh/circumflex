// Command guess-report scores highlight.GuessLang against the real web.
//
// It walks the HN front page, fetches every linked article, and reports what
// the language guesser made of each code block. Pages that declare their own
// language are the point: a declaration is a free label, so the blocks that
// carry one measure the guesser's precision and name the languages worth
// detecting next, without anyone hand-labeling a corpus.
//
// Usage:
//
//	go run ./cmd/guess-report [-n 30] [-category topstories] [-samples 3]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bensadeh/circumflex/article"
	"github.com/bensadeh/circumflex/highlight"
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/services/firebase"
)

const (
	fetchConcurrency = 8
	overallTimeout   = 5 * time.Minute
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "guess-report: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		stories  = flag.Int("n", 30, "number of front-page stories to walk")
		category = flag.String("category", "topstories", "firebase story list")
		samples  = flag.Int("samples", 3, "unlabeled misses to print per shape")
		only     = flag.String("url", "", "score a single page instead of the front page")
		show     = flag.String("show", "", "print undeclared blocks detected as this language")
		missed   = flag.String("missed", "", "print declared blocks of this language the guesser missed")
		save     = flag.String("save", "", "write the collected blocks to this JSON corpus file")
		load     = flag.String("load", "", "comma-separated corpus files to rescore instead of fetching")
	)

	flag.Parse()

	var (
		blocks   []sample
		urlCount int
		failures int
		err      error
	)

	if *load != "" {
		blocks, err = loadCorpus(strings.Split(*load, ","))
	} else {
		blocks, urlCount, failures, err = fetchBlocks(*only, *category, *stories)
	}

	if err != nil {
		return err
	}

	if *save != "" {
		if err := saveCorpus(*save, blocks); err != nil {
			return err
		}
	}

	report(blocks, urlCount, failures, *samples)
	reportDetected(blocks, *show)
	reportMissed(blocks, *missed)

	return nil
}

func fetchBlocks(only, category string, stories int) ([]sample, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), overallTimeout)
	defer cancel()

	urls := []string{only}

	if only == "" {
		var err error

		urls, err = frontPageURLs(ctx, stories, category)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	fmt.Fprintf(os.Stderr, "walking %d articles…\n", len(urls))

	blocks, failures := collect(ctx, urls)

	return blocks, len(urls), failures, nil
}

// storedSample is a corpus file entry: the ground truth alone, never the
// guess — rescoring a saved corpus is the point of loading one.
type storedSample struct {
	Declared string `json:"declared,omitempty"`
	Host     string `json:"host"`
	Text     string `json:"text"`
}

func saveCorpus(path string, blocks []sample) error {
	stored := make([]storedSample, 0, len(blocks))
	for _, b := range blocks {
		stored = append(stored, storedSample{Declared: b.declared, Host: b.host, Text: b.text})
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

// loadCorpus rescores saved blocks with the current guesser, deduplicating
// across files — the same story sits on the top and best lists at once.
func loadCorpus(paths []string) ([]sample, error) {
	var out []sample

	seen := map[string]bool{}

	for _, path := range paths {
		data, err := os.ReadFile(strings.TrimSpace(path))
		if err != nil {
			return nil, err
		}

		var stored []storedSample
		if err := json.Unmarshal(data, &stored); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}

		for _, s := range stored {
			key := s.Host + "\x00" + s.Text
			if seen[key] {
				continue
			}

			seen[key] = true

			out = append(out, sample{
				declared: s.Declared,
				guessed:  highlight.GuessLang(s.Text),
				text:     s.Text,
				host:     s.Host,
			})
		}
	}

	return out, nil
}

// reportMissed prints the declared blocks of one language the guesser stayed
// silent on — the raw material for writing that language's detector.
func reportMissed(blocks []sample, lang string) {
	if lang == "" {
		return
	}

	fmt.Printf("DECLARED %s BLOCKS THE GUESSER MISSED\n", strings.ToUpper(lang))

	for _, b := range blocks {
		if b.guessed != "" || !strings.EqualFold(canonical(b.declared), canonical(lang)) {
			continue
		}

		fmt.Printf("  [%s]\n", b.host)
		fmt.Println(indent(firstLines(b.text, 15)))
	}
}

// reportDetected prints what a given detector claimed on unlabeled blocks —
// the one bucket with no ground truth, where a false positive would otherwise
// never surface.
func reportDetected(blocks []sample, lang string) {
	if lang == "" {
		return
	}

	fmt.Printf("BLOCKS DETECTED AS %s  (no declaration to check against)\n", strings.ToUpper(lang))

	for _, b := range blocks {
		if b.declared != "" || !strings.EqualFold(b.guessed, lang) {
			continue
		}

		fmt.Printf("  [%s]\n", b.host)
		fmt.Println(indent(firstLines(b.text, 5)))
	}
}

func frontPageURLs(ctx context.Context, count int, category string) ([]string, error) {
	stories, err := firebase.NewService().FetchItems(ctx, count, category)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", category, err)
	}

	var urls []string

	for _, s := range stories {
		// Ask HN and Show HN text posts point back at the item itself.
		if s.URL != "" && !isItemURL(s.URL) {
			urls = append(urls, s.URL)
		}
	}

	return urls, nil
}

func isItemURL(url string) bool {
	_, ok := hn.ParseItemURL(url)

	return ok
}

// sample is one code block seen in the wild.
type sample struct {
	declared string
	guessed  string
	text     string
	host     string
}

func collect(ctx context.Context, urls []string) ([]sample, int) {
	var (
		mu       sync.Mutex
		out      []sample
		failures int
		wg       sync.WaitGroup
	)

	sem := make(chan struct{}, fetchConcurrency)

	for _, url := range urls {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			parsed, err := article.Parse(ctx, url, false)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				failures++

				return
			}

			for _, b := range parsed.CodeBlocks() {
				out = append(out, sample{
					declared: b.Declared,
					guessed:  b.Guessed,
					text:     b.Text,
					host:     hostOf(url),
				})
			}
		})
	}

	wg.Wait()

	return out, failures
}

func hostOf(url string) string {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	host, _, _ := strings.Cut(trimmed, "/")

	return strings.TrimPrefix(host, "www.")
}

func report(blocks []sample, articles, failures, samples int) {
	var (
		labeled   []sample
		unlabeled []sample
	)

	for _, b := range blocks {
		if b.declared != "" {
			labeled = append(labeled, b)
		} else {
			unlabeled = append(unlabeled, b)
		}
	}

	fmt.Printf("\n%d articles walked, %d unreachable, %d code blocks\n",
		articles, failures, len(blocks))
	fmt.Printf("%d declared a language, %d did not\n\n", len(labeled), len(unlabeled))

	reportAccuracy(labeled)
	reportMisses(labeled)
	reportCoverage(unlabeled)
	reportUnlabeledMisses(unlabeled, samples)
}

// reportAccuracy scores the guesser against the pages' own declarations. A
// wrong answer here is a contract violation — the guesser is allowed to miss,
// never to name the wrong language.
func reportAccuracy(labeled []sample) {
	if len(labeled) == 0 {
		return
	}

	var agreed, missed int

	var wrong []sample

	agreedLangs := map[string]int{}

	for _, b := range labeled {
		switch {
		case canonical(b.guessed) == canonical(b.declared):
			agreed++
			agreedLangs[canonical(b.declared)]++

		case b.guessed == "":
			missed++

		default:
			wrong = append(wrong, b)
		}
	}

	fmt.Println("SCORED AGAINST THE PAGES' OWN DECLARATIONS")
	fmt.Printf("  agreed  %4d  (%4.1f%%)   %s\n", agreed, pct(agreed, len(labeled)),
		strings.Join(byCount(agreedLangs), " "))
	fmt.Printf("  missed  %4d  (%4.1f%%)   guesser stayed silent\n", missed, pct(missed, len(labeled)))
	fmt.Printf("  WRONG   %4d  (%4.1f%%)   contract violations\n\n", len(wrong), pct(len(wrong), len(labeled)))

	for _, b := range wrong {
		fmt.Printf("  %s declared %s, guessed %s\n", b.host, b.declared, b.guessed)
		fmt.Println(indent(firstLines(b.text, 4)))
	}
}

// reportMisses ranks the languages the guesser stayed silent on — the list of
// detectors worth writing next. Sites outrank blocks: one shell-heavy article
// contributes thirty blocks and says far less than the same language turning
// up on three unrelated sites.
func reportMisses(labeled []sample) {
	counts := map[string]int{}
	hosts := map[string]map[string]bool{}

	for _, b := range labeled {
		if b.guessed != "" {
			continue
		}

		// Group by lexer identity — sh, shell, and bash are one gap, not
		// three, and splitting them buries the language in the ranking.
		lang := canonical(b.declared)
		counts[lang]++

		if hosts[lang] == nil {
			hosts[lang] = map[string]bool{}
		}

		hosts[lang][b.host] = true
	}

	if len(counts) == 0 {
		return
	}

	langs := make([]string, 0, len(counts))
	for l := range counts {
		langs = append(langs, l)
	}

	sort.Slice(langs, func(i, j int) bool {
		if a, b := len(hosts[langs[i]]), len(hosts[langs[j]]); a != b {
			return a > b
		}

		if counts[langs[i]] != counts[langs[j]] {
			return counts[langs[i]] > counts[langs[j]]
		}

		return langs[i] < langs[j]
	})

	fmt.Println("MISSED LANGUAGES  (declared but not detected, ranked by sites)")

	for _, lang := range langs {
		fmt.Printf("  %-14s %d sites   %3d blocks\n", lang, len(hosts[lang]), counts[lang])
	}

	fmt.Println()
}

// canonical names a declaration by its lexer, falling back to the raw string
// so an unresolvable declaration still shows up rather than vanishing into an
// empty bucket. Bash and Bash Session fold together: pages label prompted
// transcripts bash, and choosing the session lexer for them is the right
// call, not a wrong language.
func canonical(lang string) string {
	if name := highlight.Canonical(lang); name != "" {
		if name == "Bash Session" {
			return "Bash"
		}

		return name
	}

	return lang
}

func reportCoverage(unlabeled []sample) {
	if len(unlabeled) == 0 {
		return
	}

	counts := map[string]int{}
	detected := 0

	for _, b := range unlabeled {
		if b.guessed == "" {
			continue
		}

		detected++
		counts[b.guessed]++
	}

	fmt.Printf("UNDECLARED BLOCKS: %d of %d detected (%.1f%%)\n",
		detected, len(unlabeled), pct(detected, len(unlabeled)))

	for _, lang := range byCount(counts) {
		fmt.Printf("  %-14s %3d\n", lang, counts[lang])
	}

	fmt.Println()
}

// reportUnlabeledMisses prints undetected blocks that no page labeled, the
// only bucket with no ground truth — they need an eyeball to sort code from
// terminal output and ASCII art, which are correct misses.
func reportUnlabeledMisses(unlabeled []sample, samples int) {
	if samples <= 0 {
		return
	}

	shown := 0

	fmt.Println("UNDETECTED AND UNDECLARED — eyeball these")

	for _, b := range unlabeled {
		if b.guessed != "" || strings.TrimSpace(b.text) == "" {
			continue
		}

		if shown >= samples {
			break
		}

		shown++

		fmt.Printf("  [%s]\n", b.host)
		fmt.Println(indent(firstLines(b.text, 6)))
	}
}

func byCount(counts map[string]int) []string {
	langs := make([]string, 0, len(counts))
	for l := range counts {
		langs = append(langs, l)
	}

	sort.Slice(langs, func(i, j int) bool {
		if counts[langs[i]] != counts[langs[j]] {
			return counts[langs[i]] > counts[langs[j]]
		}

		return langs[i] < langs[j]
	})

	return langs
}

func firstLines(text string, n int) string {
	lines := strings.Split(text, "\n")
	if len(lines) > n {
		lines = append(lines[:n], "…")
	}

	return strings.Join(lines, "\n")
}

func indent(text string) string {
	var b strings.Builder

	for line := range strings.SplitSeq(text, "\n") {
		b.WriteString("    │ " + line + "\n")
	}

	return b.String()
}

func pct(part, total int) float64 {
	if total == 0 {
		return 0
	}

	return float64(part) / float64(total) * 100
}
