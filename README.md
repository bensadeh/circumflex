<p align="center">
  <img src="screenshots/clx.png" width="150" alt="circumflex"/>
</p>

#

<p align="center">
  <code>circumflex</code> is a command line tool for browsing Hacker&nbsp;News in your terminal
</p>


<p align="center">
  <img src="screenshots/main-view.png" alt="Main view"/>
</p>

### Main features

- 🛋 **Everything in one place** — read both the comment section and articles in Reader Mode
- 🌈 **Syntax highlighting** — syntax-aware formatting for comments and headlines
- ⚡️ **Vim-style navigation** — scroll through, jump between and collapse threads with familiar keybindings

**You might also like:**

- 🔍 **Search** — quickly find old submissions by searching all of HN
- 🪟 **Wide view** — story list and content side by side on wide terminals
- 🤹 **Native terminal colors** — you bring your own color scheme, `circumflex` does the rest
- 💎 **Nerd Fonts** — full support for Nerd Fonts as icons
- ❤️ **Add to favorites** — save interesting submissions for later

## Installing

The binary name for `circumflex` is `clx`.

```console
# Homebrew
brew install circumflex

# Nix
nix-shell -p circumflex

# AUR
yay -S circumflex

# Go
go install github.com/bensadeh/circumflex/cmd/clx@latest

# From source
go run ./cmd/clx
```

## Features

### Comment section

Press <kbd>Enter</kbd> to view the comment section.

The comment section has two modes: `read mode` and `navigate mode`.

In `read mode`, you can scroll using the usual vim bindings. You can also jump between top-level comments (<kbd>
n</kbd>/<kbd>N</kbd>), and you can expand and collapse threads by quote level (<kbd>h</kbd>/<kbd>l</kbd>) or all at once
(<kbd>Enter</kbd>).

In `navigate mode`, you can individually select comments and collapse specific threads.

<p align="center">
  <img src="screenshots/comment-section-1.png" width="49%" alt="comment section"/>
  <img src="screenshots/comment-section-2.png" width="49%" alt="comment section"/>
</p>


`circumflex` is read-only and does not support logging in, voting or commenting.

### Reader Mode

Press <kbd>Space</kbd> to read the linked article in Reader Mode. Just like in the comment section, you can jump between
headers using <kbd>n</kbd>/<kbd>N</kbd>, and you can scroll using the usual vim bindings.

<p align="center">
  <img src="screenshots/reader-mode-1.png" width="49%" alt="reader mode"/>
  <img src="screenshots/reader-mode-2.png" width="49%" alt="reader mode"/>
</p>

### Wide view

On wide terminals, the comment section and Reader Mode open in a pane next to the front page instead of replacing it.
The split kicks in at 180 columns; tune the threshold (or force it with `always`/`never`) using
`-w`/`--wide-view`:

```console
clx --wide-view always
```

### Search

Press <kbd>/</kbd> to search all of Hacker News, powered by Algolia. Sort by popularity or date with <kbd>s</kbd> and
narrow the date range with <kbd>d</kbd>.

The comment section and Reader Mode also support search in the content with <kbd>/</kbd>: search for the term and jump
between matches with <kbd>n</kbd>/<kbd>N</kbd>.

### Favorites

Press <kbd>f</kbd> to add the highlighted submission to your favorites. Remove it with <kbd>x</kbd>.

You can also add a submission by `ID` from the command line:

```console
clx add [id]
```

Favorites are stored in `favorites.toml` in the config directory, human-readable and VCS-friendly. A
`favorites.json` from earlier versions is migrated automatically.

### History

Visited submissions are marked as read, and comments added since your last visit are highlighted.

History is stored in `history.json` in the cache directory. Disable tracking with `-d`/`--no-history`, or clear it with:

```console
clx clear
```

### Categories

Switch between categories with <kbd>Tab</kbd>. The header shows `top`, `best`, `ask`, `show` and `favorites` by default.
Pick which ones appear (and in what order) with the `--categories` flag:

```console
clx --categories top,new,best
```

Available categories are `top`, `best`, `new`, `ask`, `show`, `jobs` and `favorites`.

### Configuration

Every flag can be set persistently in `config.toml`. To customize, write out the default config:

```console
clx default-config
```

Flags take precedence over the config file.

The config directory is `~/.config/circumflex` on Linux and `~/Library/Application Support/circumflex` on macOS; the
cache directory is `~/.cache/circumflex` and `~/Library/Caches/circumflex`.

### Theme

`circumflex` uses your terminal's color scheme by default. To customize, write out the default config and edit it:

```console
clx default-theme
```

The theme lives in `theme.toml` in the config directory. 

## Keymaps

Main view keybindings — press <kbd>i</kbd> in any view for the full list, including comment and reader mode.

| Key              | Action                   |
|:-----------------|:-------------------------|
| <kbd>Enter</kbd> | View comments            |
| <kbd>Space</kbd> | Reader Mode              |
| <kbd>Tab</kbd>   | Next category            |
| <kbd>/</kbd>     | Search Hacker News       |
| <kbd>r</kbd>     | Refresh stories          |
| <kbd>o</kbd>     | Open story in browser    |
| <kbd>c</kbd>     | Open comments in browser |
| <kbd>f</kbd>     | Add to favorites         |
| <kbd>x</kbd>     | Remove from favorites    |
| <kbd>u</kbd>     | Toggle read              |
| <kbd>q</kbd>     | Quit                     |

## Usage

Run `clx help` or `man clx` for a full list of available commands and flags.
