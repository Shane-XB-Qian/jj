# jj

## Feature

* similar & forked from [gof](https://github.com/mattn/gof.git) - thanks `mattn`.
* but simplify: just for independent `open2edit` or `open2cd`.
* or at least `jj` should be a more quick / cool name .. :)

## Installation

- `go get github.com/Shane-XB-Qian/jj` from this repo.
- pls think to write `alias jj='jj || if [ $? -eq 6 ]; then $(cat "$HOME/.jj_tmp_fs"); fi'` into your `.bashrc`.

## Usage

```sh
$ jj
```

* Run `jj` and type `ctrl-o`, then start to edit `open2edit` with `vim`, whatever was a file or dir.
* Run `jj` and type `enter`, then start to edit `open2edit` with `vim` if was a file, Or quit but `cd` to that dir `open2cd` if was a dir.

## Keyboard shortcuts

|Key                                                      |Description                         |
|---------------------------------------------------------|------------------------------------|
|<kbd>CTRL-K</kbd>,<kbd>CTRL-P</kbd>,<kbd>ARROW-UP</kbd>  |Move-up line                        |
|<kbd>CTRL-J</kbd>,<kbd>CTRL-N</kbd>,<kbd>ARROW-DOWN</kbd>|Move-down line                      |
|<kbd>CTRL-A</kbd>,<kbd>HOME</kbd>                        |Go to head of prompt                |
|<kbd>CTRL-E</kbd>,<kbd>END</kbd>                         |Go to trail of prompt               |
|<kbd>ARROW-LEFT</kbd>                                    |Move-left cursor                    |
|<kbd>ARROW-RIGHT</kbd>                                   |Move-right cursor                   |
|<kbd>CTRL-O</kbd>                                        |Edit the selected file/dir          |
|<kbd>CTRL-I</kbd>                                        |Toggle view header/trailing of lines|
|<kbd>CTRL-L</kbd>                                        |Redraw                              |
|<kbd>CTRL-U</kbd>                                        |Clear prompt                        |
|<kbd>CTRL-W</kbd>                                        |Remove backward word                |
|<kbd>BS</kbd>                                            |Remove backward character           |
|<kbd>DEL</kbd>                                           |Delete character on the cursor      |
|<kbd>CTRL-Z</kbd>                                        |Toggle selection                    |
|<kbd>CTRL-R</kbd>                                        |Toggle fuzzy option                 |
|<kbd>CTRL-F</kbd>                                        |Toggle dirOnly option               |
|<kbd>CTRL-V</kbd>                                        |Toggle mruHist option               |
|<kbd>Enter</kbd>                                         |Decide : `open2edit` / `open2cd`    |
|<kbd>CTRL-D</kbd>,<kbd>CTRL-C</kbd>,<kbd>ESC</kbd>       |Cancel                              |

## Options

|Option        |Description                      |
|--------------|---------------------------------|
|-f            |Fuzzy match (warn: maybe slow)   |
|-d [path]     |Specify root directory           |

- `Option` if ran from `alias` would Not work, but mostly cur pwd should be just fine.
- `fuzzy` filter can be switched when in `jj` by `ctrl-r`, but warn: fuzzy maybe slow.
- `dirOnly` btw: if to display/show dir only can be switched when in `jj` by `ctrl-f`.
- `mruHist` btw: if to display/show mru hist can be switched when in `jj` by `ctrl-v`.
- `mruHist` btw: hist of mru /from newer (bottom) to older/ - may had dup items in it.

## License

MIT

## Author

- Yasuhiro Matsumoto (mattn) made `gof`.
- Shane.XB.Qian is simplifying `gof` to `jj`.

## NOTE

- mostly just a play to myself: only tested linux/bash and edit default with vim.
- to simple cases should be just fine - quickly to jump/edit like some shell do..
