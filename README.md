<p align="center">
  <img src="images/circumflex.png" width="300" alt="^"/>
</p>

#
`circumflex` is Hacker&nbsp;News on the command line. It let's you browse submissions and comments in a way that feels native to the terminal. 

## Features
- Concise overview of top submissions
- Pleasing and structured comments view

<p align="center">
  <img src="images/mainview.png" width="700" alt="^"/>
</p>

<p align="center">
  <img src="images/comments.png" width="700" alt="^"/>
</p>

<p align="center">
  <img src="images/linkHighlights.png" width="700" alt="^"/>
</p>


## Viewing comments

### Appearence
Hacker News's text-centric approach lends itself well to be viewed in the terminal. Comments are color-indented to distinguish parent posts from their children and their siblings. Should your terminal support the relevant ANSI escape sequences, comments are properly formatted in *italics*, [hyperlinks](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda) and `code blocks`. To give context in posts with many replies, Original Poster (OP), Parent Poster (PP) and moderators are labelled. Lastly, references ([x]) are color-coded for easier readability.

### Navigation

`circumflex` pipes comments to the pager `less`. Here is a short recap of commonly used navigation commands:

<pre>
  <kbd>↓</kbd>, <kbd>j</kbd>: forward one line
  <kbd>↑</kbd>, <kbd>k</kbd>: backward one line
  <kbd>d</kbd>: forward one half-window
  <kbd>u</kbd>: backward one half-window
  <kbd>q</kbd>: exit
</pre>

### Moving between top-level posts *(or: How I Stopped Worrying and Learned to Love `less`)*
One of the most useful features of browsing comment threads is comment folding. However, since `less` is a pager, it does not support changing (or collapsing) the text it presents.

As an alternative, `circumflex` prints every top level comment with the anchor `::`. Using `less`'s search functionality, top-level comments can brew navigated by searching for the anchor and typing <kbd>n</kbd> or <kbd>N</kbd> to move forwards or backwards between the posts.

<pre>  
  <kbd>/</kbd>: search
  <kbd>n</kbd>: repeat search forwards
  <kbd>N</kbd>: repeat search backward
</pre>

## Installation
`circumflex` is written in Go. Clone the repo and run:

    $ go install

Then run with:

    $ clx

## Known issues
The first keystroke is lost when viewing comments in `less`, see [gdamore/tcell#194](https://github.com/gdamore/tcell/issues/194).

## Under the hood
`circumflex` uses:
* [cobra](https://github.com/spf13/cobra) for the CLI
* [tcell](https://github.com/gdamore/tcell) and [cview](https://gitlab.com/tslocum/cview) for the UI
* [cheeaun's unofficial Hacker News API](https://github.com/cheeaun/node-hnapi) for fetching submissions and comments
* [`less`](http://greenwoodsoftware.com/less/) for viewing comments
* [go-term-text](https://github.com/MichaelMure/go-term-text) for wrapping and indenting comments

Screenshots use:
* [iTerm2](https://iterm2.com/) for the terminal
* [Palenight Theme](https://github.com/JonathanSpeek/palenight-iterm2) for the color scheme
* [JetBrains Mono](https://github.com/JetBrains/JetBrainsMono) for the font
