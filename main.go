package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type HTTPResponse struct {
	URL    string          `json:"url"`
	Title  string          `json:"title"`
	Status json.RawMessage `json:"status_code"`
}

func main() {
	// Define command-line flags
	sc2Flag := flag.Bool("sc2", false, "Show URLs with yesterday's non-200 status and today's 200 status")
	silentFlag := flag.Bool("silent", false, "Show URLs with changes in title or status code")

	//file1 := "httpx.resolved.all.yesterday.json"
	//file2 := "httpx.resolved.all.today.json"

	var file1, file2 string
	flag.StringVar(&file1, "fy", "", "Path to yesterday's file")
	flag.StringVar(&file2, "ft", "", "Path to today's file")

	// Parse command-line flags
	flag.Parse()

	// Check if both file paths are provided
	if file1 == "" || file2 == "" {
		fmt.Println("Both -fy and -ft flags must be provided.")
		return
	}
	// Parse command-line flags
	//flag.Parse()

	// Open the files for reading
	f1, err := os.Open(file1)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file 1:", err)
		return
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file 2:", err)
		return
	}
	defer f2.Close()

	// Create a map to store the responses from file1
	responses1 := make(map[string]HTTPResponse)

	// Parse file 1
	scanner := bufio.NewScanner(f1)
	for scanner.Scan() {
		var response HTTPResponse
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing file 1:", err)
			continue // Continue processing the rest of the data
		}
		// Convert the URL to lowercase
		response.URL = strings.ToLower(response.URL)
		responses1[response.URL] = response
	}

	// Create a slice to store change messages
	var changeMessages []string

	// Parse file 2 and compare
	scanner = bufio.NewScanner(f2)
	for scanner.Scan() {
		var response HTTPResponse
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing file 2:", err)
			continue // Continue processing the rest of the data
		}

		// Convert the URL to lowercase for comparison
		response.URL = strings.ToLower(response.URL)

		// Check if the status code is not a server error code (500, 501, 502, 503, 504)
		statusStr := string(response.Status)
		isServerError := strings.HasPrefix(statusStr, "50")

		if response1, ok := responses1[response.URL]; ok {
			if (response1.Title != response.Title || string(response1.Status) != string(response.Status)) && !isServerError {

				if *sc2Flag {
					// Show URLs with yesterday's non-200 status and today's 200 status
					status1Str := string(response1.Status)
					is200Today := strings.HasPrefix(statusStr, "2") && !strings.HasPrefix(status1Str, "2")
					if is200Today {
						message := fmt.Sprintf("URL: %s\n   Title (Yesterday): %s\n   Title (Today): %s\n   Status Code (Yesterday): %s\n   Status Code (Today): %s\n",
							response.URL, response1.Title, response.Title, response1.Status, response.Status)
						changeMessages = append(changeMessages, message)
					}
				} else if *silentFlag {
					// Show URLs with changes in title or status code
					message := fmt.Sprintf("%s",
						response.URL)
					changeMessages = append(changeMessages, message)

				} else {
					// Show all changes
					message := fmt.Sprintf("URL: %s\n   Title (Yesterday): %s\n   Title (Today): %s\n   Status Code (Yesterday): %s\n   Status Code (Today): %s\n",
						response.URL, response1.Title, response.Title, response1.Status, response.Status)
					changeMessages = append(changeMessages, message)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file 2:", err)
	}

	// Print change messages
	for _, message := range changeMessages {
		fmt.Println(message)
	}
}
