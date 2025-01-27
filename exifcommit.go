package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	exiftool "github.com/barasher/go-exiftool"
)

const exifLabel = "ImageDescription"

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Just print the error and continue
func ignoreErr(err error) {
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}
}

// Read the current value of the tag in the targetFile.
func readTag(targetFile string) string {
	et, err := exiftool.NewExiftool()
	if err != nil {
		fmt.Printf("Error when intializing: %v\n", err)
		return "not found"
	}
	defer et.Close()

	metadata := et.ExtractMetadata(targetFile)
	value, err := metadata[0].GetString(exifLabel)
	if err != nil {
		// Not really an error.
		return "not found"
	}
	err = et.Close()
	ignoreErr(err)

	// fmt.Printf("found: %v\n", value)
	return value
}

// This function doesn't work. It results in the following error:
// error while closing exiftool: [error while waiting for exiftool to exit: exit status 1]
// So we are using writeExif instead.
func writeTag(targetFile string, newDescription string) bool {
	fmt.Printf("attempting EXIF on file <%s> tag [%s]\n", targetFile, newDescription)

	// et, err := exiftool.NewExiftool(exiftool.BackupOriginal())
	et, err := exiftool.NewExiftool(exiftool.ClearFieldsBeforeWriting())
	ignoreErr(err)
	defer et.Close()

	metadata := et.ExtractMetadata(targetFile)

	metadata[0].SetString(exifLabel, newDescription)
	// metadata[0].SetString("ISO", "190")
	et.WriteMetadata(metadata)
	err = et.Close()
	ignoreErr(err)

	altered := et.ExtractMetadata(targetFile)
	testTag, err := altered[0].GetString(exifLabel)
	ignoreErr(err)
	fmt.Println("altered:" + testTag)

	return true
}

func writeExif(targetFiles []string, newDescription string) bool {
	etArgs := append([]string{fmt.Sprintf(`-%s=%s`, exifLabel, newDescription)}, targetFiles...)

	cmd := exec.Command("exiftool", etArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// exiftool -delete_original! ${targetFiles}
	delArgs := append([]string{"-delete_original!"}, targetFiles...)
	deleteCmd := exec.Command("exiftool", delArgs...)
	deleteCmd.Stdin = os.Stdin
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	deleteCmd.Run()

	return true
}

func parseCommitFile(fileName string) bool {
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
	r := regexp.MustCompile(`file: `)
	for fileScan.Scan() {
		splits = r.Split(fileScan.Text(), 2)
		if len(splits) > 1 {
			targetFiles = append(targetFiles, splits[1])
		}
	}

	if len(targetFiles) < 1 {
		fmt.Println("No target files. Aborting.")
		os.Exit(1)
	}
	if len(newDescription) < 1 {
		fmt.Println("No new description. Aborting.")
		os.Exit(1)
	}

	writeExif(targetFiles, newDescription)
	// for _, tf := range targetFiles {
	// 	writeTag(tf, newDescription)
	// }

	return true
}

func main() {
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

	if len(fileList) < 1 {
		fmt.Println("No files found. Aborting.")
		os.Exit(1)
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
	var description string
	for _, fn := range fileList {
		// Read current EXIF tag
		description = readTag(fn)
		commitMessage += fmt.Sprintf("\n# file: %s\n%s", fn, description)
	}
	tmpFile.WriteString(commitMessage)
	tmpFile.Sync()

	// Run editor with tmp file
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fall back to a logical default
		editor = "vim"
	}
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// Remove tmp file
	defer os.Remove(tmpFile.Name())

	// Parse after editing
	parseCommitFile(tmpFile.Name())
}
