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
	if len(os.Args) != 3 {
		fmt.Println("Usage: mixitup <source> <destination>")
		return
	}

	source := os.Args[1]
	destination := os.Args[2]

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
