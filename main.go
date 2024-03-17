package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"bufio"
	"os"
	"time"
)

const (
	GREEN  = "\033[92m"
	WHITE  = "\033[97m"
	RESET  = "\033[0m"
)

var sizeCache = make(map[string]int64)

func formatSize(sizeBytes int64) string {
	var unit string
	var value float64

	switch {
	case sizeBytes < 1024:
		unit = "B"
		value = float64(sizeBytes)
	case sizeBytes < 1024*1024:
		unit = "KB"
		value = float64(sizeBytes) / 1024
	case sizeBytes < 1024*1024*1024:
		unit = "MB"
		value = float64(sizeBytes) / (1024 * 1024)
	case sizeBytes < 1024*1024*1024*1024:
		unit = "GB"
		value = float64(sizeBytes) / (1024 * 1024 * 1024)
	default:
		unit = "TB"
		value = float64(sizeBytes) / (1024 * 1024 * 1024 * 1024)
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}

func getSize(startPath string) int64 {
	if size, ok := sizeCache[startPath]; ok {
		return size
	}

	var totalSize int64 = 0
	err := filepath.WalkDir(startPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking through the directory: %v\n", err)
		return 0
	}

	sizeCache[startPath] = totalSize
	return totalSize
}

func parseSize(sizeStr string) int64 {
	units := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	sizeStr = strings.ToUpper(strings.ReplaceAll(sizeStr, " ", ""))
	unit := sizeStr[len(sizeStr)-2:]
	value, err := strconv.ParseFloat(sizeStr[:len(sizeStr)-2], 64)
	if err != nil {
		fmt.Printf("Error parsing size: %v\n", err)
		return 0
	}
	return int64(value) * units[unit]
}

func printTree(startPath, prefix string, depth int, fileInclude, fileExclude []string, minSize, maxSize string, currentDepth int, stats map[string]int, onlyPath bool) {
	startTime := time.Now()
	if stats == nil {
		stats = map[string]int{"folders": 0, "files": 0}
	}

	if depth >= 0 && currentDepth > depth {
		return
	}

	sizeOfCurrentPath := getSize(startPath)
	formattedSize := formatSize(sizeOfCurrentPath)
	if prefix == "" {
		fmt.Printf("%s%s (%s)%s\n", GREEN, startPath, formattedSize, RESET)
	} else {
		fmt.Printf("%s├── %s%s (%s)%s\n", prefix, GREEN, filepath.Base(startPath), formattedSize, RESET)
	}

	prefix = strings.ReplaceAll(prefix, "├──", "│  ")
	prefix = strings.ReplaceAll(prefix, "└──", "   ")

	files, err := ioutil.ReadDir(startPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}
	for i, file := range files {
		path := filepath.Join(startPath, file.Name())
		if file.IsDir() {
			stats["folders"]++
			sizeInBytes := getSize(path)
			if (minSize == "" || sizeInBytes >= parseSize(minSize)) && (maxSize == "" || sizeInBytes <= parseSize(maxSize)) {
				if i == len(files)-1 {
					printTree(path, prefix+"└── ", depth, fileInclude, fileExclude, minSize, maxSize, currentDepth+1, stats, onlyPath)
				} else {
					printTree(path, prefix+"├── ", depth, fileInclude, fileExclude, minSize, maxSize, currentDepth+1, stats, onlyPath)
				}
			}
		} else if !onlyPath {
			stats["files"]++
			fileExtension := strings.TrimPrefix(filepath.Ext(file.Name()), ".")
			if contains(strings.ToLower(fileExtension), fileExclude) || (len(fileInclude) > 0 && !contains(strings.ToLower(fileExtension), fileInclude)) {
				continue
			}

			sizeInBytes := file.Size()
			formattedSize := formatSize(sizeInBytes)
			if (minSize == "" || sizeInBytes >= parseSize(minSize)) && (maxSize == "" || sizeInBytes <= parseSize(maxSize)) {
				if i == len(files)-1 {
					fmt.Printf("%s└── %s%s (%s)%s\n", prefix, WHITE, file.Name(), formattedSize, RESET)
				} else {
					fmt.Printf("%s├── %s%s (%s)%s\n", prefix, WHITE, file.Name(), formattedSize, RESET)
				}
			}
		}
	}

	if currentDepth == 0 {
		fmt.Println()
		fmt.Println("Summarize:")
		fmt.Printf("Scanned paths: %d\n", stats["folders"])
		fmt.Printf("Scanned files: %d\n", stats["files"])
		endTime := time.Now()
		fmt.Printf("Cost time: %.2f s\n", endTime.Sub(startTime).Seconds())
	}
}

func contains(item string, list []string) bool {
	for _, b := range list {
		if b == item {
			return true
		}
	}
	return false
}

func main() {
	var (
		startPath   string
		depth       int
		fileInclude stringSlice
		fileExclude stringSlice
		minSize     string
		maxSize     string
		onlyPath    bool
	)

	flag.StringVar(&startPath, "path", ".", "The starting path, defaults to the current directory")
	flag.IntVar(&depth, "depth", 999, "Depth of the directory tree, defaults is 999")
	flag.Var(&fileInclude, "file-include", "List of file types to include")
	flag.Var(&fileExclude, "file-exclude", "List of file types to exclude")
	flag.StringVar(&minSize, "min-size", "", "Minimum file size, defaults to none (unlimited), example: 10KB, 20MB")
	flag.StringVar(&maxSize, "max-size", "", "Maximum file size, defaults to none (unlimited), example: 10KB, 20MB")
	flag.BoolVar(&onlyPath, "only-path", false, "Whether to only print paths, defaults to False")
	flag.Parse()

	stats := map[string]int{"folders": 0, "files": 0}
	printTree(startPath, "", depth, fileInclude, fileExclude, minSize, maxSize, 0, stats, onlyPath)

	fmt.Println("Press Enter to quit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprint(*i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}
