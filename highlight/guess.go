package highlight

import (
	"encoding/json"
	"path"
	"slices"
	"strings"
)

// GuessLang names the language of an unlabeled code block from structural
// signals alone, or returns "" when nothing matches confidently.
// Statistical classifiers were evaluated and rejected: on snippet-sized input
// go-enry's Bayes classifier names a wrong language with full confidence
// (Go → JavaScript, JSON → Python, logs → Markdown), and a wrongly colored
// block is worse than a plain one. Every detector here must stay silent on
// prose, logs, and ASCII diagrams.
func GuessLang(text string) string {
	if lang := shebangLang(text); lang != "" {
		return lang
	}

	lines := strings.Split(text, "\n")

	// Lisp sits outside the table: the table pairs a predicate with one fixed
	// language, and a confirmed s-expression block still has to name its dialect.
	if isLisp(text, lines) {
		return lispDialect(text)
	}

	for _, d := range detectors {
		if d.match(text, lines) {
			return d.lang
		}
	}

	return ""
}

// The table runs specific before general: languages whose evidence is
// unmistakable (a diff header, valid JSON, a tab-indented recipe) go first,
// the C-shaped and JavaScript-shaped families follow, and YAML closes the
// list because a colon-keyed line is the weakest shape here — anything
// another detector can claim should never fall through to it.
var detectors = []struct {
	lang  string
	match func(text string, lines []string) bool
}{
	{"diff", isDiff},
	{"json", isJSON},
	{"markdown", isMarkdownDoc},
	{"toml", isTOML},
	{"nix", isNix},
	{"console", isShellSession},
	{"jsx", isComponentJSX},
	{"html", isHTML},
	{"xml", isXML},
	{"css", isCSS},
	{"objective-c", isObjectiveC},
	{"cpp", isCPP},
	{"c", isC},
	{"sql", isSQL},
	{"docker", isDockerfile},
	{"make", isMakefile},
	{"php", isPHP},
	{"bash", isShell},
	{"go", isGo},
	{"rust", isRust},
	{"python", isPython},
	{"ruby", isRuby},
	{"ocaml", isOCaml},
	{"typescript", isTypeScript},
	{"csharp", isCSharp},
	{"java", isJava},
	{"kotlin", isKotlin},
	{"swift", isSwift},
	{"jsx", isJSX},
	{"javascript", isJavaScript},
	{"yaml", isYAML},
}

// isJSX routes JavaScript with markup in expression position to the react
// lexer, which tags the JSX tags and attributes plain javascript leaves
// unstyled.
func isJSX(text string, lines []string) bool {
	return containsAny(text, []string{"(<", "=> <"}) && isJavaScript(text, lines)
}

// shebangLang maps an interpreter line to a lexer name; the strongest signal
// there is, when present.
func shebangLang(text string) string {
	first, _, _ := strings.Cut(text, "\n")
	if !strings.HasPrefix(first, "#!") {
		return ""
	}

	fields := strings.Fields(first[2:])
	if len(fields) == 0 {
		return ""
	}

	interpreter := path.Base(fields[0])
	if interpreter == "env" && len(fields) > 1 {
		interpreter = fields[1]
	}

	switch {
	case interpreter == "sh" || interpreter == "bash" || interpreter == "zsh" ||
		interpreter == "ksh" || interpreter == "dash":
		return "bash"

	case strings.HasPrefix(interpreter, "python"):
		return "python"

	case strings.HasPrefix(interpreter, "node"):
		return "javascript"

	case interpreter == "perl" || interpreter == "ruby" || interpreter == "awk" ||
		interpreter == "php" || interpreter == "lua":
		return interpreter

	default:
		return ""
	}
}

func isDiff(_ string, lines []string) bool {
	if anyLinePrefix(lines, "diff --git", "@@ -") ||
		(anyLinePrefix(lines, "--- ") && anyLinePrefix(lines, "+++ ")) {
		return true
	}

	// Headerless hunks: articles quote changed lines without the surrounding
	// machinery. Both signs must appear — a markdown list dashes every line,
	// but never mixes the two bullets — and three marked lines keep a stray
	// +1/-1 arithmetic pair from counting.
	var plus, minus int

	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "++"), strings.HasPrefix(l, "--"):
			// ++i; and --i; statements, SQL comments, em-dash rules.
		case strings.HasPrefix(l, "+"):
			plus++
		case strings.HasPrefix(l, "-"):
			minus++
		}
	}

	return plus >= 1 && minus >= 1 && plus+minus >= 3
}

// isMarkdownDoc reports a block that embeds fenced code blocks of its own —
// a whole document served inside one pre. No language's source can hold a
// pair of line-leading fences (they would have closed the page's own
// fencing), and chroma's markdown lexer highlights the embedded fences by
// their declared languages, which no single-language guess would.
func isMarkdownDoc(_ string, lines []string) bool {
	fences := 0

	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			fences++
		}
	}

	return fences >= 2
}

// isJSON accepts only objects and arrays: bare strings and numbers are valid
// JSON too, but nothing worth coloring.
func isJSON(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	return (strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) &&
		json.Valid([]byte(trimmed))
}

// isShellSession keys on the "$ " prompt followed by something command-shaped
// — commands are lowercase or paths and never close with another $, which
// keeps dollar-delimited LaTeX out. The zsh "% " prompt needs the stronger
// commandLine check: LaTeX comments open lines with % too, so a lowercase
// word after it proves nothing there.
func isShellSession(_ string, lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		if rest, ok := strings.CutPrefix(t, "% "); ok {
			if _, command := commandLine(rest); command {
				return true
			}
		}

		if !strings.HasPrefix(t, "$ ") || len(t) < 3 || strings.HasSuffix(t, "$") {
			continue
		}

		if c := t[2]; (c >= 'a' && c <= 'z') || c == '.' || c == '/' || c == '~' {
			return true
		}
	}

	return false
}

// isComponentJSX recognizes markup-first JSX: component tags are capitalized
// where html tags never are, and expression attributes use ={.
func isComponentJSX(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	return len(trimmed) >= 2 && trimmed[0] == '<' &&
		trimmed[1] >= 'A' && trimmed[1] <= 'Z' &&
		strings.Contains(trimmed, "={")
}

// isHTML claims markup only when the leading tag is an actual html element;
// everything else tag-shaped (pom files, Android layouts, SVG) is XML.
func isHTML(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype") {
		return true
	}

	if !markupShaped(trimmed) {
		return false
	}

	_, ok := htmlTags[strings.ToLower(leadingTag(trimmed))]

	return ok
}

func isXML(text string, _ []string) bool {
	trimmed := strings.TrimSpace(text)

	return strings.HasPrefix(trimmed, "<?xml") || markupShaped(trimmed)
}

// markupShaped wants a tag name after the bracket and either a closing tag
// or a self-closing element with an attribute — an angle-bracket autolink
// like <https://example.com/> has the slash but never the rest.
func markupShaped(trimmed string) bool {
	if len(trimmed) < 2 || trimmed[0] != '<' || !isASCIILetter(trimmed[1]) {
		return false
	}

	if leadingTag(trimmed) == "" {
		return false
	}

	return strings.Contains(trimmed, "</") ||
		(strings.Contains(trimmed, "/>") && strings.Contains(trimmed, "="))
}

// leadingTag returns the first tag's name, or "" when the bracket opens a
// URL scheme instead of an element.
func leadingTag(trimmed string) string {
	end := 1
	for end < len(trimmed) && (isASCIILetter(trimmed[end]) ||
		(trimmed[end] >= '0' && trimmed[end] <= '9') || trimmed[end] == '-') {
		end++
	}

	if end < len(trimmed) && trimmed[end] == ':' {
		return ""
	}

	return trimmed[1:end]
}

var htmlTags = map[string]struct{}{
	"html": {}, "head": {}, "body": {}, "div": {}, "span": {}, "p": {}, "a": {},
	"ul": {}, "ol": {}, "li": {}, "dl": {}, "dt": {}, "dd": {}, "table": {},
	"thead": {}, "tbody": {}, "tr": {}, "td": {}, "th": {}, "h1": {}, "h2": {},
	"h3": {}, "h4": {}, "h5": {}, "h6": {}, "img": {}, "br": {}, "hr": {},
	"form": {}, "input": {}, "button": {}, "label": {}, "select": {}, "option": {},
	"textarea": {}, "script": {}, "style": {}, "link": {}, "meta": {}, "title": {},
	"nav": {}, "header": {}, "footer": {}, "section": {}, "article": {}, "aside": {},
	"main": {}, "figure": {}, "figcaption": {}, "em": {}, "strong": {}, "b": {},
	"i": {}, "u": {}, "code": {}, "pre": {}, "blockquote": {}, "small": {},
	"sup": {}, "sub": {}, "iframe": {}, "video": {}, "audio": {}, "source": {},
	"canvas": {}, "details": {}, "summary": {},
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isObjectiveC(text string, lines []string) bool {
	return anyLinePrefix(lines, "#import ") ||
		containsAny(text, []string{"@interface", "@implementation", "@property"})
}

func isCPP(text string, lines []string) bool {
	return hasInclude(lines) &&
		containsAny(text, []string{
			"std::", "<iostream>", "template<", "namespace ",
			"class ", "public:", "private:", "protected:", "virtual ", "nullptr",
			"constexpr ", "static_cast<",
		})
}

// isC accepts an include line outright; without one it wants a typed
// declaration plus C's own library calls. The dot guard on the call keeps
// Java's System.out.printf and PHP object calls from counting, and the
// C-family languages that share the declaration shape declare their calls
// differently (Console.Write, cout, System.out).
func isC(text string, lines []string) bool {
	return hasInclude(lines) || (cDeclaration(lines) && cLibraryCall(text))
}

func cDeclaration(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasSuffix(t, ";") || !strings.Contains(t, "=") {
			continue
		}

		for _, p := range []string{"int ", "char ", "long ", "short ", "float ", "double ", "unsigned ", "size_t ", "uint", "int8", "int16", "int32", "int64"} {
			if strings.HasPrefix(t, p) {
				return true
			}
		}
	}

	return false
}

func cLibraryCall(text string) bool {
	for _, call := range []string{"printf(", "fprintf(", "scanf(", "malloc(", "calloc(", "free(", "memcpy(", "strlen(", "sizeof("} {
		for i := 0; ; {
			j := strings.Index(text[i:], call)
			if j < 0 {
				break
			}

			i += j

			if i == 0 || (text[i-1] != '.' && !isASCIILetter(text[i-1]) && text[i-1] != '_') {
				return true
			}

			i += len(call)
		}
	}

	return false
}

func hasInclude(lines []string) bool {
	return anyLinePrefix(lines, "#include <", "#include \"")
}

// isSQL requires a leading SQL verb plus corroborating clauses — one for an
// all-caps verb, two otherwise, since lowercase from/where also read as
// English prose.
func isSQL(_ string, lines []string) bool {
	// -- comment lines carry neither the verb nor honest clause evidence,
	// and a trailing comment ending in a period would read as a sentence to
	// the prose defenses below; every check runs on the statements alone.
	var code []string

	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" && !strings.HasPrefix(t, "--") {
			code = append(code, t)
		}
	}

	if len(code) == 0 {
		return false
	}

	firstLine := code[0]
	statements := strings.Join(code, "\n")

	first, _, _ := strings.Cut(firstLine, " ")
	verbs := []string{"select", "insert", "update", "delete", "create", "alter", "with", "explain"}

	if !slices.Contains(verbs, strings.ToLower(first)) {
		return false
	}

	// Prose defenses: statements don't end sentences with a period, don't
	// open with a colon clause ("UPDATE FAILED: …"), and aren't shouted
	// entirely in caps ("SELECT YOUR FAVORITE ITEMS FROM THE MENU");
	// lowercase SQL is additionally only trusted with a terminator.
	if strings.HasSuffix(statements, ".") ||
		strings.Contains(firstLine, ": ") ||
		statements == strings.ToUpper(statements) {
		return false
	}

	needed := 2
	if first == strings.ToUpper(first) {
		needed = 1
	} else if !strings.Contains(statements, ";") {
		return false
	}

	clauses := 0
	folded := " " + strings.ToLower(strings.Join(strings.Fields(statements), " ")) + " "

	for _, c := range []string{" from ", " join ", " where ", " values", " table ", " group by ", " order by "} {
		if strings.Contains(folded, c) {
			clauses++
		}
	}

	return clauses >= needed
}

// isPHP takes the tag as decisive; without it, a $var assignment plus echo
// or $this-> names the language JavaScript's => threshold would otherwise
// claim.
func isPHP(text string, lines []string) bool {
	if anyLinePrefix(lines, "<?php") {
		return true
	}

	return atLeastTwo(
		dollarAssignment(lines),
		anyLinePrefix(lines, "echo ", "print "),
		strings.Contains(text, "$this->"),
	)
}

func isDockerfile(_ string, lines []string) bool {
	return anyLinePrefix(lines, "FROM ") &&
		anyLinePrefix(lines, "RUN ", "COPY ", "CMD ", "ENTRYPOINT ", "WORKDIR ", "ARG ", "ENV ")
}

func isShell(text string, lines []string) bool {
	// A CI workflow embeds real shell under its run: keys, but the block is
	// YAML — the workflow chrome around the scripts decides.
	if anyLinePrefix(lines, "- uses: ", "uses: ", "runs-on:", "- name: ") ||
		anyLineIs(lines, "jobs:", "steps:", "on:") {
		return false
	}

	// A closed heredoc is decisive on its own, and has to be: the document's
	// body is another language's source, whose signals would otherwise win
	// the block for that language.
	if shellHeredoc(lines) {
		return true
	}

	// So is a block that is nothing but commands — install and build
	// instructions carry no shell syntax at all, just invocations.
	if commandBlock(lines) {
		return true
	}

	// PHP and Perl assign to $vars; shell assigns without the sigil, so a
	// dollar-assignment line disqualifies the block outright.
	if dollarAssignment(lines) {
		return false
	}

	return atLeastTwo(
		anyLineIs(lines, "do", "done", "fi", "then", "esac", "elif"),
		dollarExpansion(text),
		containsAny(text, []string{">/dev/null", "2>&1"}),
		containsAny(text, []string{" | grep", " | awk", " | sed", " | sort", " | xargs", " | head", " | tail", " | wc"}),
		shellForIn(lines),
		anyLinePrefix(lines, "echo ", "cd ", "sudo ", "mkdir ", "curl ", "set -"),
		shellExport(lines),
		quotedExpansion(text) || containsAny(text, []string{"$#", "$?"}),
	)
}

// commandBlock reports a block consisting solely of command invocations:
// every line is a command, a comment, a continuation, or blank, and at least
// one command is present. All-or-nothing is the safety: one line of prose or
// source anywhere rejects the whole block.
func commandBlock(lines []string) bool {
	// First pass: programs an unambiguous line already proved. A README that
	// demonstrates its own tool writes `pullrun run img --flag` once, then
	// bare `pullrun pull img` — the flagged line vouches for the name.
	proven := map[string]bool{}

	for _, l := range lines {
		if name, ok := commandLine(commandText(l)); ok {
			proven[name] = true
		}
	}

	commands := 0
	continued := false

	for _, l := range lines {
		t := commandText(l)
		if t == "" {
			continued = false

			continue
		}

		wasContinued := continued
		continued = strings.HasSuffix(t, "\\")

		if wasContinued || strings.HasPrefix(t, "#") {
			continue
		}

		name, ok := commandLine(t)
		if !ok && !proven[name] {
			return false
		}

		commands++
	}

	return commands > 0
}

// commandText trims a line and drops a trailing comment, so an annotated
// invocation (cmd arg  # 968 ms) reads as the invocation alone.
func commandText(l string) string {
	t := strings.TrimSpace(l)

	if body, _, ok := strings.Cut(t, " # "); ok {
		return strings.TrimSpace(body)
	}

	return t
}

// commandLine reports a line shaped like an invocation — a known program
// first, or an executable path, or an unknown program dense with option
// flags — and names the program either way, so a rejected line can still be
// vouched for. Command words that double as English (make, open, cat) demand
// a non-prose argument, so "make sure the server is running" stays a
// sentence.
func commandLine(t string) (string, bool) {
	if t == "" || strings.HasSuffix(t, ";") {
		return "", false
	}

	fields := strings.Fields(t)

	for len(fields) > 0 && (isEnvAssignment(fields[0]) ||
		fields[0] == "sudo" || fields[0] == "doas" || fields[0] == "env" ||
		fields[0] == "exec" || fields[0] == "nohup") {
		fields = fields[1:]
	}

	if len(fields) == 0 {
		return "", false
	}

	name := fields[0]
	args := fields[1:]

	// A colon marks a compiler diagnostic (./user.go:6:2: …), never an
	// executable's path.
	if (strings.HasPrefix(name, "./") || strings.HasPrefix(name, "~/")) &&
		!strings.Contains(name, ":") {
		return name, true
	}

	// A relative binary path — build/tinyrenderer scene.obj. URLs and
	// expression characters mean the slash was division or markup instead,
	// and one-letter segments (r/kl) mean it was never a path at all.
	if strings.Contains(name, "/") && !strings.Contains(name, "://") &&
		!strings.ContainsAny(name, "():<>{}[]=,;\"'`") && isASCIILetter(name[0]) &&
		pathSegmentsSubstantial(name) {
		return name, true
	}

	if _, ok := shellCommands[name]; ok {
		return name, !proseArgs(args)
	}

	if _, ok := ambiguousCommands[name]; ok {
		return name, commandArgs(args)
	}

	// pullrun run alpine:3.18 --cmd "echo" --attach: an unknown program is
	// still unmistakably invoked when option flags follow it — two long
	// options alone, or one flag among further non-prose arguments.
	if len(args) >= 2 && strings.TrimLeft(name, "abcdefghijklmnopqrstuvwxyz0123456789_-.") == "" {
		longOpts, punctuated := 0, 0

		for _, a := range args {
			if len(a) > 2 && strings.HasPrefix(a, "--") && isASCIILetter(a[2]) {
				longOpts++
			}

			if strings.ContainsAny(a, "-/.=~$:@") {
				punctuated++
			}
		}

		return name, longOpts >= 2 || (longOpts >= 1 && punctuated >= 2)
	}

	return name, false
}

// commandArgs reports arguments that read as flags, paths or targets rather
// than prose: any punctuation an English word lacks qualifies, as do the
// conventional bare subcommands (make install, swift test).
func commandArgs(args []string) bool {
	for _, a := range args {
		if strings.ContainsAny(a, "-/.=~$&|:\"'`@*") {
			return true
		}

		switch a {
		case "build", "test", "run", "install", "clean", "check", "fmt", "update", "upgrade":
			return true
		}
	}

	return false
}

func pathSegmentsSubstantial(name string) bool {
	for seg := range strings.SplitSeq(name, "/") {
		if len(seg) < 2 {
			return false
		}
	}

	return true
}

// proseArgs reports a tail of three or more plain English words — "npm is a
// package manager" — which no real invocation strings together without a
// flag, path or punctuation somewhere.
func proseArgs(args []string) bool {
	if len(args) < 3 {
		return false
	}

	for _, a := range args {
		if strings.ContainsAny(a, "-/.=~$:@\"'`|&") {
			return false
		}
	}

	return true
}

func isEnvAssignment(f string) bool {
	name, _, ok := strings.Cut(f, "=")

	return ok && name != "" && strings.TrimLeft(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_") == ""
}

// shellCommands are program names that essentially never open a line of any
// other language or of prose. Words that read as English verbs (make, open,
// find, touch) live in ambiguousCommands instead; `go` is absent outright —
// a goroutine launch opens Go lines with the same word.
var shellCommands = map[string]struct{}{
	"git": {}, "gh": {}, "docker": {}, "docker-compose": {}, "podman": {},
	"kubectl": {}, "helm": {}, "minikube": {}, "terraform": {}, "ansible": {},
	"npm": {}, "npx": {}, "pnpm": {}, "yarn": {}, "node": {}, "deno": {}, "bun": {},
	"pip": {}, "pip3": {}, "pipx": {}, "uv": {}, "uvx": {}, "poetry": {},
	"python": {}, "python3": {}, "gem": {}, "bundle": {}, "cargo": {}, "rustup": {},
	"cmake": {}, "ninja": {}, "gcc": {}, "g++": {}, "clang": {}, "javac": {},
	"mvn": {}, "gradle": {}, "dotnet": {}, "composer": {}, "mix": {}, "opam": {},
	"dune": {}, "zig": {}, "swiftc": {}, "xcodebuild": {}, "xcrun": {},
	"brew": {}, "apt": {}, "apt-get": {}, "dnf": {}, "yum": {}, "pacman": {},
	"apk": {}, "snap": {}, "flatpak": {}, "dpkg": {},
	"systemctl": {}, "journalctl": {}, "ssh": {}, "ssh-keygen": {}, "scp": {},
	"rsync": {}, "curl": {}, "wget": {}, "ping": {}, "dig": {},
	"tar": {}, "unzip": {}, "gzip": {}, "gunzip": {}, "zstd": {},
	"cd": {}, "ls": {}, "cp": {}, "mv": {}, "rm": {}, "mkdir": {}, "chmod": {},
	"chown": {}, "ln": {}, "pwd": {}, "whoami": {},
	"grep": {}, "rg": {}, "sed": {}, "awk": {}, "xargs": {}, "tee": {},
	"ffmpeg": {}, "jq": {}, "yq": {}, "sqlite3": {}, "psql": {}, "mysql": {},
	"redis-cli": {}, "aws": {}, "gcloud": {}, "az": {}, "fly": {}, "flyctl": {},
	"vim": {}, "nvim": {}, "tmux": {}, "htop": {}, "man": {}, "which": {},
	"echo": {}, "printf": {}, "source": {}, "chsh": {}, "ollama": {},
}

// ambiguousCommands double as ordinary English sentence openers, so a line
// they start must also carry a command-shaped argument to count.
var ambiguousCommands = map[string]struct{}{
	"make": {}, "open": {}, "find": {}, "touch": {}, "sort": {}, "kill": {},
	"cat": {}, "less": {}, "head": {}, "tail": {}, "cut": {}, "code": {},
	"time": {}, "watch": {}, "export": {}, "swift": {}, "java": {}, "ruby": {},
	"perl": {}, "php": {}, "top": {}, "free": {}, "date": {}, "clear": {},
}

// shellHeredoc reports a << WORD redirection whose all-caps delimiter later
// closes on its own line. PHP's heredoc spells <<<, Ruby's leans on <<~ and
// ends lines with = or (, Perl's ends the line with ;, and a C++ stream or
// bit shift never leaves the shifted name alone on a line — the closing line
// is what makes the pair a document.
func shellHeredoc(lines []string) bool {
	for i, l := range lines {
		before, after, ok := strings.Cut(l, "<<")
		if !ok || strings.Contains(l, "<<<") || strings.Contains(l, "<<~") ||
			strings.HasSuffix(strings.TrimSpace(l), ";") {
			continue
		}

		// Shell opens a heredoc after a command; an = before the operator
		// means HCL, Terraform or Ruby assigning one instead.
		if strings.HasSuffix(strings.TrimSpace(before), "=") {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(after, "-"))
		if len(fields) == 0 {
			continue
		}

		delim := strings.Trim(fields[0], `'"`)
		if delim == "" || strings.TrimLeft(delim, "ABCDEFGHIJKLMNOPQRSTUVWXYZ_") != "" {
			continue
		}

		for _, later := range lines[i+1:] {
			if strings.TrimSpace(later) == delim {
				return true
			}
		}
	}

	return false
}

// dollarExpansion counts $(cmd) substitutions and ${var} expansions —
// skipping Make's uppercase $(VAR) form and templating's ${{ doubles.
func dollarExpansion(text string) bool {
	for i := 0; i+2 < len(text); i++ {
		if text[i] != '$' {
			continue
		}

		c := text[i+2]

		switch text[i+1] {
		case '(':
			if c >= 'a' && c <= 'z' {
				return true
			}

		case '{':
			if c != '{' && (isASCIILetter(c) || c == '_' || (c >= '0' && c <= '9')) {
				return true
			}
		}
	}

	return false
}

// quotedExpansion looks for "$name — one "${ occurrence must not also feed
// this signal, or quoted template interpolation counts a single construct
// twice.
func quotedExpansion(text string) bool {
	for i := 0; i+2 < len(text); i++ {
		if text[i] == '"' && text[i+1] == '$' &&
			(isASCIILetter(text[i+2]) || text[i+2] == '_') {
			return true
		}
	}

	return false
}

func dollarAssignment(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasPrefix(t, "$") {
			continue
		}

		name := strings.TrimLeft(t[1:], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
		if len(name) < len(t)-1 && strings.HasPrefix(strings.TrimSpace(name), "=") {
			return true
		}
	}

	return false
}

// shellExport separates shell's export VAR=value from JavaScript's export
// declarations, which always put spaces around any assignment.
func shellExport(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "export ") && strings.Contains(t, "=") && !strings.Contains(t, " = ") {
			return true
		}
	}

	return false
}

func isGo(text string, lines []string) bool {
	return atLeastTwo(
		strings.Contains(text, " := "),
		anyLinePrefix(lines, "func "),
		anyLinePrefix(lines, "package "),
		containsAny(text, []string{"fmt.", "err != nil"}),
		anyLineIs(lines, "import ("),
	)
}

func isRust(text string, lines []string) bool {
	return atLeastTwo(
		anyLinePrefix(lines, "fn ", "pub fn "),
		containsAny(text, []string{"let mut ", "&mut ", "&str"}),
		unspacedPathSep(text),
		containsAny(text, []string{"println!", "#[derive", ".unwrap()", "?;", "vec!"}),
		anyLinePrefix(lines, "let "),
		rustMatchArm(text, lines),
		rustLifetime(text),
	)
}

// rustLifetime reports a lifetime annotation in generic or reference
// position (<'a>, &'a). It anchors on < or & immediately before the
// apostrophe, so a bare char literal never counts, and a trailing
// apostrophe means the quote closed a char literal instead — keeping C's
// c<'a' comparison and &'x' char-address out.
func rustLifetime(text string) bool {
	for i := 0; i+1 < len(text); i++ {
		if (text[i] != '<' && text[i] != '&') || text[i+1] != '\'' {
			continue
		}

		j := i + 2
		for j < len(text) && (isASCIILetter(text[j]) || text[j] == '_' ||
			(text[j] >= '0' && text[j] <= '9')) {
			j++
		}

		if j > i+2 && (j >= len(text) || text[j] != '\'') {
			return true
		}
	}

	return false
}

// unspacedPathSep reports Rust's Type::path form, an identifier character on
// both sides. Haskell's type-signature :: is always spaced, and CSS's
// ::before pseudo-elements follow a selector or nothing.
func unspacedPathSep(text string) bool {
	for i := 0; ; i += 2 {
		j := strings.Index(text[i:], "::")
		if j < 0 {
			return false
		}

		i += j

		before := i > 0 && (isASCIILetter(text[i-1]) || text[i-1] == '_' ||
			(text[i-1] >= '0' && text[i-1] <= '9') || text[i-1] == '>')
		after := i+2 < len(text) && (isASCIILetter(text[i+2]) || text[i+2] == '_' || text[i+2] == '<')

		if before && after {
			return true
		}
	}
}

// rustMatchArm pairs a match line opening a block with arrow arms, so match
// expressions read as Rust before the => alone can read as JavaScript.
func rustMatchArm(text string, lines []string) bool {
	if !strings.Contains(text, "=>") {
		return false
	}

	for _, l := range lines {
		t := strings.TrimSpace(l)
		if (strings.HasPrefix(t, "match ") || strings.Contains(t, " match ")) && strings.HasSuffix(t, "{") {
			return true
		}
	}

	return false
}

func isPython(text string, lines []string) bool {
	return atLeastTwo(
		pythonDef(lines),
		pythonImport(lines),
		pythonForIn(lines),
		containsAny(text, []string{"self.", "__init__", "__name__"}),
		containsAny(text, []string{"elif ", " is None", "f\""}),
		strings.Contains(text, "print("),
		pythonColonBlock(lines),
		bareAssignment(lines),
	)
}

// pythonColonBlock reports a compound-statement header closed by the colon —
// if x:, while x:, try:. The for-in header stays with pythonForIn so a single
// line never earns two signals.
func pythonColonBlock(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasSuffix(t, ":") {
			continue
		}

		if t == "try:" || t == "else:" || t == "finally:" ||
			anyPrefix(t, "if ", "elif ", "while ", "with ", "except ", "except:") {
			return true
		}
	}

	return false
}

// bareAssignment reports name = value with no declaration keyword and no
// terminator: Python's plain binding. The semicolon exclusion keeps C, Java
// and Nix out; the space around = keeps shell out; snake-case-only names
// keep TOML's dashed keys out. Ruby binds identically, so this stays one
// corroborating signal among several.
func bareAssignment(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		name, rest, ok := strings.Cut(t, " = ")
		if !ok || name == "" || rest == "" || strings.HasSuffix(t, ";") {
			continue
		}

		if strings.TrimLeft(name, "abcdefghijklmnopqrstuvwxyz0123456789_") == "" {
			return true
		}
	}

	return false
}

func anyPrefix(t string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}

	return false
}

// isCSharp keys on markers the rest of the C family doesn't write: a using
// directive (C++ says using namespace or aliases with =, Java says import),
// $"" interpolation, and bracketed attributes where Rust writes #[] and Java
// writes @. Access modifiers are shared with Java, so they corroborate but
// never carry a block alone.
func isCSharp(text string, lines []string) bool {
	// A terminated var statement and a LINQ operator are each conventions the
	// languages that could otherwise claim the block break: Go declares vars
	// without the semicolon, and Java and JavaScript name methods in
	// camelCase. Neither half is trusted alone — Java also writes var.
	if csharpVar(lines) && linqCall(text) {
		return true
	}

	return atLeastTwo(
		csharpUsing(lines),
		csharpInterpolation(text),
		csharpAttribute(lines),
		anyLinePrefix(lines, "namespace "),
		containsAny(text, []string{"Console.Write", "nameof(", "async Task", "string[] args", "IEnumerable<"}),
		anyLinePrefix(lines, "public ", "private ", "protected ", "internal "),
	)
}

// csharpUsing accepts the directive form only: C++'s using either names a
// namespace, qualifies with ::, or assigns an alias, and all three shapes are
// lowercase or carry an =.
func csharpUsing(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		rest, ok := strings.CutPrefix(t, "using ")
		if !ok || !strings.HasSuffix(t, ";") ||
			strings.Contains(t, "=") || strings.Contains(t, "::") {
			continue
		}

		rest = strings.TrimPrefix(rest, "static ")
		if rest != "" && rest[0] >= 'A' && rest[0] <= 'Z' {
			return true
		}
	}

	return false
}

// csharpInterpolation finds an interpolated string opener. A "$" literal
// holds the same two bytes in the opposite order of nesting, so a preceding
// quote or backslash disqualifies the match.
func csharpInterpolation(text string) bool {
	for i := 0; i+1 < len(text); i++ {
		if text[i] != '$' || text[i+1] != '"' {
			continue
		}

		if i == 0 || (text[i-1] != '"' && text[i-1] != '\\') {
			return true
		}
	}

	return false
}

// csharpAttribute matches an attribute on its own line: Rust brackets its
// derives behind #, Java writes @Override, and a markdown link closes on ).
func csharpAttribute(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if len(t) < 3 || t[0] != '[' || !strings.HasSuffix(t, "]") {
			continue
		}

		if t[1] >= 'A' && t[1] <= 'Z' {
			return true
		}
	}

	return false
}

func csharpVar(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "var ") && strings.HasSuffix(t, ";") {
			return true
		}
	}

	return false
}

// linqCall names the standard query operators, kept a closed set: an open
// PascalCase-call rule would swallow Go's fmt.Println and every Windows API
// binding written in another language.
func linqCall(text string) bool {
	return containsAny(text, []string{
		".Where(", ".Select(", ".SelectMany(", ".OrderBy(", ".OrderByDescending(",
		".ThenBy(", ".GroupBy(", ".FirstOrDefault(", ".SingleOrDefault(",
		".ToList(", ".ToArray(", ".ToDictionary(", ".Aggregate(", ".Distinct(",
	})
}

// isTypeScript checks before isJavaScript so annotated code reaches the lexer
// that colors the annotations. A full JavaScript shape counts as one signal,
// so a lone `: string` in prose or YAML (`type: string`) never carries a
// block, and neither does JavaScript alone.
func isTypeScript(text string, lines []string) bool {
	return atLeastTwo(
		tsPrimitiveAnnotation(text),
		tsAliasOrInterface(lines),
		tsOptionalMember(text),
		tsTypedMembers(text, lines),
		isJavaScript(text, lines),
	)
}

// tsPrimitiveAnnotation reports `: ` followed by one of TypeScript's own
// primitive names. The set shares no word with Python's builtins (str/int/
// bool against string/number/boolean), so annotated Python never matches;
// requiring an identifier character before the colon rules out OCaml's
// spaced `x : string` style.
func tsPrimitiveAnnotation(text string) bool {
	for i := 0; ; {
		j := strings.Index(text[i:], ": ")
		if j < 0 {
			return false
		}

		i += j + 2

		if i < 3 {
			continue
		}

		if before := text[i-3]; !isASCIILetter(before) && before != '_' && before != ')' &&
			(before < '0' || before > '9') {
			continue
		}

		rest := text[i:]
		for _, prim := range []string{"string", "number", "boolean", "void", "any", "unknown", "never"} {
			after, ok := strings.CutPrefix(rest, prim)
			if !ok {
				continue
			}

			after = strings.TrimPrefix(after, "[]")
			if after == "" || !isASCIILetter(after[0]) {
				return true
			}
		}
	}
}

// tsTypedMembers reports two member lines annotated with a named type — the
// interface-body shape (id: ItemId). The value side must read as a type, a
// capitalized name or quoted-literal union, so an object literal's values
// (numbers, strings, arrow functions) never qualify; the brace requirement
// keeps OpenAPI-style YAML, which writes the same key shapes, out.
func tsTypedMembers(text string, lines []string) bool {
	if !strings.Contains(text, "{") {
		return false
	}

	count := 0

	for _, l := range lines {
		t := strings.TrimSpace(l)
		t = strings.TrimSuffix(t, ",")
		t = strings.TrimSuffix(t, ";")

		name, typed, ok := strings.Cut(t, ": ")
		if !ok || typed == "" {
			continue
		}

		name = strings.TrimSuffix(name, "?")
		if name == "" || strings.TrimLeft(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_") != "" {
			continue
		}

		if (typed[0] >= 'A' && typed[0] <= 'Z') ||
			(typed[0] == '\'' && strings.Contains(typed, " | ")) {
			count++
		}
	}

	return count >= 2
}

// tsAliasOrInterface reports a type alias or interface declaration line. Go
// aliases types with the same `type X = ` shape, but a Go block earns its two
// signals first — the Go detector runs earlier.
func tsAliasOrInterface(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		t = strings.TrimPrefix(t, "export ")

		if strings.HasPrefix(t, "interface ") && strings.HasSuffix(t, "{") {
			return true
		}

		if rest, ok := strings.CutPrefix(t, "type "); ok &&
			strings.Contains(rest, "= ") && rest != "" && rest[0] >= 'A' && rest[0] <= 'Z' {
			return true
		}
	}

	return false
}

// tsOptionalMember reports `?: ` glued to the name it follows — an optional
// property or parameter. Kotlin's elvis and GNU C's a ?: b both put a space
// before the operator, so the preceding identifier character is what counts.
func tsOptionalMember(text string) bool {
	for i := 1; i+2 < len(text); i++ {
		if text[i] == '?' && text[i+1] == ':' && text[i+2] == ' ' &&
			(isASCIILetter(text[i-1]) || text[i-1] == '_' || (text[i-1] >= '0' && text[i-1] <= '9')) {
			return true
		}
	}

	return false
}

func isJavaScript(text string, lines []string) bool {
	// OCaml satisfies these signals from the outside — let-opened lines
	// everywhere and => inside format strings — so its unmistakable markers
	// disqualify the block even when the OCaml detector stayed silent.
	if containsAny(text, []string{"(*", "let rec "}) || anyLineSuffix(lines, ";;") {
		return false
	}

	return atLeastTwo(
		strings.Contains(text, "=>"),
		anyLinePrefix(lines, "const ", "let ", "var "),
		containsAny(text, []string{"console.log", "===", "!=="}),
		containsAny(text, []string{"function ", "await ", "async "}),
		strings.Contains(text, "`${"),
		anyLinePrefix(lines, "export function", "export default", "export const",
			"export class", "export async", "export {", "import {"),
		containsAny(text, []string{"document.", "window.", ".addEventListener"}),
		containsAny(text, []string{"for (let ", "for (const ", "for (var "}),
		containsAny(text, []string{"(<", "=> <"}), // JSX flowing into an expression
	)
}

// pythonDef requires the trailing colon that separates a Python definition
// from Ruby's def or Go's func.
func pythonDef(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if (strings.HasPrefix(t, "def ") || strings.HasPrefix(t, "class ")) && strings.HasSuffix(t, ":") {
			return true
		}
	}

	return false
}

// pythonImport rejects lines with quotes (JavaScript's import-from-module
// form) and capitalized module paths — Python modules are lowercase where
// Swift imports Foundation and Java/Kotlin import java.util.Random.
func pythonImport(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.ContainsAny(t, `"'`) {
			continue
		}

		var module string

		if rest, ok := strings.CutPrefix(t, "import "); ok {
			module, _, _ = strings.Cut(rest, " ")
		} else if rest, ok := strings.CutPrefix(t, "from "); ok && strings.Contains(t, " import ") {
			module, _, _ = strings.Cut(rest, " ")
		} else {
			continue
		}

		module = strings.TrimSuffix(module, ",")
		if module != "" && module == strings.ToLower(module) {
			return true
		}
	}

	return false
}

// pythonForIn recognizes a colon-terminated for-in header — Python's
// for x in y: form. Shell's for-in never closes with a colon (shellForIn
// requires its absence) and JavaScript and C parenthesize the header, so the
// bare header plus the trailing colon is a signal independent of the def line.
func pythonForIn(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "for ") && !strings.HasPrefix(t, "for (") &&
			strings.Contains(t, " in ") && strings.HasSuffix(t, ":") {
			return true
		}
	}

	return false
}

// shellForIn skips parenthesized loops: shell's for-in never has them,
// JavaScript's for (const key in obj) always does.
func shellForIn(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "for ") && !strings.HasPrefix(t, "for (") &&
			strings.Contains(t, " in ") && !strings.HasSuffix(t, ":") {
			return true
		}
	}

	return false
}

func atLeastTwo(signals ...bool) bool {
	hits := 0

	for _, s := range signals {
		if s {
			hits++
		}
	}

	return hits >= 2
}

func anyLinePrefix(lines []string, prefixes ...string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		for _, p := range prefixes {
			if strings.HasPrefix(t, p) {
				return true
			}
		}
	}

	return false
}

func anyLineIs(lines []string, words ...string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		if slices.Contains(words, t) {
			return true
		}
	}

	return false
}

func containsAny(text string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(text, target) {
			return true
		}
	}

	return false
}
