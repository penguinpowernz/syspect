package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Open the audit.log file for reading
	file, err := os.Open("/var/log/audit/audit.log")
	if err != nil {
		fmt.Println("Error opening audit.log:", err)
		return
	}
	defer file.Close()

	// Create a reader to read the file
	reader := bufio.NewReader(file)

	// Seek to the end of the file
	_, err = file.Seek(0, os.SEEK_END)
	if err != nil {
		fmt.Println("Error seeking to the end of the file:", err)
		return
	}

	// Define a regular expression to match execve events in the audit.log
	execveRegex := regexp.MustCompile(`type=EXECVE.*`)

	// Continuously read new lines from the end of the file
	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			// Ignore EOF errors, wait for more data to be written
			if err.Error() != "EOF" {
				fmt.Println("Error reading audit.log:", err)
			}
		} else {
			// Check if the line contains an execve event
			if execveRegex.MatchString(line) {
				// Extract and print the arguments
				args := extractArguments(line)
				fmt.Printf("Executed Command Arguments: %v\n", args)
			}
		}
	}
}

func extractArguments(line string) []string {
	// Define a regular expression to extract the arguments from the line
	argRegex := regexp.MustCompile(`a\d+="([^"]+)"`)

	// Find and collect the arguments in the line
	matches := argRegex.FindAllStringSubmatch(line, -1)
	args := []string{}
	for _, match := range matches {
		if len(match) == 2 {
			args = append(args, match[1])
		}
	}

	return args
}
