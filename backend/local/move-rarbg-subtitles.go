package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RARBG's 3/4/5_English.srt mapping doesn't appear to be consistent,
// so assuming SDH is always provided and largest, smallest is forced
var subtitleTypeMap = map[int]string{
	0: ".en.sdh",
	1: ".en",
	2: ".en.forced",
}

const (
	tvRoot     = "G:\\TV Shows"
	moviesRoot = "G:\\Movies"
	mp4Ext     = ".mp4"
	mkvExt     = ".mkv"
)

func main() {
	root := tvRoot
	var pathsToDelete []string

	cleanUpNullFiles(root)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "Subs" {
			err = handleSubsDir(path)
			if err == nil {
				pathsToDelete = append(pathsToDelete, path)
			}
		}
		return nil
	})
	log.Printf("donezo. final err: %v", err)
	if err == nil {
		log.Printf("finna delete: \n%v", strings.Join(pathsToDelete, "\n"))
		//for _, path := range pathsToDelete {
		//	err := os.RemoveAll(path)
		//	if err != nil {
		//		log.Printf("failed to delete %v: %v", path, err)
		//	}
		//}
	}
}

func handleSubsDir(subsRoot string) error {
	if strings.HasPrefix(subsRoot, tvRoot) {
		return handleTvSubsDir(subsRoot)
	} else if strings.HasPrefix(subsRoot, moviesRoot) {
		return handleMovieSubsDir(subsRoot)
	}
	return nil
}

func handleMovieSubsDir(subsRoot string) error {
	// e.g. G:\Movies\Kimi.2022.1080p.WEBRip.x265-RARBG\Subs
	files, err := os.ReadDir(getTargetDir(subsRoot))
	if err != nil {
		return err
	}
	var targetVideoFile string
	for _, file := range files {
		if filepath.Ext(file.Name()) == mp4Ext {
			targetVideoFile = strings.TrimSuffix(file.Name(), mp4Ext)
		} else if filepath.Ext(file.Name()) == mkvExt {
			targetVideoFile = strings.TrimSuffix(file.Name(), mkvExt)
		}
	}
	copySubs(subsRoot, targetVideoFile)
	return nil
}

func handleTvSubsDir(subsRoot string) error {
	// e.g. G:\TV Shows\Better Call Saul\Better.Call.Saul.S05.1080p.BluRay.x265-RARBG\Subs
	err := filepath.WalkDir(subsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() != "Subs" {
			copySubs(path, d.Name())
		}
		return nil
	})
	return err
}

func getTargetDir(path string) string {
	return strings.Split(path+string(os.PathSeparator), string(os.PathSeparator)+"Subs"+string(os.PathSeparator))[0]
}

func copySubs(subsPath, targetVideoFile string) {
	if targetVideoFile == "" {
		log.Printf("no target video file for path: %v", subsPath)
		return
	}

	// log.Printf("handling subs dir: %s", subsPath)
	subtitles := getSortedEnglishSubs(subsPath)
	for i, sub := range subtitles {
		oldPath := filepath.Join(subsPath, sub.Name())
		newPath := filepath.Join(getTargetDir(subsPath),
			targetVideoFile+determineSubtitleType(sub.Name(), i, len(subtitles)),
		)

		if areFilesEqual(oldPath, newPath) {
			//log.Printf("files are equal, not copying: %v", oldPath)
		} else {
			info, _ := sub.Info()
			log.Printf("copying (size %v) %v \n\tto %v", info.Size(), oldPath, newPath)
			copyFile(oldPath, newPath)
		}
	}
	// log.Printf("finished %v \n\n", subsPath)
}

func copyFile(src string, dest string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func areFilesEqual(src, dest string) bool {
	srcBytes, _ := os.ReadFile(src)
	destBytes, _ := os.ReadFile(dest)
	return bytes.Equal(srcBytes, destBytes)
}

func getSortedEnglishSubs(path string) []os.DirEntry {
	var englishSubs []os.DirEntry
	subtitles, _ := os.ReadDir(path)
	for _, sub := range subtitles {
		if strings.Contains(sub.Name(), "English") {
			englishSubs = append(englishSubs, sub)
		}
	}
	// sort by size desc
	sort.Slice(englishSubs, func(i, j int) bool {
		iInfo, _ := englishSubs[i].Info()
		jInfo, _ := englishSubs[j].Info()
		return iInfo.Size() > jInfo.Size()
	})
	return englishSubs
}

func determineSubtitleType(filename string, sortIndex int, totalFiles int) string {
	if totalFiles == 1 {
		// if only one file, assume non-SDH
		return subtitleTypeMap[1] + filepath.Ext(filename)
	}
	return subtitleTypeMap[sortIndex] + filepath.Ext(filename)
}

func cleanUpNullFiles(root string) {
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(path) == ".srt" {
			data, _ := os.ReadFile(path)
			if string(bytes.Trim(data, "\x00")) == "" {
				os.RemoveAll(path)
				log.Printf("removed: %v", path)
			}
		}
		return nil
	})
}
