package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"os/exec"

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

		converted := false
		if filepath.Ext(file) == ".flac" {
			log.Println("Converting to MP3: ", file)
			mp3File, err := convertToMP3(file)
			if err != nil {
				log.Println("Error converting to MP3:", err)
			} else {
				file = mp3File
				converted = true
			}
		}

		resultFileName, err := movemusic.CopyMusic(file, destination, false)

		if err != nil {
			if errors.Is(err, movemusic.ErrFileExists) {
				log.Println("File exists, skipping ", resultFileName)
			} else {
				log.Println("Error copying file:", err)
				break
			}

		} else {
			log.Println("Copied ", resultFileName)
		}

		if converted {
			log.Println("Removing temporary file: ", file)
			err = os.Remove(file)
			if err != nil {
				log.Println("Error removing temporary file:", err)
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

func convertToMP3(sourceFile string) (string, error) {
	// Get the user's home folder
	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	tempFolder := filepath.Join(home, ".mixitup-conversion")
	err = os.MkdirAll(tempFolder, os.ModePerm)
	if err != nil {
		return "", err
	}

	tempFileName := filepath.Join(tempFolder, strings.TrimSuffix(filepath.Base(sourceFile), filepath.Ext(sourceFile))+".mp3")
	//log.Println("Temp file name: ", tempFileName)
	cmd := exec.Command("ffmpeg", "-i", sourceFile, tempFileName)

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return tempFileName, nil
}
