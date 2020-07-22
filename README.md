# jj

## Feature

* similar & forked from `gof` - thanks `mattn` !
* but make it be simple: just for independent `open2edit` or `open2cd` !

## Installation

- `go get` from this repo
- pls think to write `alias jj='jj || if [ $? -eq 6 ]; then $(cat "$HOME/.jj_tmp_fs"); fi'` into your `.bashrc`

## Usage

```sh
$ jj
```

* Run `jj` and type `ctrl-o`, then start to edit (`open2edit`) with `vim`, whatever was a file or dir.
* Run `jj` and type `enter`, then start to edit (`open2edit`) with `vim` -if was a file, or quit but cd to that dir (`open2cd`) -if was a dir.

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
|<kbd>Enter</kbd>                                         |Decide (`open2edit` or `open2cd`)   |
|<kbd>CTRL-D</kbd>,<kbd>CTRL-C</kbd>,<kbd>ESC</kbd>       |Cancel                              |

## Options

|Option        |Description                      |
|--------------|---------------------------------|
|-f            |Fuzzy match (warn: maybe slow)   |
|-d [path]     |Specify root directory           |

## License

MIT

## Author

Shane.XB.Qian based on `gof` by:
Yasuhiro Matsumoto (a.k.a mattn)

## NOTE

mostly just a play to myself: only tested linux/bash and default with vim.
but to simple cases: it should be good / tricky to quickly jump and edit like some shell or sth do ..
