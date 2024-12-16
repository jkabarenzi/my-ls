package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"
)

type fileInfoSlice []fs.FileInfo

const (
	resetColor   = "\033[0m"
	dirColor     = "\033[1;34m"
	symlinkColor = "\033[1;36m"
)

func (f fileInfoSlice) Len() int {
	return len(f)
}

func (f fileInfoSlice) Less(i, j int) bool {
	return f[i].Name() < f[j].Name()
}

func (f fileInfoSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func getColor(file fs.FileInfo) string {
	if file.IsDir() {
		return dirColor
	}
	if file.Mode()&fs.ModeSymlink != 0 {
		return symlinkColor
	}
	return resetColor
}

func list(path string, r, l, a, rev, t bool) {
	d, e := os.ReadDir(path)
	if e != nil {
		fmt.Println("Error reading directory:", e)
		return
	}
	var f fileInfoSlice
	for _, e := range d {
		if !a && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			fmt.Println("Error getting file info:", err)
			continue
		}
		f = append(f, info)
	}
	if t {
		sort.Slice(f, func(i, j int) bool {
			return f[i].ModTime().Before(f[j].ModTime())
		})
	} else {
		sort.Sort(f)
	}
	if rev {
		sort.Sort(sort.Reverse(f))
	}
	for _, file := range f {
		color := getColor(file)
		if l {
			stat, ok := file.Sys().(*syscall.Stat_t)
			if !ok {
				fmt.Println("Error getting file stats")
				return
			}
			fmt.Printf("%s%+v %3d %9d %9d %12d %s %s%s\n",
				color,
				file.Mode(),
				stat.Nlink,
				stat.Uid,
				stat.Gid,
				file.Size(),
				file.ModTime().Format(time.Stamp),
				file.Name(),
				resetColor)
		} else {
			fmt.Printf("%s%s\t%s", color, file.Name(), resetColor)
		}
		if r && file.IsDir() {
			fmt.Println()
			list(fmt.Sprintf("%s/%s", path, file.Name()), r, l, a, rev, t)
		}
	}
	fmt.Println()
}

func main() {
	r := flag.Bool("R", false, "List directories recursively")
	l := flag.Bool("l", false, "Use long listing format")
	a := flag.Bool("a", false, "Show hidden files")
	rev := flag.Bool("r", false, "Reverse order while sorting")
	t := flag.Bool("t", false, "Sort by modification time")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		list(".", *r, *l, *a, *rev, *t)
	} else {
		for _, arg := range args {
			// Check if the argument is a directory or a file
			fileInfo, err := os.Stat(arg)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if fileInfo.IsDir() {
				list(arg, *r, *l, *a, *rev, *t)
			} else {
				color := getColor(fileInfo)
				if *l {
					stat, ok := fileInfo.Sys().(*syscall.Stat_t)
					if !ok {
						fmt.Println("Error getting file stats")
						return
					}
					fmt.Printf("%s%+v %3d %9d %9d %12d %s %s%s\n",
						color,
						fileInfo.Mode(),
						stat.Nlink,
						stat.Uid,
						stat.Gid,
						fileInfo.Size(),
						fileInfo.ModTime().Format(time.Stamp),
						fileInfo.Name(),
						resetColor)
				} else {
					fmt.Printf("%s%s\t%s", color, fileInfo.Name(), resetColor)
				}
				fmt.Println()
			}
		}
	}
}
