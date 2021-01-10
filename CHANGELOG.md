# Changelog

## 0.x
_WIP_

**New features:**
- Added option to preserve right margin in comment section
- Added option to customize and colorize submission headlines

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
