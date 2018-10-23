package main

import (
	"bufio"
	"fmt"
	"github.com/aryann/difflib"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"sort"
	"strings"
)

var whitelistOpts struct {
	MasterWhitelistPath string `long:"base" description:"Path to approved whitelist"`
	UpdatedWhitelistPath string `long:"updated" required:"true" description:"Path to whitelist with changes (required)"`
	ReportFile string `long:"out" description:"Write Markdown report to this file (default is stdout)"`
}

var parser = flags.NewParser(&whitelistOpts, flags.Default)

func main() {
	log.SetOutput(os.Stderr)

	_, err := parser.Parse()

	if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
		return
	}

	if err != nil {
		return
	}

	output := whitelistCheck(whitelistOpts.MasterWhitelistPath, whitelistOpts.UpdatedWhitelistPath, whitelistOpts.ReportFile)
	if whitelistOpts.ReportFile != "" {
		var report []string
		report = append(report, "Report line 1")
		report = append(report, "Report line 2")
		report = append(report, "Report line 3")
		writeLines(report, whitelistOpts.ReportFile)
	}

	haveOutput := false
	for _, line := range output {
		log.Print(line)
		haveOutput = true
	}
	if haveOutput {
		os.Exit(1)
	}
}

func whitelistCheck(masterPath string, updatedPath string, logFile string) []string {
	var output []string

	var masterLines []string
	var err error
	if masterPath != "" {
		masterLines, err = readLines(masterPath)
		if err != nil {
			log.Fatalf("Error reading master whitelist file: %s\n", err)
		}
	}
	updatedLines, err := readLines(updatedPath)
	if err != nil {
		log.Fatalf("Error reading updated whitelist file: %s\n", err)
	}

	var uniqueLines []string
	var orderedUpdatedLines []string
	diff := difflib.Diff(masterLines, updatedLines)
	for _, diffEntry := range diff {
		if diffEntry.Delta == difflib.RightOnly {
			//log.Printf("Found only in updated: %s", diffEntry.Payload)
			uniqueLines = append(uniqueLines, diffEntry.Payload)
			orderedUpdatedLines = append(orderedUpdatedLines, diffEntry.Payload)
		}
	}

	for lineNum, line := range uniqueLines {
		if strings.HasPrefix(strings.ToLower(line), "bad") {
			output = append(output, fmt.Sprintf("Bad data at line %d: '%s'", lineNum+ 1,line))
		}
	}

	sort.Slice(orderedUpdatedLines, caseInsensitiveSort(orderedUpdatedLines))

	for lineNum := range uniqueLines {
		if uniqueLines[lineNum] != orderedUpdatedLines[lineNum] {
			output = append(output, fmt.Sprintf("Unordered line %d: Expected '%s' to be next", lineNum+1, orderedUpdatedLines[lineNum]))
		}
	}

	return output
}




func caseInsensitiveSort(lines []string) (func(i, j int) bool) {
	return func(i, j int) bool {
		return strings.ToLower(lines[i]) < strings.ToLower(lines[j])
	}
}


// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}