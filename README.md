<p align="center">
  <img src="images/circumflex.png" width="300" alt="^"/>
</p>

#

`circumflex` is a command line tool for browsing Hacker News submissions and reading comments. Work in progress.

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

    go install

## Frameworks and Credits
`circumflex` uses:
* [cobra](https://github.com/spf13/cobra) for the CLI interface
* [tcell](https://github.com/gdamore/tcell) and [cview](https://gitlab.com/tslocum/cview) for the UI
* [colly](https://github.com/gocolly/colly) and [go-hackernews](https://github.com/jacktantram/go-hackernews) for scraping
* `less` as the pager to view comments
