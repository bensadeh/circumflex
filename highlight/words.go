package highlight

import (
	"slices"
	"strings"
)

// isRuby wants two of the shapes Ruby leaves everywhere: a def line without
// Python's colon or Elixir's do, block arguments between pipes, a bare end.
// Elixir is the near neighbor and lands at most one signal — its def always
// carries do, its puts is IO.puts mid-line, and its module attributes take
// no equals sign.
func isRuby(text string, lines []string) bool {
	return atLeastTwo(
		rubyDef(lines),
		anyLinePrefix(lines, "puts ", "require '", "require \"", "require_relative "),
		strings.Contains(text, " do |"),
		rubyIvarAssignment(lines),
		containsAny(text, []string{".each do", ".map do", ".new(", ".new }", "attr_accessor ", "attr_reader "}),
		anyLineIs(lines, "end"),
	)
}

func rubyDef(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "def ") && !strings.HasSuffix(t, ":") &&
			!strings.HasSuffix(t, " do") && !strings.Contains(t, " do |") {
			return true
		}
	}

	return false
}

func rubyIvarAssignment(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if !strings.HasPrefix(t, "@") {
			continue
		}

		name := strings.TrimLeft(t[1:], "abcdefghijklmnopqrstuvwxyz0123456789_")
		if len(name) < len(t)-1 && strings.HasPrefix(name, " = ") {
			return true
		}
	}

	return false
}

// isOCaml sits before the JavaScript-shaped detectors because OCaml's let
// bindings and arrow-bearing strings satisfy their signals. Its own markers
// don't appear elsewhere: (* comments, let rec, a lowercase type
// declaration, |-led match arms with ->.
func isOCaml(text string, lines []string) bool {
	return atLeastTwo(
		strings.Contains(text, "(*"),
		strings.Contains(text, "let rec "),
		anyLinePrefix(lines, "let () ="),
		ocamlTypeDecl(lines),
		ocamlMatchArm(lines),
		strings.Contains(text, " <- "),
		ocamlOpen(lines),
		anyLineSuffix(lines, ";;"),
		ocamlSpacedAnnotation(text),
	)
}

// ocamlSpacedAnnotation reports OCaml's spaced type colon — resolver :
// string — a lowercase identifier on both sides. A JavaScript ternary has
// the same silhouette, which is why this only ever corroborates: a block
// would need a second OCaml-only marker to resolve.
func ocamlSpacedAnnotation(text string) bool {
	for i := 1; i+3 < len(text); i++ {
		if text[i] == ' ' && text[i+1] == ':' && text[i+2] == ' ' &&
			(text[i-1] >= 'a' && text[i-1] <= 'z' || text[i-1] == '_') &&
			(text[i+3] >= 'a' && text[i+3] <= 'z') {
			return true
		}
	}

	return false
}

// ocamlTypeDecl reports type name = with a lowercase name — TypeScript and
// Rust capitalize their aliases, Haskell capitalizes its type constructors.
func ocamlTypeDecl(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		rest, ok := strings.CutPrefix(t, "type ")
		if !ok {
			continue
		}

		name, _, ok := strings.Cut(rest, " ")
		if ok && strings.Contains(rest, "=") && name != "" &&
			strings.TrimLeft(name, "abcdefghijklmnopqrstuvwxyz0123456789_'") == "" {
			return true
		}
	}

	return false
}

// ocamlMatchArm reports a line-leading | with -> on the same line. Rust and
// TypeScript arms arrow with =>, and Haskell's case arms carry no pipe.
func ocamlMatchArm(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "| ") && strings.Contains(t, " -> ") {
			return true
		}
	}

	return false
}

func ocamlOpen(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		rest, ok := strings.CutPrefix(t, "open ")
		if ok && rest != "" && rest[0] >= 'A' && rest[0] <= 'Z' {
			return true
		}
	}

	return false
}

// isJava leans on vocabulary no sibling shares: java. imports, System.out,
// the main signature. The access-modifier class line and an annotation are
// corroboration only — C# writes the former and TypeScript decorators mimic
// the latter, but neither language produces a second Java signal.
func isJava(text string, lines []string) bool {
	return atLeastTwo(
		anyLinePrefix(lines, "import java.", "import javax.", "import static java."),
		containsAny(text, []string{"System.out.", "System.err."}),
		strings.Contains(text, "public static void main"),
		javaAnnotationLine(lines),
		containsAny(text, []string{"new ArrayList<", "new HashMap<", "List<String>", "String[] "}),
		anyLinePrefix(lines, "public class ", "public interface ", "public final class ", "public abstract class "),
	)
}

// javaAnnotationLine reports a bare @Capitalized annotation on its own line.
// Python decorates in lowercase; C# brackets its attributes instead.
func javaAnnotationLine(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if len(t) > 1 && t[0] == '@' && t[1] >= 'A' && t[1] <= 'Z' &&
			!strings.Contains(t, " ") {
			return true
		}
	}

	return false
}

// isKotlin requires a fun line before anything else counts: fun is the one
// keyword its neighbors lack (Go says func, Scala says def), and the
// corroborating tokens — ?. and !! and val — all appear in other languages.
func isKotlin(text string, lines []string) bool {
	if !anyLinePrefix(lines, "fun ", "suspend fun ", "override fun ", "private fun ", "internal fun ") {
		return false
	}

	return atLeastTwo(
		anyLinePrefix(lines, "val ", "var "),
		containsAny(text, []string{"data class ", "companion object", "?.", "!!"}),
		strings.Contains(text, "println("),
		containsAny(text, []string{"import kotlin", "import android", "@Composable"}),
	)
}

// isSwift keys on framework imports, guard let, and property wrappers. A
// typed let/var declaration corroborates but is shared with TypeScript and
// Kotlin, both of which claim their blocks earlier in the table.
func isSwift(text string, lines []string) bool {
	return atLeastTwo(
		swiftImport(lines),
		strings.Contains(text, "guard let "),
		anyLinePrefix(lines, "extension ", "protocol ", "@objc", "@State", "@Published", "@main", "@IBOutlet", "@IBAction", "@MainActor"),
		strings.Contains(text, "func ") && strings.Contains(text, " -> "),
		swiftTypedBinding(lines),
	)
}

var swiftFrameworks = []string{
	"Foundation", "UIKit", "SwiftUI", "AppKit", "Cocoa", "Combine",
	"CoreData", "CoreGraphics", "Dispatch", "XCTest",
}

func swiftImport(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		rest, ok := strings.CutPrefix(t, "import ")
		if ok && slices.Contains(swiftFrameworks, rest) {
			return true
		}
	}

	return false
}

// swiftTypedBinding reports let/var name: Type with a capitalized type.
func swiftTypedBinding(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)

		rest, ok := strings.CutPrefix(t, "let ")
		if !ok {
			rest, ok = strings.CutPrefix(t, "var ")
		}

		if !ok {
			continue
		}

		_, typed, ok := strings.Cut(rest, ": ")
		if ok && typed != "" && typed[0] >= 'A' && typed[0] <= 'Z' {
			return true
		}
	}

	return false
}

func anyLineSuffix(lines []string, suffix string) bool {
	for _, l := range lines {
		if strings.HasSuffix(strings.TrimSpace(l), suffix) {
			return true
		}
	}

	return false
}
