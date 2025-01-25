package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseCommitFile(fileName string) {
	var targetFiles []string
	file, err := os.Open(fileName)
	checkErr(err)
	defer file.Close()
	fileScan := bufio.NewScanner(file)

	// Get the first line of the file for the description.
	fileScan.Scan()
	newDescription := fileScan.Text()
	fmt.Println("New Description:", newDescription)

	var splits []string
	// Scan for the target files.
	r := regexp.MustCompile(`# file: `)
	for fileScan.Scan() {
		splits = r.Split(fileScan.Text(), 2)
		if len(splits) > 1 {
			targetFiles = append(targetFiles, splits[1])
		}
	}

	fmt.Println(targetFiles)

	// TODO: Process EXIF change
}

func main() {
	exifLabel := "ImageDescription"

	// Handle arg or input.
	var searchTerm string
	if len(os.Args) > 2 {
		fmt.Println("Notice: Ignoring all but first argument.")
	}
	if len(os.Args) > 1 {
		searchTerm = os.Args[1]
	} else {
		fmt.Println("Search term:")
		fmt.Scanln(&searchTerm)
	}
	fmt.Printf("Searching [%s]\n", searchTerm)

	// Search Paths
	searchPaths := []string{"."}
	env := os.Getenv("EXIFCOMMIT_PATH")
	searchPaths = append(searchPaths, filepath.SplitList(env)...)

	// Find files
	var fileList []string
	for _, searchPath := range searchPaths {
		matches, _ := filepath.Glob(fmt.Sprintf("%s/*%s*", searchPath, searchTerm))
		fileList = append(fileList, matches...)
	}

	// Create temp file and open to write
	tmpFile, err := os.CreateTemp(".", fmt.Sprintf(".EXIF-%s-", searchTerm))
	checkErr(err)
	defer tmpFile.Close()
	var commitMessage string

	// Commit explanation.
	commitMessage = fmt.Sprintf("\n# First line of this file is used for the %s", exifLabel)
	commitMessage += "\n# An empty line aborts the change."
	commitMessage += "\n#\n# Files to be modified, and their current value"
	commitMessage += "\n# (remove to exclude from editing):\n#"

	// Add list of files
	for _, fn := range fileList {
		// TODO: Read current EXIF tag
		commitMessage += fmt.Sprintf("\n# file: %s\n%s", fn, "Place holder description\n")
	}
	tmpFile.WriteString(commitMessage)
	tmpFile.Sync()

	// Run editor with tmp file
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fall back to a logical default
		editor = "vim"
	}
	fmt.Printf("Using editor: %s\n", editor)
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Parse after editing
	parseCommitFile(tmpFile.Name())

	// Remove tmp file
	os.Remove(tmpFile.Name())
}
