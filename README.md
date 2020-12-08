<p align="center">
  <img src="screenshots/circumflex.png" width="350" alt="^"/>
</p>

#
<div align="center">
  
[![Latest release](https://img.shields.io/github/v/release/bensadeh/circumflex?label=stable&color=e1acff&labelColor=292D3E)](https://github.com/bensadeh/circumflex/releases)
[![Changelog](https://img.shields.io/badge/docs-changelog-9cc4ff?labelColor=292D3E)](https://github.com/bensadeh/circumflex/blob/master/CHANGELOG.md)
[![License](https://img.shields.io/github/license/bensadeh/circumflex?color=c3e88d&labelColor=292D3E)](https://github.com/bensadeh/circumflex/blob/master/LICENSE)
[![Go Report Card](https://img.shields.io/github/go-mod/go-version/bensadeh/circumflex?color=ffe585&labelColor=292D3E)](https://github.com/bensadeh/circumflex/blob/master/go.mod)
</div>

`circumflex` is Hacker&nbsp;News on the command line. Browse submissions and read comments without leaving your terminal.

<p align="center">
  <img src="screenshots/mainview.png" width="700" alt="^"/>
</p>


## Installation
### Homebrew
The following command adds bensadeh/circumflex to your list of [taps](https://docs.brew.sh/Taps) and installs `circumflex` from [this formula](https://github.com/bensadeh/homebrew-circumflex): 
```console
brew install bensadeh/circumflex/circumflex
```

To run `circumflex`:

```console
clx
```

Press <kbd>i</kbd> for help.

## Features

### Main features
`circumflex` lets you:
* 🗞 Browse Hacker News by category (New, Newest, Ask HN or Show HN)
* 💬 Read comments in the pager `less`

### Secondary features
Additionally, `circumflex` supports the following nice-to-have features:
* ⚡️ Vim keybindings
* 🌐 UTF-8 encoding
* 🎨 Text formatting in **bold**, _italics_, [hyperlinks](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda) and `code`
* 🖍 Uses your terminal's own color scheme
* 📐 Comments are indented and color-coded
* 🔄 References in comments ([1],[2] etc.) are colored for easier cross-referencing


<p align="center">
  <img src="screenshots/comments.png" width="700" alt="^"/>
</p>

<p align="center">
  <img src="screenshots/linkHighlights.png" width="700" alt="^"/>
</p>


## Known issues
The first keystroke is lost when moving from cview (submission menu) to viewing comments in `less`, see [gdamore/tcell#194](https://github.com/gdamore/tcell/issues/194).

## Under the hood
`circumflex` uses:
* [tcell](https://github.com/gdamore/tcell) and [cview](https://gitlab.com/tslocum/cview) for the UI
* [cheeaun's unofficial Hacker News API](https://github.com/cheeaun/node-hnapi) for fetching submissions and comments
* [`less`](http://greenwoodsoftware.com/less/) for viewing comments
* [go-term-text](https://github.com/MichaelMure/go-term-text) for wrapping and indenting comments

Screenshots use:
* [iTerm2](https://iterm2.com/) for the terminal
* [Palenight Theme](https://github.com/JonathanSpeek/palenight-iterm2) for the color scheme
* [JetBrains Mono](https://github.com/JetBrains/JetBrainsMono) for the font
