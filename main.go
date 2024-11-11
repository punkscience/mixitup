package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func main() {
	var source = ""
	var destination = ""

	// Look for a confif file in the user's home configs folder
	// If it exists, read the source and destination from the config file
	homeFolder, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return
	}
	configFile := filepath.Join(homeFolder, ".config", "mixitup", "config")
	if _, err := os.Stat(configFile); err == nil {
		// Config file exists
		file, err := os.Open(configFile)
		if err != nil {
			fmt.Println("Error opening config file:", err)
			return
		}
		defer file.Close()

		var source, destination string
		_, err = fmt.Fscanf(file, "source=%s\n", &source)
		if err != nil {
			fmt.Println("Error reading source from config file:", err)
			return
		}
		_, err = fmt.Fscanf(file, "destination=%s\n", &destination)

		log.Println("Source:", source)
		log.Println("Destination:", destination)
	}

	if source == "" && destination == "" {

		if len(os.Args) != 3 {
			source = os.Args[1]
			destination = os.Args[2]
		} else {
			fmt.Println("Usage: mixitup <source> <destination>")
			return
		}
	}

	log.Println("Scanning source directory for music files...")
	musicFiles, err := findMusicFiles(source)
	if err != nil {
		fmt.Println("Error finding music files:", err)
		return
	}

	if len(musicFiles) == 0 {
		fmt.Println("No music files found in source directory. Stopping.")
		return
	}

	log.Println("Shuffling...")
	rand.Shuffle(len(musicFiles), func(i, j int) { musicFiles[i], musicFiles[j] = musicFiles[j], musicFiles[i] })

	log.Println("Copying music files to ", destination)
	for _, file := range musicFiles {
		if !copyFile(file, destination) {
			break
		}
	}
}

func findMusicFiles(root string) ([]string, error) {
	var musicFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.ToLower(filepath.Ext(path)) == ".flac" ||
			strings.ToLower(filepath.Ext(path)) == ".mp3" ||
			strings.ToLower(filepath.Ext(path)) == ".wav") {
			musicFiles = append(musicFiles, path)
		}
		return nil
	})
	return musicFiles, err
}

func copyFile(src, dstDir string) bool {
	// Check the file size of the source file
	fileInfo, err := os.Stat(src)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return false
	}
	fileSize := int(fileInfo.Size())

	// If the destination does not have enough space, return false
	if !hasEnoughSpace(dstDir, fileSize) {
		fmt.Println("No more space on destination. Stopping.")
		return false
	}

	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println("Error opening source file:", err)
		return false
	}
	defer srcFile.Close()

	dstPath := filepath.Join(dstDir, filepath.Base(src))
	dstFile, err := os.Create(dstPath)
	if err != nil {
		fmt.Println("Error creating destination file:", err)
		return false
	}
	defer dstFile.Close()

	// Print out what we are copying
	fmt.Println("Copying", src, "to", dstPath)

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return false
	}

	return true
}

func hasEnoughSpace(path string, fileSize int) bool {
	var freeBytesAvailable uint64
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		fmt.Println("Error getting filesystem stats:", err)
		return false
	}
	freeBytesAvailable = stat.Bavail * uint64(stat.Bsize)
	if err != nil {
		fmt.Println("Error getting filesystem stats:", err)
		return false
	}
	// Check if there is at least 100MB free space
	return freeBytesAvailable > uint64(fileSize)
}
