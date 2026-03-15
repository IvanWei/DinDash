package main

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func pathExist (path string) (bool, os.FileInfo) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, info
}

func copyFile (srcPath string, destPath string) error {
	source, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func main() {
	var skipCount, copyCount, errCount int
	// Set default Logger (output with timestamp and level)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	srcFlagArgs := flag.String("src", "", "Source MP3 folder path")
	destFlagArgs := flag.String("dest", "", "Destination path (USB/Disk)")

	flag.Parse()

	srcFolderPath := *srcFlagArgs
	destFolderPath := *destFlagArgs

	if destFolderPath == "" {
		// Automatically print usage instructions
		flag.PrintDefaults()
		return
	}

	// --- Step 1: Check if Destination Exists ---
	isExist, info := pathExist(destFolderPath)

	if !isExist {
		slog.Error("USB device not found. Please ensure it is plugged in and correctly mounted.", "FOLDER", destFolderPath)
		os.Exit(1)
	}
	if info != nil && !info.IsDir() {
		slog.Error("Path exists but is not a valid directory/mount point.", "FOLDER", destFolderPath)
		os.Exit(1)
	}

	// --- Step 2: (Advanced) Test Write Permissions ---
	testFile := filepath.Join(destFolderPath, ".write_test")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		slog.Error("Cannot write to USB! Check if it is read-only or full.", "ERR", err)
		os.Exit(1)
	}
	os.Remove(testFile)

	// Regex: Match all numbers
	re := regexp.MustCompile(`\d+`)

	err = filepath.WalkDir(srcFolderPath, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// --- Skip hidden files (starting with ".") ---
		if strings.HasPrefix(dir.Name(), ".") {
			return nil
		}

		// Only process mp3 files
		if !dir.IsDir() && strings.ToLower(filepath.Ext(dir.Name())) == ".mp3" {
			// 1. Get relative path to avoid capturing numbers from parent directories
			relPath, _ := filepath.Rel(srcFolderPath, path)

			// 2. Remove extension so the regex doesn't catch the "3" in ".mp3"
			purePath := strings.TrimSuffix(relPath, filepath.Ext(relPath))

			// 3. Extract all numbers
			numbers := re.FindAllString(purePath, -1)

			// 4. Compose new filename (e.g., 09-01.mp3)
			var newFileName string
			if len(numbers) > 0 {
				newFileName = strings.Join(numbers, "-")

			} else {
				newFileName = "unknown-" + dir.Name()[:2]
			}

			// --- Check if file already exists in destination ---
			destPath := filepath.Join(destFolderPath, newFileName + ".mp3")
			isExist, _ := pathExist(destPath)

			if isExist {
				skipCount++
				return nil
			}

			// 5. Execute Copy
			slog.Info("Copying...", "to", newFileName+".mp3")
			err = copyFile(path, destPath)
			if err != nil {
			    slog.Error("Copy failed", "file", path, "ERR", err)
			    errCount++
			    return nil
			}
			slog.Info("Copy successful", "to", newFileName+".mp3")

			cmd := exec.Command("kid3-cli", "-c", "set title '" + newFileName + "'", "-c", "set artist ''", "-c", "set album ''", destPath)

			if cmdErr := cmd.Run(); cmdErr != nil {
				slog.Error("ID3 Tag cleanup failed", "file", newFileName, "Err", cmdErr)
			} else {
				slog.Info("ID3 Tag cleanup completed", "file", newFileName)
			}
			copyCount++
		}

		return nil
	})

	if err != nil {
		slog.Error("Scanning process error", "ERR", err)
		os.Exit(1)
	}

	slog.Info("Running final disk cleanup...")
	exec.Command("dot_clean", "-m", destFolderPath).Run()

	slog.Info(strings.Repeat("-", 50))
	slog.Info("📊 Execution Statistics",
		"Total", skipCount + copyCount + errCount,
		"Added", copyCount,
		"Skipped", skipCount,
		"Failed", errCount,
	)
}
