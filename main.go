package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"

	"github.com/punkscience/movemusic"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: mixitup <source> <destination>")
		return
	}

	source := os.Args[1]
	destination := os.Args[2]

	log.Println("Looking for music files at: ", source)
	musicFiles, err := findMusicFiles(source)
	if err != nil {
		fmt.Println("Error finding music files:", err)
		return
	}

	log.Println("Shuffling...")
	rand.Shuffle(len(musicFiles), func(i, j int) { musicFiles[i], musicFiles[j] = musicFiles[j], musicFiles[i] })

	log.Println("Copying music files to: ", destination)
	for _, file := range musicFiles {
		if !hasEnoughSpace(destination) {
			log.Println("Not enough space in destination")
			break
		}

		filename, err := movemusic.CopyMusic(file, destination, false)

		if err != nil {
			fmt.Println("Error moving file:", err)
			break
		} else {
			fmt.Println("Created file:", filename)
		}
	}
}

func findMusicFiles(root string) ([]string, error) {
	var musicFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".flac" || filepath.Ext(path) == ".mp3") {
			musicFiles = append(musicFiles, path)
		}
		return nil
	})
	return musicFiles, err
}

func hasEnoughSpace(path string) bool {
	var stat syscall.Statfs_t
	syscall.Statfs(path, &stat)
	// Check if there is at least 100MB free space
	return stat.Bavail*uint64(stat.Bsize) > 100*1024*1024
}
