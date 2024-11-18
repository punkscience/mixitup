package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/punkscience/movemusic"
	"github.com/ricochet2200/go-disk-usage/du"
	"github.com/schollz/progressbar/v3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: mixitup <source> <destination>")
		return
	}

	source := os.Args[1]
	destination := os.Args[2]

	log.Println("Looking for music files at: ", source)

	pbar := progressbar.Default(-1, "Finding music files...")
	musicFiles, err := findMusicFiles(source, pbar)
	if err != nil {
		fmt.Println("Error finding music files:", err)
		return
	}
	pbar.Finish()

	rand.Shuffle(len(musicFiles), func(i, j int) { musicFiles[i], musicFiles[j] = musicFiles[j], musicFiles[i] })

	log.Println("Copying music files to: ", destination)
	for _, file := range musicFiles {
		// Get the file size
		fileInfo, err := os.Stat(file)
		if err != nil {
			log.Println("Error getting file info:", err)
			break
		}

		// Check if there is enough space in the destination
		if !hasEnoughSpace(destination, fileInfo.Size()) {
			log.Println("Not enough space in destination")
			break
		}

		log.Println("Copying file: ", file)
		filename, err := movemusic.CopyMusic(file, destination, false)

		if err != nil {
			if errors.Is(err, movemusic.ErrFileExists) {
				log.Println("File exists, skipping ", filename)
			} else {
				log.Println("Error copying file:", err)
				break
			}

		}
	}
}

func findMusicFiles(root string, pbar *progressbar.ProgressBar) ([]string, error) {
	var musicFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".flac" || filepath.Ext(path) == ".mp3") {
			musicFiles = append(musicFiles, path)
			pbar.Add(1)
		}
		return nil
	})
	return musicFiles, err
}

func hasEnoughSpace(path string, size int64) bool {
	usage := du.NewDiskUsage(path)

	return usage.Available() >= uint64(size)
}
