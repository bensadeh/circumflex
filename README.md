<p align="center">
  <img src="images/circumflex.png" width="300" alt="^"/>
</p>

#
> Work in progress

`circumflex` is a command line tool for browsing Hacker&nbsp;News submissions and reading comments.

## Features
- Overview of top submissions
- View comments in `less`
  * Comments are wrapped and color-indented
  * Comments are properly formatted to show *italics*, <ins>hyperlinks</ins> and `code blocks`
  * Original poster (OP) and moderators are labelled

<p align="center">
  <img src="images/mainview.png" width="700" alt="^"/>
</p>

<p align="center">
  <img src="images/comments.png" width="700" alt="^"/>
</p>

## Installation
`circumflex` is written in Go. Simply:

    $ go install

Then run with:

    $ clx

## Frameworks and Credits
`circumflex` uses:
* [cobra](https://github.com/spf13/cobra) for the CLI
* [tcell](https://github.com/gdamore/tcell) and [cview](https://gitlab.com/tslocum/cview) for the UI
* [cheeaun's unofficial Hacker News API](https://github.com/cheeaun/node-hnapi) for fetching submissions and comments
* [less](http://greenwoodsoftware.com/less/) for viewing comments
* [Palenight Theme for iTerm2](https://github.com/JonathanSpeek/palenight-iterm2) for the color scheme
* [JetBrains Mono](https://github.com/JetBrains/JetBrainsMono) for the font
