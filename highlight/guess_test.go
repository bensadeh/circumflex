package highlight

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The corpus contract: detectors may miss (a plain block is fine) but must
// never name a wrong language, so every non-code sample pins "".
func TestGuessLang(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{
			"js es module",
			"export function performSomeAction(data) {\n   return {\n      type: SOME_ACTION,\n      data\n   }\n}",
			"javascript",
		},
		{
			"js dom access",
			"// These two containers are siblings in the DOM\n" +
				"const appContainer = document.getElementById('app-container');\n" +
				"const modalContainer = document.getElementById('app-container');",
			"javascript",
		},
		{
			"shell exports",
			"export EDITOR=vim\nexport PATH=\"$PATH:/opt/bin\"",
			"bash",
		},
		{
			"jsx",
			"const array = []\nfor (let item of items)\n    array.push(<li>{item.text}</li>)\n\n" +
				"render (\n    <ul>\n        {array}\n    </ul>\n)",
			"jsx",
		},
		{
			"jvns shell loop",
			"for term in $(toe -a | awk '{print $1}')\ndo\n  echo $term\n" +
				"  infocmp -1 -T \"$term\" 2>/dev/null | grep 'clear=' | sed 's/clear=//g;s/,//g'\ndone",
			"bash",
		},
		{
			"shebang",
			"#!/usr/bin/env python3\nmain()",
			"python",
		},
		{
			"shell session",
			"$ ls -la\ntotal 24\ndrwxr-xr-x   5 julia  staff   160 Mar  7 10:00 .",
			"console",
		},
		{
			"go",
			"func main() {\n    s := fetch(ctx)\n    if err != nil {\n" +
				"        log.Fatal(err)\n    }\n    fmt.Println(s)\n}",
			"go",
		},
		{
			"python",
			"def fib(n):\n    a, b = 0, 1\n    for _ in range(n):\n" +
				"        a, b = b, a + b\n    return a\n\nprint(fib(10))",
			"python",
		},
		{
			"javascript",
			"const items = await fetch(url).then(r => r.json());\nitems.forEach((item) => {\n" +
				"  console.log(`${item.name}: ${item.count}`);\n});",
			"javascript",
		},
		{
			"rust",
			"fn main() {\n    let nums: Vec<i32> = (1..10).filter(|n| n % 2 == 0).collect();\n" +
				"    println!(\"{:?}\", nums);\n}",
			"rust",
		},
		{
			"rust lifetimes and raw pointer",
			"fn get_str<'a>(s: *const String) -> &'a str {\n    unsafe { &*s }\n}",
			"rust",
		},
		{
			"c char comparison is not a lifetime",
			"#include <stdio.h>\n\nint classify(char c) {\n    if (c<'a' || c>'z') return 0;\n    return 1;\n}",
			"c",
		},
		{
			"c",
			"#include <stdio.h>\n\nint main(void) {\n    printf(\"%d\\n\", 1);\n    return 0;\n}",
			"c",
		},
		{
			"cpp",
			"#include <iostream>\n\nint main() {\n    std::cout << \"hi\\n\";\n}",
			"cpp",
		},
		{
			"json",
			"{\n  \"name\": \"example\",\n  \"dependencies\": {\n    \"left-pad\": \"^1.3.0\"\n  }\n}",
			"json",
		},
		{
			"sql caps",
			"SELECT users.name, COUNT(orders.id)\nFROM users\nJOIN orders ON orders.user_id = users.id;",
			"sql",
		},
		{
			"sql lowercase",
			"select name from users where active = true order by name;",
			"sql",
		},
		{
			"diff",
			"diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n@@ -1,3 +1,4 @@\n-old\n+new",
			"diff",
		},
		{
			"html",
			`<div class="post">
  <h1>Title</h1>
  <p>Some text</p>
</div>`,
			"html",
		},
		{
			"dockerfile",
			"FROM golang:1.26\nWORKDIR /app\nCOPY . .\nRUN go build -o clx .",
			"docker",
		},
		{
			"terminal output",
			"total 24\ndrwxr-xr-x   5 julia  staff   160 Mar  7 10:00 .\n" +
				"-rw-r--r--   1 julia  staff  1204 Mar  7 10:00 escape.txt",
			"",
		},
		{
			"ascii diagram",
			"+-----------+       +-----------+\n| terminal  | ----> | pty       |\n" +
				"+-----------+       +-----------+",
			"",
		},
		{
			"prose",
			"This is just a paragraph of explanatory text that happens to be\n" +
				"in a pre block, maybe because the author wanted to preserve the\nline breaks.",
			"",
		},
		{
			"prose starting with sql verb",
			"With a terminal, everything you type turns into bytes that flow\n" +
				"from the keyboard into the line discipline where they wait.",
			"",
		},
		{
			"log output",
			"2024-03-01T10:22:01Z INFO  server started on :8080\n" +
				"2024-03-01T10:22:19Z ERROR upstream timeout after 5s",
			"",
		},
		{
			"escape code table",
			"\\x1b[31m red \\x1b[0m\n\\x1b[1;32m bold green \\x1b[0m\nESC [ 38 ; 5 ; 214 m",
			"",
		},
		{
			"one-liner command",
			"npm install left-pad",
			"",
		},
		{
			"haskell type signature is not rust",
			"add :: Int -> Int -> Int\nadd x y = x + y",
			"",
		},
		{
			"rust match is not javascript",
			"let result = match code {\n    200 => \"ok\",\n    404 => \"missing\",\n};",
			"rust",
		},
		{
			"js for-in is not bash",
			"for (const key in obj) {\n  console.log(`${key}`);\n}",
			"javascript",
		},
		{
			"php interpolation is php, not bash",
			"$greeting = \"Hello ${name}\";\necho \"$greeting\";",
			"php",
		},
		{
			"tagless php is not javascript",
			"function greet($name) {\n    $map = ['a' => 1, 'b' => 2];\n    echo $name;\n}",
			"php",
		},
		{
			"terraform is not bash",
			"resource \"aws_instance\" \"web\" {\n  ami = \"${var.ami_id}\"\n}",
			"",
		},
		{
			"github actions is not bash",
			"steps:\n  - uses: actions/checkout@v4\n  - run: make release\n    tag: \"${{ github.ref_name }}\"",
			"",
		},
		{
			"makefile is not bash",
			"CC = gcc\nall: main.o\n        $(CC) -o app main.o 2>&1",
			"",
		},
		{
			"js template with dollar string",
			"const label = `${item.name}`;\nconst price = \"$\" + amount.toFixed(2);",
			"javascript",
		},
		{
			"pom is xml, not html",
			"<dependency>\n  <groupId>org.apache.commons</groupId>\n  <artifactId>commons-lang3</artifactId>\n</dependency>",
			"xml",
		},
		{
			"android layout is xml",
			"<LinearLayout xmlns:android=\"http://schemas.android.com/apk/res/android\"\n    android:layout_width=\"match_parent\" />",
			"xml",
		},
		{
			"svg is xml",
			"<svg viewBox=\"0 0 100 100\">\n  <circle cx=\"50\" cy=\"50\" r=\"40\" />\n</svg>",
			"xml",
		},
		{
			"component jsx is not html",
			"<Button onClick={() => save()}>\n  Save changes\n</Button>",
			"jsx",
		},
		{
			"swift is not python",
			"import Foundation\n\nlet greeting = \"Hello\"\nprint(greeting)",
			"",
		},
		{
			"kotlin is not python",
			"import java.util.Random\n\nfun main() {\n    print(Random().nextInt())\n}",
			"",
		},
		{
			"cpp class detects as cpp",
			"#include \"widget.h\"\n\nclass Widget {\npublic:\n    void draw();\n};",
			"cpp",
		},
		{
			"objective-c import",
			"#import <Foundation/Foundation.h>\n\n@interface Greeter : NSObject\n@end",
			"objective-c",
		},
		{
			"lowercase latex is not a console session",
			"$ f(x) = x^2 + 1 $",
			"",
		},
		{
			"doctype is html",
			"<!DOCTYPE html>\n<html>\n<body><p>hi</p></body>\n</html>",
			"html",
		},
		{
			"php tag",
			"<?php\necho $greeting;",
			"php",
		},
		{
			"autolink is not html",
			"<https://example.com/some/path>",
			"",
		},
		{
			"shouted prose is not sql",
			"SELECT YOUR FAVORITE ITEMS FROM THE MENU BELOW",
			"",
		},
		{
			"log line is not sql",
			"UPDATE FAILED: could not write TO TABLE users",
			"",
		},
		{
			"latex is not a console session",
			"$ E = mc^2 $",
			"",
		},
		{
			"cpp lambda without include stays plain",
			"std::sort(items.begin(), items.end(),\n    [](const Item& a, const Item& b) -> bool { return a.id < b.id; });",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, GuessLang(tt.text))
		})
	}
}
