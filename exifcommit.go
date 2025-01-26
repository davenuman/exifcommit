package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
)

const exifLabel = "ImageDescription"

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Read the current value of the tag in the targetFile.
func readTag(targetFile string) string {
	// fmt.Printf("attempting EXIF on file <%s> tag [%s]\n", targetFile, tagValue)

	// jmp := NewJpegMediaParser()
	rawExif, err := exif.SearchFileAndExtractExif(targetFile)
	checkErr(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	checkErr(err)

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	checkErr(err)

	tagName := exifLabel

	rootIfd := index.RootIfd

	// We know the tag we want is on IFD0 (the first/root IFD).
	results, err := rootIfd.FindTagWithName(tagName)
	checkErr(err)

	// This should never happen.
	if len(results) != 1 {
		log.Panicf("there wasn't exactly one result")
	}
	ite := results[0]

	valueRaw, err := ite.Value()
	checkErr(err)

	value := valueRaw.(string)
	fmt.Println(value)
	return value
}

// We do this not because it is easy...
func writeTag(targetFile string, newDescription string) bool {
	rawExif, err := exif.SearchFileAndExtractExif(targetFile)
	checkErr(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	checkErr(err)

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	checkErr(err)

	rootIfd := index.RootIfd

	// We know the tag we want is on IFD0 (the first/root IFD).
	results, err := rootIfd.FindTagWithName(exifLabel)
	checkErr(err)

	// This should never happen.
	if len(results) != 1 {
		log.Panicf("there wasn't exactly one result")
	}
	ite := results[0]

	valueRaw, err := ite.Value()
	checkErr(err)

	value := valueRaw.(string)
	fmt.Println(value)

	// jpegstructure method

	fmt.Printf("attempting EXIF on file <%s> tag [%s]\n", targetFile, newDescription)

	jmp := jis.NewJpegMediaParser()

	intfc, err := jmp.ParseFile(targetFile)
	checkErr(err)

	sl := intfc.(*jis.SegmentList)
	// sl.Print()

	// Update the tag.

	rootIb, err := sl.ConstructExifBuilder()
	checkErr(err)
	fmt.Println("does not get this far.")

	ifdPath := "IFD0"

	ifdIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPath)
	checkErr(err)

	fmt.Println(ifdIb)
	err = ifdIb.SetStandardWithName(exifLabel, newDescription)
	checkErr(err)

	// Update the exif segment.

	err = sl.SetExif(rootIb)
	checkErr(err)

	b := new(bytes.Buffer)

	err = sl.Write(b)
	checkErr(err)

	return true
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
	r := regexp.MustCompile(`file: `)
	for fileScan.Scan() {
		splits = r.Split(fileScan.Text(), 2)
		if len(splits) > 1 {
			targetFiles = append(targetFiles, splits[1])
		}
	}

	for _, tf := range targetFiles {
		writeTag(tf, newDescription)
	}

	// TODO: Process EXIF change
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
	var description string
	for _, fn := range fileList {
		// TODO: Read current EXIF tag
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

	// Parse after editing
	parseCommitFile(tmpFile.Name())

	// Remove tmp file
	os.Remove(tmpFile.Name())
}
