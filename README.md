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

`circumflex` is a text user interface (TUI) application for browsing Hacker&nbsp;News. Stay up to date on the latest stories and discussions without leaving your terminal.

<p align="center">
  <img src="screenshots/mainview.png" width="700" alt="^"/>
</p>


## Getting started
Install `circumflex` with Homebrew: 
```console
# Add 'bensadeh/circumflex' to list of taps and install
brew install bensadeh/circumflex/circumflex

# Run circumflex
clx
```

Press <kbd>i</kbd> for a list over keybindings.

## Overview
### Features and limitations
`circumflex` is a TUI application that lets you browse Hacker News in your terminal. It can list submissions by category 
and show the comment section for each article. It also respects and adjusts to whatever color scheme you use.

`circumflex` does not support any login related functionality. This includes up-/down-voting, flagging,
submitting articles and posting comments.

<p align="center">
  <img src="screenshots/comments.png" width="700" alt="^"/>
</p>

### Comment section
Comments are pretty-printed and piped to the pager `less`. To present a nice and readable comment section, 
`circumflex` features:
* Text in **bold**, _italics_, [hyperlinks](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda) and 
  `code` where available
* Indented and color-coded replies
* Colored references (`[1]`, `[2]`, `[…]`)
* Labels for Original Posters (`OP`), Parent Posters (`PP`) and moderators (`mod`)
* Ability to jump between top-level comments by searching for `::`

<p align="center">
  <img src="screenshots/linkHighlights.png" width="700" alt="^"/>
</p>

## Known issues
The first keystroke is lost when moving from cview (submission menu) to viewing comments in `less`, see [gdamore/tcell#194](https://github.com/gdamore/tcell/issues/194).

## Under the hood
`circumflex` uses:
* [tcell](https://github.com/gdamore/tcell) and [cview](https://gitlab.com/tslocum/cview) for the TUI
* [cheeaun's unofficial Hacker News API](https://github.com/cheeaun/node-hnapi) for fetching submissions and comments
* [`less`](http://greenwoodsoftware.com/less/) for viewing comments
* [go-term-text](https://github.com/MichaelMure/go-term-text) for wrapping and indenting comments

Screenshots use:
* [iTerm2](https://iterm2.com/) for the terminal
* [Palenight Theme](https://github.com/JonathanSpeek/palenight-iterm2) for the color scheme
* [JetBrains Mono](https://github.com/JetBrains/JetBrainsMono) for the font
