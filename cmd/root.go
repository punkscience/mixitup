package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ricochet2200/go-disk-usage/du"
	"github.com/spf13/cobra"

	"errors"
	"log"
	"math/rand"
	"path/filepath"

	"github.com/punkscience/movemusic"
	"github.com/schollz/progressbar/v3"
)

var rootCmd = &cobra.Command{
	Use:   "mixitup",
	Short: "Fill up a drive with randomized music files.",
	Long:  `A command line application which reads all the music files in a source folder, randomizes them, and copies them to a destination folder.`,
	// Optional: Run function for root command if it does something by default
	Run: func(cmd *cobra.Command, args []string) {
		source := strings.Trim(cmd.Flag("source").Value.String(), " ")
		target := strings.Trim(cmd.Flag("target").Value.String(), " ")
		filter := strings.Trim(cmd.Flag("filters").Value.String(), " ")

		if source == "" || target == "" {
			fmt.Println("Usage: mixitup --source <source> --target <destination> [--filter <filter>]")
			return
		}

		// Parse the filter into a list of strings
		filterList := strings.Split(filter, ";")

		// Print the filter list
		fmt.Println("Filter list:", filterList)

		log.Println("Looking for music files at: ", source)

		pbar := progressbar.Default(-1, "Finding music files...")
		musicFiles, err := findMusicFiles(source, pbar, filterList)
		if err != nil {
			fmt.Println("Error finding music files:", err)
			return
		}
		pbar.Finish()

		rand.Shuffle(len(musicFiles), func(i, j int) { musicFiles[i], musicFiles[j] = musicFiles[j], musicFiles[i] })

		log.Println("Copying music files to: ", target)
		for _, file := range musicFiles {
			// Get the file size
			fileInfo, err := os.Stat(file)
			if err != nil {
				log.Println("Error getting file info:", err)
				break
			}

			// Check if there is enough space in the destination
			if !hasEnoughSpace(target, fileInfo.Size()) {
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

			resultFileName, err := movemusic.CopyMusic(file, target, false)

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
	},
}

// Execute is the main entry point for the CLI application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add any global flags or configuration here
	rootCmd.PersistentFlags().String("source", "", "Music folder or source path.")
	rootCmd.PersistentFlags().String("target", "", "Target drive / folder.")
	rootCmd.PersistentFlags().String("filters", "", "Semicolon separated list of file keywords to exclude.")
}

func findMusicFiles(root string, pbar *progressbar.ProgressBar, filterList []string) ([]string, error) {
	var musicFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() &&
			(filepath.Ext(path) == ".flac" || filepath.Ext(path) == ".mp3") &&
			!containsAny(path, filterList) {
			musicFiles = append(musicFiles, path)
			pbar.Add(1)
		}
		return nil
	})
	return musicFiles, err
}

func containsAny(str string, list []string) bool {
	for _, item := range list {
		if strings.Contains(strings.ToLower(str), strings.ToLower(item)) {
			return true
		}
	}
	return false
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
