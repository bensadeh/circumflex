<p align="center">
  <img src="images/circumflex.png" width="300" alt="^"/>
</p>

#
> Work in progress

`circumflex` is a command line tool for browsing Hacker&nbsp;News submissions and reading comments.

## Features
- Simple overview of top submissions
- Comments are properly wrapped, indented and piped to `less`

<p align="center">
  <img src="images/mainview.png" width="700" alt="^"/>
</p>

<p align="center">
  <img src="images/comments.png" width="700" alt="^"/>
</p>

## Installation
`circumflex` is written in Go. Simply:

    $ go install

And you're good to go.

## Frameworks and Credits
`circumflex` uses:
* [cobra](https://github.com/spf13/cobra) for the CLI
* [tcell](https://github.com/gdamore/tcell) and [cview](https://gitlab.com/tslocum/cview) for the UI
* [colly](https://github.com/gocolly/colly) and [go-hackernews](https://github.com/jacktantram/go-hackernews) for scraping
* `less` for viewing comments
* [Palenight Theme for iTerm2](https://github.com/JonathanSpeek/palenight-iterm2) for the color scheme
* [JetBrains Mono](https://www.jetbrains.com/lp/mono/) for the font
