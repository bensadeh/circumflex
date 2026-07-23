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
			"annotated def with for-in loop",
			"def pack_number(digits: list[int], base: int) -> int:\n    number = 0\n\n" +
				"    for digit in digits:\n      assert digit < base\n\n" +
				"      number = number * base\n      number = number + digit\n\n    return number",
			"python",
		},
		{
			"annotated def with nested range loops",
			"def unpack_trits(packed: bytes) -> list[int]:\n    trits: list[int] = []\n\n" +
				"    for byte in packed:\n        b = byte\n        for i in range(5):\n" +
				"            b = b * 3\n            trit = b >> 8\n            b = b & 0xFF\n\n    return trits",
			"python",
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
		{
			"elisp use-package",
			"(use-package eglot\n  :ensure nil\n  :hook ((scala-ts-mode . eglot-ensure)\n" +
				"         (before-save . eglot-format-buffer)))",
			"elisp",
		},
		{
			"elisp fragment opening on a keyword",
			":bind ((\"C-c i e\" . eglot)\n       (\"C-c i r\" . eglot-rename))\n:init\n" +
				"(setq eglot-autoshutdown t)",
			"elisp",
		},
		{
			"elisp advice with unbalanced excerpt",
			"(defun my/eglot-uri-fix (orig-fn uri &rest args)\n" +
				"  (if (string-prefix-p \"jar:///\" uri)\n" +
				"      (apply orig-fn (replace-regexp-in-string \"^jar:///\" \"jar:file:///\" uri) args)\n" +
				"    (apply orig-fn uri args)))\n\n(advice-add 'eglot--uri-to-path :around #'my/eglot-uri-fix)",
			"elisp",
		},
		{
			"clojure",
			"(ns app.core\n  (:require [clojure.string :as str]))\n\n(defn greet [name]\n" +
				"  (println (str \"Hello, \" name)))",
			"clojure",
		},
		{
			"scheme",
			"(define (factorial n)\n  (if (= n 0)\n      1\n      (* n (factorial (- n 1)))))",
			"scheme",
		},
		{
			"common lisp",
			"(defpackage :myapp\n  (:use :cl))\n\n(in-package :myapp)\n\n(defun main ()\n" +
				"  (format t \"hello~%\"))",
			"common-lisp",
		},
		{
			"lisp without dialect evidence stays generic",
			"(mapcar (lambda (x) (* x x))\n        (list 1 2 3 4))",
			"lisp",
		},
		{
			"lettered list is not lisp",
			"(a) open the config file\n(b) add the hook\n(c) restart the editor",
			"",
		},
		{
			"citations are not lisp",
			"(Smith 2019)\n(Jones and Ng 2020)\n(Chen 2021)",
			"",
		},
		{
			"parenthesized arithmetic is not lisp",
			"(x + y) * (a - b)\n(p - q) / (r + s)",
			"",
		},
		{
			"csharp attributes and interpolation",
			"using Unity.Pipeline.Commands;\nusing UnityEngine;\n\n" +
				"public static class MyPipelineCommands\n{\n" +
				"    [CliCommand(\"greet\", \"Log a greeting and return its length\")]\n" +
				"    public static int Greet(\n" +
				"        [CliArg(\"name\", \"Who to greet\", Required = true)] string name)\n    {\n" +
				"        Debug.Log($\"Hello, {name}!\");\n        return name.Length;\n    }\n}",
			"csharp",
		},
		{
			"csharp linq is not javascript",
			"var users = people.Where(p => p.Age > 18).Select(p => p.Name);\nvar count = users.Count();",
			"csharp",
		},
		{
			"csharp console app",
			"using System;\n\nnamespace Hello\n{\n    class Program\n    {\n" +
				"        static void Main(string[] args) => Console.WriteLine(\"hi\");\n    }\n}",
			"csharp",
		},
		{
			"java class is not csharp",
			"public class Greeter {\n    public static void main(String[] args) {\n" +
				"        System.out.println(\"hi\");\n    }\n}",
			"",
		},
		{
			"java var declaration is not csharp",
			"public void load() {\n    var records = new ArrayList<String>();\n" +
				"    records.add(reader.readLine());\n}",
			"",
		},
		{
			"cpp namespace without include is not csharp",
			"namespace detail {\n\nint helper(int x) {\n    return x * 2;\n}\n\n}",
			"",
		},
		{
			"cpp using namespace is not csharp",
			"using namespace std;\n\nint main() {\n    cout << \"hi\";\n}",
			"",
		},
		{
			"toml table is not a csharp attribute",
			"[Server]\nhost = \"localhost\"\nport = 8080",
			"",
		},
		{
			"markdown link line is not a csharp attribute",
			"[Read the docs](https://example.com/docs)\n[File an issue](https://example.com/issues)",
			"",
		},
		{
			"typescript namespace is not csharp",
			"namespace Validation {\n    export interface StringValidator {\n" +
				"        isAcceptable(s: string): boolean;\n    }\n}",
			"",
		},
		{
			"nix flake",
			"{\n  inputs.nixpkgs.url = \"github:NixOS/nixpkgs/nixos-unstable\";\n\n" +
				"  outputs = { self, nixpkgs }: {\n" +
				"    devShells.default = nixpkgs.legacyPackages.x86_64-linux.mkShell {\n" +
				"      buildInputs = with nixpkgs.legacyPackages.x86_64-linux; [ go gopls ];\n" +
				"    };\n  };\n}",
			"nix",
		},
		{
			"nix let-in derivation",
			"let\n  pkgs = import <nixpkgs> { inherit system; };\nin\n" +
				"pkgs.mkShell {\n  buildInputs = [ pkgs.go ];\n}",
			"nix",
		},
		{
			"gradle assignment is not nix",
			"plugins {\n    id 'java'\n}\n\nsourceCompatibility = 1.8\ntargetCompatibility = 1.8",
			"",
		},
		{
			"java field block is not nix",
			"{\n    private int retries = 3;\n    private String host = \"localhost\";\n}",
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
