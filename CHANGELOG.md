# Changelog

## 0.2.4 
(WIP)

**Bugfixes:**
- Fixed a bug where the number of submissions to view was not calculated correctly
- Info line about entering less now appears right after submission info line

## 0.2.3 
(2020-12-05)

**New features:**
- Added keybinding: Press <kbd>0</kbd>-<kbd>9</kbd> to go directly to submission

**Cosmetic:**
- Indented comment bar now also uses brighter colors
- References now also uses brighter colors
- Comment section: Added a notice about entering `less` and how to exit from it

## 0.2.2 
(2020-11-28)

**Cosmetic:**
- Selected items now uses the terminal's default colors in order to correctly highlight the selection regardless of 
  color scheme

## 0.2.1 
(2020-11-27)

**New features:**
- Added keybinding: Press <kbd>r</kbd> to refresh

**Bugfixes:**
- Fixed a bug where `circumflex` would crash while resizing the terminal while on the help screen and on pages larger 
  than 1

## 0.2 
(2020-11-22)

**New features:**
- Added support for resizing the terminal while `circumflex` is running
- Added keybindings: <kbd>g</kbd> / <kbd>G</kbd> to go to first and last element
- Added keybinding: <kbd>c</kbd> to open submission comments in browser

**Cosmetic:**
- On the submissions page, YC startup labels are now orange text on black background

**Backend:**
- Large parts of the code have been refactored and placed into an MVC pattern
- cheeaun's [unofficial Hacker News API](https://github.com/cheeaun/node-hnapi): Changed API endpoint to use Cloudflare CDN

## 0.1 
(2020-11-15)

- First Release
