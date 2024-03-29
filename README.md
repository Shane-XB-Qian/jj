# jj

![screenshot](https://github.com/Shane-XB-Qian/jj/blob/master/screenshot.png)

## Feature

* similar & forked from [gof](https://github.com/mattn/gof.git) (thanks `mattn`)
* but simplified it & extended just for independent `open2edit` `open2cd` & fun?
* and bug fix & feat extend or at least `jj` should be a more quick/cool name :smile:

## Installation

- `go get github.com/Shane-XB-Qian/jj` (or just download the bin `jj` from this repo).
- please think to write `alias jj='jj || if [ $? -eq 6 ]; then $(cat "$HOME/.jj_tmp_fs"); fi'` into your `.bashrc`.

## Usage

```sh
$ jj
```

* Run `jj` and emit `ctrl-o`, then start to edit (`open2edit`) with `vim`, whatever it was a file or dir.
* Run `jj` and emit `<enter>`, then start to edit (`open2edit`) with `vim` if it was a file, or would quit but `cd` to that dir (`open2cd`) if it was a dir.

- `fuzzy`   if to 'fuzzy' filter        -can be switched -when in `jj` -by `ctrl-r`.
- `dirOnly` if to display/show dir only -can be switched -when in `jj` -by `ctrl-f`.
- `mruHist` if to display/show MRU hist -can be switched -when in `jj` -by `ctrl-v`.
- `curOnly` if to display/show cur only -can be switched -when in `jj` -by `ctrl-\`.

* cur only means only show fs under cur path -or 'cwd' (if set 'root dir')
* cur only is higher priority than mru hist when if both switched to true.

* Tips: Use `tab` to complete dir.
* Note: `mruHist` stored abs path.
* Use `ctrl-z` if to multi-select, and/but to `open2cd` only the last one.

## Options

|Option        |Description                      |
|--------------|---------------------------------|
|-f            |Fuzzy match                      |
|-w            |Init with dir only on            |
|-c            |Init with cur only on            |
|-m            |Init with mru hist on            |
|-d [path]     |Specify root directory           |

- check `-h` for detail.

## Keyboard shortcuts

|Key                                                      |Description                         |
|---------------------------------------------------------|------------------------------------|
|<kbd>CTRL-K</kbd>,<kbd>CTRL-P</kbd>,<kbd>ARROW-UP</kbd>  |Move-up   line -or wrap to bottom   |
|<kbd>CTRL-J</kbd>,<kbd>CTRL-N</kbd>,<kbd>ARROW-DOWN</kbd>|Move-down line -or wrap to top      |
|<kbd>CTRL-G</kbd>                                        |Move-down line -to bottom           |
|<kbd>CTRL-A</kbd>,<kbd>HOME</kbd>                        |Go to head of prompt                |
|<kbd>CTRL-E</kbd>,<kbd>END</kbd>                         |Go to trail of prompt               |
|<kbd>ARROW-LEFT</kbd>                                    |Move-left cursor                    |
|<kbd>ARROW-RIGHT</kbd>                                   |Move-right cursor                   |
|<kbd>CTRL-O</kbd>                                        |Edit the selected file/dir          |
|<kbd>CTRL-I</kbd>,<kbd>TAB</kbd>                         |Complete dir                        |
|<kbd>CTRL-Y</kbd>                                        |Echo cur item to input              |
|<kbd>CTRL-T</kbd>                                        |Toggle view header/trailing of lines|
|<kbd>CTRL-L</kbd>                                        |Redraw                              |
|<kbd>CTRL-U</kbd>                                        |Clear prompt                        |
|<kbd>CTRL-W</kbd>                                        |Remove backward word (sep by '/')   |
|<kbd>CTRL-H</kbd>,<kbd>BS</kbd>                          |Remove backward character           |
|<kbd>DEL</kbd>                                           |Delete character on the cursor      |
|<kbd>CTRL-Z</kbd>                                        |Toggle selection                    |
|<kbd>CTRL-R</kbd>                                        |Toggle fuzzy option                 |
|<kbd>CTRL-F</kbd>                                        |Toggle dirOnly option               |
|<kbd>CTRL-\\</kbd>                                       |Toggle curOnly option               |
|<kbd>CTRL-V</kbd>                                        |Toggle mruHist option               |
|<kbd>Enter</kbd>                                         |Decide : `open2edit` / `open2cd`    |
|<kbd>CTRL-D</kbd>,<kbd>CTRL-C</kbd>,<kbd>ESC</kbd>       |Cancel                              |

## License

MIT

## Author

- Yasuhiro Matsumoto (mattn) made `gof`.
- Shane.XB.Qian is simplifying `gof` to `jj`.

## Note

- mostly just a play to myself: only tested with linux/bash and edit default with vim.
- to simple cases should be just fine: quick to jump/edit & have more fun with bash :smile:
