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
		for _, path := range pathsToDelete {
			err := os.RemoveAll(path)
			if err != nil {
				log.Printf("failed to delete %v: %v", path, err)
			}
		}
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
			subtitles, _ := os.ReadDir(path)
			sort.Sort(BySizeDesc(subtitles))
			for i, sub := range subtitles {
				oldPath := filepath.Join(path, sub.Name())
				newPath := filepath.Join(filepath.Dir(filepath.Dir(path)), // two directories up
					d.Name()+subtitleTypeMap[i]+filepath.Ext(sub.Name()))
				info, _ := sub.Info()
				log.Printf("renaming (size %v) %v \n\tto %v", info.Size(), oldPath, newPath)
				err = os.Rename(oldPath, newPath)
				if err != nil {
					log.Fatal(err)
				}
			}
			log.Printf("finished %v \n\n", path)
		}
		return nil
	})
	return err
}

type BySizeDesc []os.DirEntry

func (b BySizeDesc) Len() int { return len(b) }

func (b BySizeDesc) Less(i, j int) bool {
	iInfo, _ := b[i].Info()
	jInfo, _ := b[j].Info()
	return iInfo.Size() > jInfo.Size()
}

func (b BySizeDesc) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
