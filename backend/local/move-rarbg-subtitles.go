package main

import (
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

func main() {
	root := "G:\\TV Shows"
	var pathsToDelete []string

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
	// e.g. G:\TV Shows\Better Call Saul\Better.Call.Saul.S05.1080p.BluRay.x265-RARBG\Subs
	err := filepath.WalkDir(subsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() != "Subs" {
			log.Printf("handling epi subs dir: %s", path)
			subtitles := getSortedEnglishSubs(path)
			for i, sub := range subtitles {
				oldPath := filepath.Join(path, sub.Name())
				newPath := filepath.Join(
					strings.Split(path, string(os.PathSeparator)+"Subs"+string(os.PathSeparator))[0],
					d.Name()+determineSubtitleType(sub.Name(), i, len(subtitles)),
				)
				info, _ := sub.Info()
				log.Printf("copying (size %v) %v \n\tto %v", info.Size(), oldPath, newPath)
				copyFile(oldPath, newPath)
			}
			log.Printf("finished %v \n\n", path)
		}
		return nil
	})
	return err
}

func copyFile(read string, write string) {
	r, err := os.Open(read)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	w, err := os.Create(write)
	if err != nil {
		panic(err)
	}
	defer w.Close()
	if _, err = w.ReadFrom(r); err != nil {
		log.Fatal(err)
	}
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
	log.Printf("sorted: %v", englishSubs)
	return englishSubs
}

func determineSubtitleType(filename string, sortIndex int, totalFiles int) string {
	if totalFiles == 1 {
		// if only one file, assume non-SDH
		return subtitleTypeMap[1] + filepath.Ext(filename)
	}
	return subtitleTypeMap[sortIndex] + filepath.Ext(filename)
}
