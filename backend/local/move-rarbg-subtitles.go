package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
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
	start := time.Now()
	var pathsToDelete []string
	var mu sync.Mutex
	for _, root := range []string{tvRoot, moviesRoot} {
		go cleanUpFiles(root)
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && d.Name() == "Subs" {
				//go func() {
				err = handleSubsDir(path)
				if err == nil {
					mu.Lock()
					pathsToDelete = append(pathsToDelete, path)
					mu.Unlock()
				}
				//}()
			}
			return err
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	//log.Printf("finna delete: \n%v", strings.Join(pathsToDelete, "\n"))
	//for _, path := range pathsToDelete {
	//	err := os.RemoveAll(path)
	//	if err != nil {
	//		log.Printf("failed to delete %v: %v", path, err)
	//	}
	//}

	log.Printf("finished after %v", time.Since(start))
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
	return copySubs(subsRoot, targetVideoFile)
}

func handleTvSubsDir(subsRoot string) error {
	// e.g. G:\TV Shows\Better Call Saul\Better.Call.Saul.S05.1080p.BluRay.x265-RARBG\Subs
	return filepath.WalkDir(subsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() != "Subs" {
			return copySubs(path, d.Name())
		}
		return nil
	})
}

func getTargetDir(path string) string {
	return strings.Split(path+string(os.PathSeparator), string(os.PathSeparator)+"Subs"+string(os.PathSeparator))[0]
}

func copySubs(subsPath, targetVideoFile string) error {
	if targetVideoFile == "" {
		log.Printf("no target video file for path: %v", subsPath)
		return nil
	}

	// log.Printf("handling subs dir: %s", subsPath)
	subtitles, err := getSortedEnglishSubs(subsPath)
	if err != nil {
		return err
	}

	for i, sub := range subtitles {
		srcPath := filepath.Join(subsPath, sub.Name())
		destPath := filepath.Join(getTargetDir(subsPath),
			targetVideoFile+determineSubtitleType(sub.Name(), i, subtitles),
		)

		if areFilesEqual(srcPath, destPath) {
			//log.Printf("files are equal, not copying: %v", srcPath)
		} else {
			info, err := sub.Info()
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to get info on %v", sub))
			}
			log.Printf("copying (size %v) %v \n    to %v \n\n", info.Size(), srcPath, destPath)
			_, err = copyFile(srcPath, destPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to copy %v to %v", srcPath, destPath))
			}
		}
	}
	// log.Printf("finished %v \n\n", subsPath)
	return nil
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
	srcBytes, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("failed to read %v: %v", src, err)
	}
	destBytes, _ := os.ReadFile(dest)
	return bytes.Equal(srcBytes, destBytes)
}

func getSortedEnglishSubs(path string) ([]os.DirEntry, error) {
	var englishSubs []os.DirEntry
	subtitles, err := os.ReadDir(path)
	if err != nil {
		return englishSubs, err
	}

	for _, sub := range subtitles {
		if strings.Contains(strings.ToLower(sub.Name()), "eng") {
			englishSubs = append(englishSubs, sub)
		}
	}
	// sort by size desc
	sort.Slice(englishSubs, func(i, j int) bool {
		iInfo, err := englishSubs[i].Info()
		if err != nil {
			log.Fatalf("failed to get info on %v: %v", englishSubs[i], err)
		}
		jInfo, err := englishSubs[j].Info()
		if err != nil {
			log.Fatalf("failed to get info on %v: %v", englishSubs[j], err)
		}
		return iInfo.Size() > jInfo.Size()
	})
	return englishSubs, nil
}

func determineSubtitleType(filename string, sortIndex int, files []os.DirEntry) string {
	if len(files) == 1 {
		// if only one file, assume non-SDH
		return subtitleTypeMap[1] + filepath.Ext(filename)
	}
	if len(files) == 2 {
		bigFile, err := files[0].Info()
		if err != nil {
			log.Fatalf("failed to get info on %v: %v", files[0], err)
		}
		smallFile, err := files[1].Info()
		if err != nil {
			log.Fatalf("failed to get info on %v: %v", files[1], err)
		}

		// small file is significantly smaller, assume we have .en and .en.forced
		if float64(smallFile.Size())/float64(bigFile.Size()) < .3 {
			return subtitleTypeMap[sortIndex+1] + filepath.Ext(filename)
		}
	}

	return subtitleTypeMap[sortIndex] + filepath.Ext(filename)
}

// del "\\?\G:\TV Shows\Marvels Agents of S.H.I.E.L.D.\"
func cleanUpFiles(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if strings.ToUpper(filepath.Ext(path)) == ".SRT" {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if string(bytes.Trim(data, "\x00")) == "" {
					err := os.RemoveAll(path)
					if err != nil {
						return err
					}
					log.Printf("removed: %v", path)
				}
			}

			if strings.ToUpper(d.Name()) == "RARBG_DO_NOT_MIRROR.EXE" {
				//|| (strings.HasPrefix(d.Name(), ".") && strings.ToUpper(filepath.Ext(path)) == ".PARTS")
				err := os.RemoveAll(path)
				if err != nil {
					return err
				}
				log.Printf("removed: %v", path)
			}
		}
		return nil
	})
}
