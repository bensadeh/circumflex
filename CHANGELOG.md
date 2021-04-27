# Changelog

## 1.1
_WIP_

**Cosmetic**
- Show item ID in comment section

**Bugfixes**
- Fixed a bug where hrefs were not stripped inside quotes
- Fixed a bug where pressing <kbd>G</kbd> while in Relative Numbering mode would not
  properly update the left margin on the favorites page

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

**New features:**
- Press <kbd>F</kbd> to add submission to Favorites by ID

## 0.13
_2021-04-02_

**New features:**
- Submissions can now be added to Favorites

**Cosmetic**
- Headlines are now syntax highlighted by default
- Definition on info screen now realigns after resizing the terminal

**Backend:**
- Extracted logic for handling submissions out of the model

## 0.12
_2021-03-03_

**New features:**
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

**Cosmetic:**
- Cleaned up the keymap screen
- `Highlight Headlines` now either reverse highlights all headlines or color highlights all headlines (YC-funded
  startups were previously colorized in option 1)

**Bugfixes:**
- Fixed a bug where the `config.env` template wasn't created with default values

**Backend:**
- Bump to Go 1.16

## 0.9
_2021-02-08_

**Cosmetic:**
- Settings screen now highlights booleans and integers

**Bugfixes:**
- Fixed a bug where jumping multiple lines would lead to an infinite loop

## 0.8
_2021-02-07_

**New features:**
- Posts of the type '`YC startup` is hiring' are now hidden by default and can be enabled in the settings

**Cosmetic:**
- Info screen now shows version number
- Increased spacing between the descriptions on the settings page for easier readability

**Backend:**
- Added version number to User-Agent ID string
- Replaced stock http with [resty](https://github.com/go-resty/resty)

## 0.7.1
_2021-01-21_

**Backend:**
- Added User-Agent ID

## 0.7
_2021-01-21_

**Cosmetic:**
- Changed `[Y]` to `🆈`
- Submission text highlighting now turned off by default
- Added `ERROR` and `SUCCESS` labels to some messages
- Changed `YC S/WXX` labels

**Bugfixes:**
- Fixed a bug where JSON errors were not handled properly

## 0.6
_2021-01-15_

**New features:**
- Numerical input on the home screen now repeats the next action N number of times (same as in Vim's Normal mode)
- Added option 'Use Relative Numbering': Relative numbering marks each line with a number relative to the distance from
  the currently selected element (similar to Vim's hybrid line number mode)

**Bugfixes:**
- Fixed a bug where the descriptions in `config.env` contained raw ANSI escape codes

## 0.5
_2021-01-10_

**New features:**
- Added option to preserve right margin in comment section
- Added option to customize and colorize submission headlines

**Cosmetic:**
- Settings will be shown in two columns if there is enough screen space

**Bugfixes:**
- Fixed a bug where setting comment width to 0 did not use the whole screen

## 0.4
_2021-01-08_

**New features:**
- Added functionality to customize `circumflex` by editing `config.env` or exporting environmental variables

**Cosmetic:**
- Information page now has three screens: Information Home Screen, Keymaps and Settings

## 0.3
_2020-12-17_

**New features:**
- Added a status bar to show contextual information
- `circumflex` now handles connection errors gracefully

**Bugfixes:**
- `circumflex` will no longer suspend the application when trying to open submission of the type 'Company X (YC W20) is
  hiring'

**Backend:**
- Use the 'level' field directly from the API instead of calculating it
- Large refactor of the program architecture

## 0.2.4
_2020-12-11_

**Bugfixes:**
- Fixed a bug where the number of submissions to view was not calculated correctly
- Fixed a bug where pressing <kbd>0</kbd> would go to the last element on the list instead of the 10th
- Info line about entering less now appears right after submission info line

## 0.2.3
_2020-12-05_

**New features:**
- Added keybinding: Press <kbd>0</kbd>-<kbd>9</kbd> to go directly to submission

**Cosmetic:**
- Indented comment bar now also uses brighter colors
- References now also uses brighter colors
- Comment section: Added a notice about entering `less` and how to exit from it

## 0.2.2
_2020-11-28_

**Cosmetic:**
- Selected items now uses the terminal's default colors in order to correctly highlight the selection regardless of
  color scheme

## 0.2.1
_2020-11-27_

**New features:**
- Added keybinding: Press <kbd>r</kbd> to refresh

**Bugfixes:**
- Fixed a bug where `circumflex` would crash while resizing the terminal while on the help screen and on pages larger
  than 1

## 0.2
_2020-11-22_

**New features:**
- Added support for resizing the terminal while `circumflex` is running
- Added keybindings: <kbd>g</kbd> / <kbd>G</kbd> to go to first and last element
- Added keybinding: <kbd>c</kbd> to open submission comments in browser

**Cosmetic:**
- On the submissions page, YC startup labels are now orange text on black background

**Backend:**
- Large parts of the code have been refactored and placed into an MVC pattern
- cheeaun's [unofficial Hacker News API](https://github.com/cheeaun/node-hnapi): Changed API endpoint to use Cloudflare
  CDN

## 0.1
_2020-11-15_

- First Release
