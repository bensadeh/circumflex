# Changelog
## 1.21
_2021-09-12_

**Bugfixes**
- Invisible anchors are now longer properly hidden on all terminals
- The filtering logic now works with zero width spaces
- Fixed a bug where the root comment headline was one character
longer than the comment width

**Cosmetic**
- Viewing articles in Reader Mode now updates 'mark as read' indicator

## 1.20
_2021-09-04_

**New features**
- New keybinding: press <kbd>n</kbd>/<kbd>N</kbd> to jump to the next top-level comment
or headline
  - (No longer required to search for the string `::`)
- Added an option to set the header to the orange and classic Hacker News header

**Cosmetic**
- Added custom filtering rules for the following sites:
  - `nytimes.com`
  - `economist.com`
  - `tomshardware.com`
  - `cnn.com`
  - `arstechnica.com`
  - `macrumors.com`
  - `wired.com`
  - `wired.co.uk`
  - `theguardian.com`
  - `axios.com`
  - `9to5mac.com`

## 1.19
_2021-08-28_

This release replaces `lynx` for rendering `HTML` in favour of handling the rendering logic directly in `circumflex`.

**Backend**
- Bump Go to 1.17
- Reader Mode: Removed `lynx` as a dependency
   - Added support for code blocks and in-line code highlighting
   - Added support for prettier tables
   - Added support for rendering different headers (`h1` - `h6`)
   - Added support for well-formatted lists and sub-lists

## 1.18
_2021-08-03_

**New features**
- Added an option to remove indentation symbols from the comment section

**Cosmetic**
- Better handling of stray newlines
- Monthly Who is hiring posts now have normal syntax highlighting

**Backend**
- Bump goreadability
- Bump cobra

## 1.17
_2021-07-28_

This release marks the one year anniversary since the first commit of `circumflex` 🎉

**Cosmetic**
- Included syntax highlighting for YC-funded startups in the comment section
- Code snippets are now in italics and magenta (was: just magenta)
- Monthly 'Who is Hiring' posts now honor the 'highlight headlines' setting
- Remove `FAANG` highlighting (it was a bit too colorful)
- Better handling of fractions
- Better handling of smiley to emoji substitution
- Better handling of syntax highlighting for URLs
- Better handling of double dashes to em-dashes conversion
- Better handling of username highlights
- Comments that have been deleted and have no replies are no longer printed
- Redesigned comment section header


## 1.16
_2021-07-24_

**New features**
- `circumflex` can now be customized with flags

**Changes**
- Highlight headlines option has been simplified and can now be either 
enabled or disabled (Removed an option to highlight headlines with the reverse
highlighting flag)

**Cosmetic**
- Code snippets are now highlighted in magenta instead of blue
- Mark as read setting now turned on by default
- Rename `create_config` command to config
- Hyperlinks are now in blue instead of dimmed blue

**Bugfixes**
- Fixed a bug where highlighting of `$` would cause a panic
- Comments are now properly shortened when they reach the edge of the screen
- Fractions now have proper spacing between them and the next word

## 1.15
_2021-07-22_

**New features**
- Submissions can now be marked as read (turned off by default)

**Cosmetic**
- Highlight mentions of `FAANG` in the comment section
- Highlight moderator author names in main screen
- Removed white from indent blocks
- Comments now show prettier Unicode fractions 
- Double single dashes (--) now appear as a single em-dash (—)
- Triple dots (...) now appear as a single ellipsis (…)
- Added an option to convert smileys to emojis
- Removed support for hyperlinks in the terminal since they were somewhat too complex
compared to the benefit / convenience they provided
- URLs are now highlighted in dimmed blue

## 1.14
_2021-07-15_

**Bugfixes**
- Fixed a bug where a single @ would be highlighted
- Fixed a bug where alternate indentation block would not apply for quotes

**Cosmetic**
- Monthly Who is hiring, Freelancer, Who wants to be hired posts are now highlighted
in their own color

**Backend**
- Bump cobra and viper
- Bump cview

## 1.13
_2021-06-30_

**New features**
- Comment syntax highlighting can now be disabled

**Cosmetic**
- Adjusted article width in Reader Mode
- Added a message when successfully running `clx create_example_config`
- Reworked pagination indicator
- `Ask HN` is now highlighted in blue
- `Tell HN` is now highlighted in magenta
- `$variables` are now highlighted in cyan
- `IANAL` is now highlighted in red
- `IAAL` is now highlighted in green

## 1.12
_2021-06-14_

**New features**
- `clx id [item-id]` respects config.env and set environment variables

**Cosmetic**
- Text inside backticks is highlighted
- Mentions in comments `@username` are highlighted
- Changed highlighting of PDF, video and audio in headlines

## 1.11
_2021-06-13_

**New features**
- Added a command to go directly to the comment section for a given ID

**Cosmetic**
- Reader Mode: Do not print Footnotes in Wikipedia articles
- Reader Mode: Improved formatting for bullet points

## 1.10
_2021-06-12_

**Bugfixes**
- Fixed a bug where deeply nested comment did not use the whole screen

## 1.9
_2021-06-12_

**Bugfixes**
- Fixed a bug where the formatting in the comment section would occasionally break

**Cosmetic**
- Default comment width is now 65 instead of 70
- Comment quotes now have an indentation block
- Changed the order of colors for the indentation blocks
- Code blocks now use the whole screen
- Better handling of references inside quotes/nested blocks in Reader Mode

## 1.8
_2021-06-02_

**Cosmetic**
- A confirmation message is now shown after adding a story to favorites by ID
- Submissions in specific formats ([pdf], [audio], etc.) are now highlighted
- Monthly `Who is hiring` posts are now highlighted
- Wikipedia articles in Reader Mode now have improved formatting
- Improved formatting for confirmation, warning and error messages

## 1.7
_2021-05-30_

**Cosmetic**
- Keymaps screen is now fixed-width

**Bugfixes**
- Fixed a bug where references was printed twice in Reader Mode

## 1.6
_2021-05-30_

**Cosmetic**
- Quotes in Reader Mode are now dimmed and italicized

**Bugfixes**
- Fixed a bug where scrolling backwards in less would lead to improper formatting


## 1.5
_2021-05-24_

**New features**
- Added option to force open article in Reader mode
- Added option to use alternate indentation blocks for compatibility issues

## 1.4
_2021-05-16_

**New features**
- Create example config from the terminal with `clx create_example_config`

**Cosmetic**
- Simplified keymaps screen

## 1.3
_2021-05-07_

**New features**
- Added a validator to prevent entering Reader Mode on sites that are known to be unsupported

**Bugfixes**
- Fixed a bug where `Reader View` mode would occasionally format references incorrectly

## 1.2
_2021-05-06_

**New features**
- Read a submission's article in `Reader View` mode

## 1.1
_2021-04-27_

**Cosmetic**
- Show item ID in comment section

**Bugfixes**
- Fixed a bug where hrefs were not stripped inside quotes
- Fixed a bug where pressing <kbd>G</kbd> while in Relative Numbering mode would not
  properly update the left margin on the favorites page
- Fixed a bug where a refresh wouldn't trigger after returning from the comment section 


## 1.0
_2021-04-24_

**Bugfixes**
- Fixed a bug where the first keystroke was lost when entering the comment section
- Fixed a bug where pressing <kbd>G</kbd> while in Relative Numbering mode would not
properly update the left margin

## 0.17
_2021-04-07_

**Bugfixes**
- Fixed a bug where hidden stories of the type `X is hiring` would cause a panic

**New features:**
- Exit info screen with <kbd>Esc</kbd> and <kbd>?</kbd> (in addition to <kbd>i</kbd>)

## 0.16
_2021-04-07_

**Backend**
- Rename submission to story

**Bugfixes**
- Fixed a bug where triple spaces would not be correctly converted to single space

## 0.15
_2021-04-04_

**Cosmetic**
- All views are now responsive

**Backend:**
- Simplified Info View logic

## 0.14
_2021-04-03_

**New features**
- Press <kbd>F</kbd> to add submission to Favorites by ID

## 0.13
_2021-04-02_

**New features**
- Submissions can now be added to Favorites

**Cosmetic**
- Headlines are now syntax highlighted by default
- Definition on info screen now realigns after resizing the terminal

**Backend:**
- Extracted logic for handling submissions out of the model

## 0.12
_2021-03-03_

**New features**
- Quotes are now dimmed and italicized

**Cosmetic**
- Changed appearance of error and success notifications

**Bugfixes**
- Fixed a bug where brackets in titles would not appear

## 0.11.1
_2021-03-01_

**Bugfixes**
- Fixed a bug where the separator between the submissions's root comment and the comment section was not properly
  formatted

## 0.11
_2021-02-28_

**Cosmetic**
- The top bar is now transparent instead of orange
- Settings screen has been redesigned

**Bugfixes**
- Fixed a bug where references (`[1]`, `[2]`, `[…]`) would highlight inside code blocks

## 0.10
_2021-02-25_

**New features:**
- `g` and `G` works the same way as it does in Vim

**Cosmetic**
- Cleaned up the keymap screen
- `Highlight Headlines` now either reverse highlights all headlines or color highlights all headlines (YC-funded
  startups were previously colorized in option 1)

**Bugfixes**
- Fixed a bug where the `config.env` template wasn't created with default values

**Backend**
- Bump to Go 1.16

## 0.9
_2021-02-08_

**Cosmetic**
- Settings screen now highlights booleans and integers

**Bugfixes**
- Fixed a bug where jumping multiple lines would lead to an infinite loop

## 0.8
_2021-02-07_

**New features**
- Posts of the type '`YC startup` is hiring' are now hidden by default and can be enabled in the settings

**Cosmetic**
- Info screen now shows version number
- Increased spacing between the descriptions on the settings page for easier readability

**Backend**
- Added version number to User-Agent ID string
- Replaced stock http with [resty](https://github.com/go-resty/resty)

## 0.7.1
_2021-01-21_

**Backend**
- Added User-Agent ID

## 0.7
_2021-01-21_

**Cosmetic**
- Changed `[Y]` to `🆈`
- Submission text highlighting now turned off by default
- Added `ERROR` and `SUCCESS` labels to some messages
- Changed `YC S/WXX` labels

**Bugfixes:**
- Fixed a bug where JSON errors were not handled properly

## 0.6
_2021-01-15_

**New features**
- Numerical input on the home screen now repeats the next action N number of times (same as in Vim's Normal mode)
- Added option 'Use Relative Numbering': Relative numbering marks each line with a number relative to the distance from
  the currently selected element (similar to Vim's hybrid line number mode)

**Bugfixes**
- Fixed a bug where the descriptions in `config.env` contained raw ANSI escape codes

## 0.5
_2021-01-10_

**New features**
- Added option to preserve right margin in comment section
- Added option to customize and colorize submission headlines

**Cosmetic**
- Settings will be shown in two columns if there is enough screen space

**Bugfixes**
- Fixed a bug where setting comment width to 0 did not use the whole screen

## 0.4
_2021-01-08_

**New features**
- Added functionality to customize `circumflex` by editing `config.env` or exporting environmental variables

**Cosmetic**
- Information page now has three screens: Information Home Screen, Keymaps and Settings

## 0.3
_2020-12-17_

**New features**
- Added a status bar to show contextual information
- `circumflex` now handles connection errors gracefully

**Bugfixes**
- `circumflex` will no longer suspend the application when trying to open submission of the type 'Company X (YC W20) is
  hiring'

**Backend**
- Use the 'level' field directly from the API instead of calculating it
- Large refactor of the program architecture

## 0.2.4
_2020-12-11_

**Bugfixes**
- Fixed a bug where the number of submissions to view was not calculated correctly
- Fixed a bug where pressing <kbd>0</kbd> would go to the last element on the list instead of the 10th
- Info line about entering less now appears right after submission info line

## 0.2.3
_2020-12-05_

**New features**
- Added keybinding: Press <kbd>0</kbd>-<kbd>9</kbd> to go directly to submission

**Cosmetic**
- Indented comment bar now also uses brighter colors
- References now also uses brighter colors
- Comment section: Added a notice about entering `less` and how to exit from it

## 0.2.2
_2020-11-28_

**Cosmetic**
- Selected items now uses the terminal's default colors in order to correctly highlight the selection regardless of
  color scheme

## 0.2.1
_2020-11-27_

**New features**
- Added keybinding: Press <kbd>r</kbd> to refresh

**Bugfixes**
- Fixed a bug where `circumflex` would crash while resizing the terminal while on the help screen and on pages larger
  than 1

## 0.2
_2020-11-22_

**New features**
- Added support for resizing the terminal while `circumflex` is running
- Added keybindings: <kbd>g</kbd> / <kbd>G</kbd> to go to first and last element
- Added keybinding: <kbd>c</kbd> to open submission comments in browser

**Cosmetic**
- On the submissions page, YC startup labels are now orange text on black background

**Backend**
- Large parts of the code have been refactored and placed into an MVC pattern
- cheeaun's [unofficial Hacker News API](https://github.com/cheeaun/node-hnapi): Changed API endpoint to use Cloudflare
  CDN

## 0.1
_2020-11-15_

- First Release
