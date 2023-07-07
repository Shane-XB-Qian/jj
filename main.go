// go

// ------------------------------------------------------------------------
// author      : © Shane.Qian@foxmail.com
// createdat   : tue 17 nov 2020 02:30:00 am hkt @china
// ------------------------------------------------------------------------
// description : quick to jump -and- quick to edit!
// ------------------------------------------------------------------------

package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"github.com/saracen/walker"

	fuzzysearch "github.com/lithammer/fuzzysearch/fuzzy"
)

const name = "jj"

const version = "0.0.7"

var revision = "HEAD"

type matched struct {
	name     string
	pos1     int
	pos2     int
	pseq     int
	selected bool
	index    int
}

var (
	cwd              = ""
	root             = ""
	input            = []rune{}
	files            []string
	selected         = []string{}
	heading          = false
	current          []matched
	cursorX, cursorY int
	offset           int
	width, height    int
	mutex            sync.Mutex
	dirty            = false
	duration         = 20 * time.Millisecond
	timer            *time.Timer
	scanning         = 0
	ignorere         *regexp.Regexp
	fuzzy            bool
	dirOnly          bool
	curOnly          bool
	mruHist          bool
)

const (
	mruMax   = 666
	mruStore = ".jj_mru_fs"
	mruTmpFs = ".jj_tmp_fs"
)

func fuzzyFilterFlag() string {
	if fuzzy {
		return "Y"
	}
	return "n"
}

func dirOnlyFilterFlag() string {
	if dirOnly {
		return "Y"
	}
	return "n"
}

func curOnlyFilterFlag() string {
	if curOnly {
		return "Y"
	}
	return "n"
}

func mruHistFlag() string {
	if mruHist {
		return "Y"
	}
	return "n"
}

func getHomeDir() string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		// shane: better return nil (vs exit) instead ?
	}
	return userHome
}

func readFs(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		// fmt.Fprintln(os.Stderr, err)
		// os.Exit(1)
		return nil
		// shane: mostly failed if not existed ?
	}
	defer f.Close()
	tmp := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// fmt.Println(sc.Text())
		tmp = append(tmp, sc.Text())
	}
	if err := sc.Err(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		// os.Exit(1)
		return nil
		// shane: somehow failed - return nil whatever ?
	}
	return tmp
}

func filesIfMru() []string {
	// mutex.Lock()
	// defer mutex.Unlock()
	if !mruHist {
		return files
	}
	tmp := readFs(getHomeDir() + "/" + mruStore)
	tmpLen := 0
	if tmp != nil {
		tmpLen = len(tmp)
	}
	// shane: to make last mru item showed at the list bottom.
	tmp2 := []string{}
	if tmpLen > 0 {
		for i := tmpLen - 1; i >= 0; i-- {
			// shane: to make mru list showed as rel path of cwd ?
			// if rl_p, err := filepath.Rel(cwd, tmp[i]); err == nil {
			// 	tmp2 = append(tmp2, rl_p)
			// } else {
			// 	tmp2 = append(tmp2, tmp[i])
			// }
			tmp2 = append(tmp2, tmp[i])
		}
	} else {
		return nil
	}
	return tmp2
}

func filterIfDirOnly(fs []string) []string {
	// mutex.Lock()
	// defer mutex.Unlock()
	if !dirOnly {
		return fs
	}
	// fs := files
	tmp := []string{}
	for _, f := range fs {
		fi, err := os.Stat(f)
		if err != nil {
			// fmt.Fprintln(os.Stderr, err)
			// os.Exit(1)
			continue
			// XXX: may fail - just ignore such ?
			// shane: mostly looks due to auth ?!
		}
		if fi.IsDir() {
			tmp = append(tmp, f)
		}
	}
	return tmp
}

func filterIfCurOnly(fs []string) []string {
	// mutex.Lock()
	// defer mutex.Unlock()
	if !curOnly {
		return fs
	}
	// fs := files
	var fl []os.DirEntry
	var ft string
	if root != "" {
		fl, _ = os.ReadDir(cwd)
		ft = cwd + "/"
	} else {
		fl, _ = os.ReadDir(".")
		ft = ""
	}
	tmp := []string{}
	for _, f := range fl {
		if f.IsDir() {
			tmp = append(tmp, ft+f.Name()+"/")
		} else {
			tmp = append(tmp, ft+f.Name())
		}
	}
	return tmp
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func tprint(x, y int, fg, bg termbox.Attribute, msg string) {
	// for _, c := range []rune(msg) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func tprintf(x, y int, fg, bg termbox.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	tprint(x, y, fg, bg, s)
}

func filter(fuzzy bool) {
	mutex.Lock()
	fs := filesIfMru()
	fs = filterIfCurOnly(fs)
	fs = filterIfDirOnly(fs)
	inp := input
	sel := selected
	mutex.Unlock()

	var tmp []matched
	if len(inp) == 0 {
		tmp = make([]matched, len(fs))
		for n, f := range fs {
			prevSelected := false
			for _, s := range sel {
				if f == s {
					prevSelected = true
					break
				}
			}
			tmp[n] = matched{
				name:     f,
				pos1:     -1,
				pos2:     -1,
				pseq:     -1,
				selected: prevSelected,
				index:    n,
			}
		}
	} else if fuzzy {
		tmp = make([]matched, 0, len(fs))
		inpl := strings.ToLower(string(inp))
		inpl_s := strings.ToLower(string(inp[0]))
		inpl_e := strings.ToLower(string(inp[len(inp)-1]))
		for _, f := range fs {
			var pos int
			var pos_s int
			var pos_e int
			if lf := strings.ToLower(f); len(f) == len(lf) {
				pos = strings.Index(lf, inpl)
				pos_s = strings.Index(lf, inpl_s)
				pos_e = strings.LastIndex(lf, inpl_e)
			} else {
				pos = bytes.Index([]byte(f), []byte(string(inp)))
				pos_s = bytes.Index([]byte(f), []byte(string(inp[0])))
				pos_e = bytes.LastIndex([]byte(f), []byte(string(inp[len(inp)-1])))
			}
			if pos_s == -1 || pos_e == -1 {
				continue
			}
			if !fuzzysearch.MatchNormalizedFold(string(inp), f) {
				continue
			}
			prevSelected := false
			for _, s := range sel {
				if f == s {
					prevSelected = true
					break
				}
			}
			poss := -1
			if pos != -1 {
				poss = len([]rune(f[:pos]))
			}
			pos1 := len([]rune(f[:pos_s]))
			pos2 := len([]rune(f[:pos_e]))
			tmp = append(tmp, matched{
				name:     f,
				pos1:     pos1,
				pos2:     pos2 + 1,
				pseq:     poss,
				selected: prevSelected,
				index:    len(tmp),
			})
		}
	} else {
		tmp = make([]matched, 0, len(fs))
		inpl := strings.ToLower(string(inp))
		for _, f := range fs {
			var pos int
			if lf := strings.ToLower(f); len(f) == len(lf) {
				pos = strings.Index(lf, inpl)
			} else {
				pos = bytes.Index([]byte(f), []byte(string(inp)))
			}
			if pos == -1 {
				continue
			}
			prevSelected := false
			for _, s := range sel {
				if f == s {
					prevSelected = true
					break
				}
			}
			pos1 := len([]rune(f[:pos]))
			tmp = append(tmp, matched{
				name:     f,
				pos1:     pos1,
				pos2:     pos1 + len(inp),
				selected: prevSelected,
				index:    len(tmp),
			})
		}
	}
	if len(inp) > 0 {
		sort.Slice(tmp, func(i, j int) bool {
			li, lj := tmp[i].pos2-tmp[i].pos1, tmp[j].pos2-tmp[j].pos1
			if fuzzy {
				pi, pj := tmp[i].pseq, tmp[j].pseq
				if pi == -1 {
					pi = 999999
				}
				if pj == -1 {
					pj = 999999
				}
				return pi < pj || pi == pj && tmp[i].index < tmp[j].index
			} else {
				return li < lj || li == lj && tmp[i].index < tmp[j].index
			}
		})
	}

	mutex.Lock()
	defer mutex.Unlock()
	current = tmp
	selected = sel
	if cursorY < 0 {
		cursorY = 0
	}
	if cursorY >= len(current) {
		cursorY = len(current) - 1
	}
	if cursorY >= 0 && height > 0 {
		if cursorY < offset {
			offset = cursorY
		} else if offset < cursorY-(height-3) {
			offset = cursorY - (height - 3)
		}
		if len(current)-(height-3)-1 < offset {
			offset = len(current) - (height - 3) - 1
			if offset < 0 {
				offset = 0
			}
		}
	} else {
		offset = 0
	}
}

func drawLines() {
	defer func() {
		recover()
	}()
	mutex.Lock()
	defer mutex.Unlock()

	width, height = termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for n := offset; n <= height-3+offset; n++ {
		if n >= len(current) {
			break
		}
		x, y, w := 2, height-3-(n-offset), 0
		name := current[n].name
		pos1 := current[n].pos1
		pos2 := current[n].pos2
		pseq := current[n].pseq
		selected := current[n].selected
		if pos1 >= 0 {
			pwidth := runewidth.StringWidth(string([]rune(current[n].name)[0:pos1]))
			if !heading && pwidth > width/2 {
				rname := []rune(name)
				wwidth := 0
				for i := 0; i < len(rname); i++ {
					w = runewidth.RuneWidth(rname[i])
					if wwidth+w > width/2 {
						name = "..." + string(rname[i:])
						pos1 -= i - 3
						pos2 -= i - 3
						pseq -= i - 3
						break
					}
					wwidth += w
				}
			}
		}
		rwidth := len([]rune(name))
		rinput := len([]rune(input))
		if rwidth+2 > width {
			name = string([]rune(name)[0:width-5]) + "..."
		}
		for f, c := range []rune(name) {
			w = runewidth.RuneWidth(c)
			if x+w > width {
				break
			}
			if fuzzy && rinput > 0 && pseq > -1 && pseq <= f && f < pseq+rinput {
				if selected {
					termbox.SetCell(x, y, c, termbox.ColorMagenta|termbox.AttrBold, termbox.ColorDefault)
				} else if cursorY == n {
					termbox.SetCell(x, y, c, termbox.ColorMagenta|termbox.AttrBold|termbox.AttrUnderline, termbox.ColorDefault)
				} else {
					termbox.SetCell(x, y, c, termbox.ColorMagenta|termbox.AttrBold, termbox.ColorDefault)
				}
			} else if pos1 <= f && f < pos2 {
				if selected {
					termbox.SetCell(x, y, c, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault)
				} else if cursorY == n {
					termbox.SetCell(x, y, c, termbox.ColorYellow|termbox.AttrBold|termbox.AttrUnderline, termbox.ColorDefault)
				} else {
					termbox.SetCell(x, y, c, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault)
				}
			} else {
				if selected {
					termbox.SetCell(x, y, c, termbox.ColorRed, termbox.ColorDefault)
				} else if cursorY == n {
					termbox.SetCell(x, y, c, termbox.ColorYellow|termbox.AttrUnderline, termbox.ColorDefault)
				} else {
					termbox.SetCell(x, y, c, termbox.ColorDefault, termbox.ColorDefault)
				}
			}
			x += w
		}
	}
	if cursorY >= 0 {
		tprint(0, height-3-(cursorY-offset), termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault, "> ")
	}
	if scanning >= 0 {
		tprint(0, height-2, termbox.ColorGreen, termbox.ColorDefault, string([]rune("-\\|/")[scanning%4]))
		scanning++
	}
	if mruHist {
		tprintf(2, height-2, termbox.ColorDefault, termbox.ColorDefault, "%d|%d (%d)z { [%s]r [%s]f } { [%s]v < [%s]\\ }", len(current), mruMax, len(selected), "fuzzy:"+fuzzyFilterFlag(), "dirOnly:"+dirOnlyFilterFlag(), "mruHist:(abs):"+mruHistFlag(), "curOnly:"+curOnlyFilterFlag())
	} else {
		tprintf(2, height-2, termbox.ColorDefault, termbox.ColorDefault, "%d/%d (%d)z { [%s]r [%s]f } { [%s]v < [%s]\\ }", len(current), len(files), len(selected), "fuzzy:"+fuzzyFilterFlag(), "dirOnly:"+dirOnlyFilterFlag(), "mruHist:(abs):"+mruHistFlag(), "curOnly:"+curOnlyFilterFlag())
	}
	tprint(0, height-1, termbox.ColorBlue|termbox.AttrBold, termbox.ColorDefault, "> ")
	tprint(2, height-1, termbox.ColorDefault|termbox.AttrBold, termbox.ColorDefault, string(input))
	termbox.SetCursor(2+runewidth.StringWidth(string(input[0:cursorX])), height-1)
	termbox.Flush()
}

// var actionKeys = []termbox.Key{
// 	termbox.KeyCtrlA,
// 	termbox.KeyCtrlB,
// 	termbox.KeyCtrlC,
// 	termbox.KeyCtrlD,
// 	termbox.KeyCtrlE,
// 	termbox.KeyCtrlF,
// 	termbox.KeyCtrlG,
// 	termbox.KeyCtrlH,
// 	termbox.KeyCtrlI,
// 	termbox.KeyCtrlJ,
// 	termbox.KeyCtrlK,
// 	termbox.KeyCtrlL,
// 	termbox.KeyCtrlM,
// 	termbox.KeyCtrlN,
// 	termbox.KeyCtrlO,
// 	termbox.KeyCtrlP,
// 	termbox.KeyCtrlQ,
// 	termbox.KeyCtrlR,
// 	termbox.KeyCtrlS,
// 	termbox.KeyCtrlT,
// 	termbox.KeyCtrlU,
// 	termbox.KeyCtrlV,
// 	termbox.KeyCtrlW,
// 	termbox.KeyCtrlX,
// 	termbox.KeyCtrlY,
// 	termbox.KeyCtrlZ,
// }

func listFiles(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	n := 0
	cb := walker.WithErrorCallback(func(pathname string, err error) error {
		return nil
	})
	fn := func(path string, info os.FileInfo) error {
		path = filepath.Clean(path)
		if p, err := filepath.Rel(cwd, path); err == nil {
			path = p
		}
		if path == "." {
			return nil
		}
		base := filepath.Base(path)
		if ignorere != nil && ignorere.MatchString(base) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		path = filepath.ToSlash(path)
		// shane: let dir end with Separator -good for identification.
		if info.IsDir() {
			if path[len(path)-1] != filepath.Separator {
				path = path + string(filepath.Separator)
			}
		}
		mutex.Lock()
		files = append(files, path)
		n++
		if n%1000 == 0 {
			dirty = true
			timer.Reset(duration)
		}
		mutex.Unlock()
		return nil
	}
	walker.WalkWithContext(ctx, cwd, fn, cb)
	mutex.Lock()
	dirty = true
	timer.Reset(duration)
	scanning = -1
	mutex.Unlock()
}

func main() {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fo, err := os.Stdout.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if strings.Index(fi.Mode().String(), "p") == 0 || strings.Index(fo.Mode().String(), "p") == 0 {
		fmt.Fprintln(os.Stderr, "err: 'jj' is simply going to work only as an Independent cmd! - shane.")
		os.Exit(1)
	}
	if e := env("SHELL", ""); e == "" {
		fmt.Fprintln(os.Stderr, "err: 'jj' is simply going to work on 'Linux', perhaps 'Mac', not sure 'Windows'. - shane.")
		os.Exit(1)
	} else {
		b, err := regexp.MatchString("bash", e)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !b {
			fmt.Fprintln(os.Stderr, "wrn: 'jj' is simply going to work with 'Bash', not sure others. - shane.")
			// os.Exit(1)
		}
	}

	var open2edit bool
	var open2CdOrEdit bool

	// var fuzzy bool
	var ignore string
	var showVersion bool
	var showHelp bool

	flag.BoolVar(&fuzzy, "f", false, "Fuzzy match")
	flag.BoolVar(&dirOnly, "w", false, "Init with dir only on")
	flag.BoolVar(&curOnly, "c", false, "Init with cur only on")
	flag.BoolVar(&mruHist, "m", false, "Init with mru hist on")
	flag.StringVar(&root, "d", "", "Root directory")
	flag.StringVar(&ignore, "i", env(`JJ_IGNORE_PATTERN`, `^(\.git|\.hg|\.svn|_darcs|\.bzr)$`), "Ignore pattern")
	flag.BoolVar(&showVersion, "v", false, "Print the version")
	flag.BoolVar(&showHelp, "h", false, "Print usage")
	flag.Parse()

	if showHelp {
		flag.Usage()
		fmt.Println("---------------------------------")
		fmt.Println("- quick to jump & quick to edit -")
		fmt.Println("- shane.qian@foxmail.com @china -")
		return
	}

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	defer colorable.EnableColorsStdout(nil)()

	// Make regular expression pattern to ignore files.
	if ignore != "" {
		ignorere, err = regexp.Compile(ignore)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Make sure current directory.
	if root == "" {
		cwd, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if runtime.GOOS == "windows" && strings.HasPrefix(root, "/") {
			cwd, _ = os.Getwd()
			cwd = filepath.Join(filepath.VolumeName(cwd), root)
		} else {
			cwd, err = filepath.Abs(root)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		st, err := os.Stat(cwd)
		if err == nil && !st.IsDir() {
			err = fmt.Errorf("directory not found: %s", cwd)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = os.Chdir(cwd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	// shane: https://github.com/mattn/gof/issues/36
	cwd, err = filepath.EvalSymlinks(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	redrawFunc := func() {
		mutex.Lock()
		d := dirty
		mutex.Unlock()
		if d {
			filter(fuzzy)
			drawLines()

			mutex.Lock()
			dirty = false
			mutex.Unlock()
		} else {
			drawLines()
		}
	}
	timer = time.AfterFunc(0, redrawFunc)
	timer.Stop()

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	// Walk and collect files recursively.
	go listFiles(ctx, &wg)

	err = termbox.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	termbox.SetInputMode(termbox.InputEsc)

	if mruHist {
		mutex.Lock()
		dirty = true
		mutex.Unlock()
	}
	redrawFunc()

loop:
	for {
		update := false

		// Polling key events
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlD, termbox.KeyCtrlC:
				termbox.Close()
				os.Exit(1)
			case termbox.KeyHome, termbox.KeyCtrlA:
				cursorX = 0
			case termbox.KeyEnd, termbox.KeyCtrlE:
				cursorX = len(input)
			case termbox.KeyEnter, termbox.KeyCtrlO:
				if cursorY >= 0 && cursorY < len(current) {
					if len(selected) == 0 {
						selected = append(selected, current[cursorY].name)
					}
					if ev.Key == termbox.KeyEnter {
						open2CdOrEdit = true
						break loop
					}
					open2edit = true
					break loop
				}
			case termbox.KeyArrowLeft:
				if cursorX > 0 {
					cursorX--
				}
			case termbox.KeyArrowRight:
				if cursorX < len([]rune(input)) {
					cursorX++
				}
			case termbox.KeyArrowUp, termbox.KeyCtrlK, termbox.KeyCtrlP:
				if cursorY < len(current)-1 {
					cursorY++
					if offset < cursorY-(height-3) {
						offset = cursorY - (height - 3)
					}
				} else if cursorY == len(current)-1 {
					cursorY = 0
					if cursorY < offset {
						offset = cursorY
					}
				}
			case termbox.KeyArrowDown, termbox.KeyCtrlJ, termbox.KeyCtrlN:
				if cursorY > 0 {
					cursorY--
					if cursorY < offset {
						offset = cursorY
					}
				} else if cursorY == 0 {
					cursorY = len(current) - 1
					if offset < cursorY-(height-3) {
						offset = cursorY - (height - 3)
					}
				}
			case termbox.KeyCtrlG:
				cursorY = 0
				if cursorY < offset {
					offset = cursorY
				}
			case termbox.KeyCtrlT:
				heading = !heading
			case termbox.KeyTab:
				if cursorY >= 0 && cursorY < len(current) {
					dirpath, _ := filepath.Split(current[cursorY].name)
					if dirpath != "" {
						input = []rune(dirpath)
					} else {
						input = []rune(current[cursorY].name)
					}
					cursorX = len(input)
					// shane: refresh 'current' & reflect 'cursorY'
					filter(fuzzy)
					for i, s := range current {
						if s.name == string(input) {
							cursorY = i
							break
						}
					}
					update = true
					// shane: 'filter()' would recalculate 'offset'
				}
			case termbox.KeyCtrlY:
				if cursorY >= 0 && cursorY < len(current) {
					input = []rune(current[cursorY].name)
					cursorX = len(input)
					update = true
				}
			case termbox.KeyCtrlL:
				update = true
			case termbox.KeyCtrlU:
				cursorX = 0
				input = []rune{}
				update = true
			case termbox.KeyCtrlW:
				inputLenKpt := len(input)
				// sq: remain tail '/'
				cursorTailMoved := false
				if len(input) != 0 && len(input) == cursorX && input[cursorX-1] == filepath.Separator {
					cursorX = cursorX - 1
					cursorTailMoved = true
				}
				head := string(input[0:cursorX])
				tail := input[cursorX:]
				// sq: `\s+|`+string(filepath.Separator)
				// sq: or just string(filepath.Separator)
				posAll := regexp.MustCompile(string(filepath.Separator)).FindAllStringIndex(head, -1)
				pos := []int{}
				if len(posAll) > 0 {
					pos = posAll[len(posAll)-1]
				}
				if len(pos) > 0 && pos[len(pos)-1] > 0 {
					input = []rune(head[0 : pos[len(pos)-1]-1])
				} else {
					input = []rune{}
				}
				input = append(input, tail...)
				cursorX = cursorX - (inputLenKpt - len(input))
				// sq: recover cursor pos if tail was '/'
				if cursorTailMoved {
					cursorX = cursorX + 1
				}
				update = true
			case termbox.KeyCtrlZ:
				if cursorY >= 0 && cursorY < len(current) {
					found := -1
					name := current[cursorY].name
					for i, s := range selected {
						if name == s {
							found = i
							break
						}
					}
					if found == -1 {
						selected = append(selected, current[cursorY].name)
					} else {
						selected = append(selected[:found], selected[found+1:]...)
					}
					update = true
				}
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if cursorX > 0 {
					input = append(input[0:cursorX-1], input[cursorX:]...)
					cursorX--
					update = true
				}
			case termbox.KeyDelete:
				if cursorX < len([]rune(input)) {
					input = append(input[0:cursorX], input[cursorX+1:]...)
					update = true
				}
			case termbox.KeyCtrlR:
				fuzzy = !fuzzy
				update = true
			case termbox.KeyCtrlF:
				dirOnly = !dirOnly
				update = true
			case termbox.KeyCtrlBackslash:
				curOnly = !curOnly
				update = true
			case termbox.KeyCtrlV:
				mruHist = !mruHist
				update = true
			default:
				if ev.Key == termbox.KeySpace {
					ev.Ch = ' '
				}
				if ev.Ch > 0 {
					out := []rune{}
					out = append(out, input[0:cursorX]...)
					out = append(out, ev.Ch)
					input = append(out, input[cursorX:]...)
					cursorX++
					update = true
				}
			}
		case termbox.EventError:
			update = false
		}

		// If need to update, start timer
		if scanning != -1 {
			if update {
				mutex.Lock()
				dirty = true
				timer.Reset(duration)
				mutex.Unlock()
			} else {
				timer.Reset(1)
			}
		} else {
			if update {
				filter(fuzzy)
			}
			drawLines()
		}
	}
	timer.Stop()

	// Request terminating
	cancel()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Close()

	wg.Wait()

	if len(selected) == 0 {
		os.Exit(1)
	}

	fArg := []string{}
	// mruHist stored Abs path.
	if root != "" && !mruHist {
		for _, f := range selected {
			fArg = append(fArg, filepath.Join(root, f))
			// XXX: should use 'cwd' instead of 'root' ?
		}
	} else {
		fArg = append(fArg, selected...)
	}

	mruListTmp := readFs(getHomeDir() + "/" + mruStore)
	mruListTmpLen := len(mruListTmp)
	if mruListTmpLen > mruMax {
		mruListTmp = mruListTmp[(mruListTmpLen - mruMax):mruListTmpLen]
	}

	for _, fa := range fArg {
		faAbs, err := filepath.Abs(fa)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		tmpN := []int{}
		for iN, iF := range mruListTmp {
			if strings.TrimSpace(iF) == strings.TrimSpace(faAbs) {
				tmpN = append(tmpN, iN)
			}
		}
		if len(tmpN) > 0 {
			for i := len(tmpN) - 1; i >= 0; i-- {
				// shane: remove the dups starting from tail.
				mruListTmp = append(mruListTmp[:tmpN[i]], mruListTmp[(tmpN[i]+1):]...)
			}
		}
		mruListTmp = append(mruListTmp, faAbs)
	}

	f, err := os.OpenFile(getHomeDir()+"/"+mruStore, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0600)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, tmpFS := range mruListTmp {
		_, err = f.WriteString(tmpFS + "\n")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	f.Close()

	// shane: just care last one -if to cd -if multiple selected.
	fs := fArg[len(fArg)-1]
	fi, err = os.Stat(fs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if open2edit || (open2CdOrEdit && !fi.IsDir()) {
		cmd := exec.Command("vim", fArg...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		// f, err := os.Create(getHomeDir()+"/"+mruTmpFs)
		f, err := os.OpenFile(getHomeDir()+"/"+mruTmpFs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_, err = f.WriteString("cd " + fs)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		f.Close()
		os.Exit(6)
	}

	os.Exit(0)
}
