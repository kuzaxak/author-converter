package cmd

import (
	"bufio"
	"fmt"
	"github.com/kuzaxak/author-converter/client"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

func parseCurlCommand(filePath string) (map[string]string, map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	headers := make(map[string]string)
	cookies := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = line[0 : len(line)-2] // Trim last ' \

		if strings.HasPrefix(line, "-H") {
			headerParts := strings.SplitN(line, ":", 2)
			if len(headerParts) < 2 {
				continue
			}

			headerName := strings.TrimSpace(headerParts[0][4:len(headerParts[0])])
			headerValue := strings.TrimSpace(strings.Trim(headerParts[1], `'`))

			// special handling for cookie header
			if strings.ToLower(headerName) == "cookie" {
				cookieParts := strings.Split(headerValue, ";")
				for _, cookiePart := range cookieParts {
					cookieNameValue := strings.SplitN(cookiePart, "=", 2)
					if len(cookieNameValue) < 2 {
						continue
					}

					cookieName := strings.TrimSpace(cookieNameValue[0])
					cookieValue := strings.TrimSpace(cookieNameValue[1])

					cookies[cookieName] = cookieValue
				}
			} else {
				headers[headerName] = headerValue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return headers, cookies, nil
}

var downloadBookCmd = &cobra.Command{
	Use:   "download [curl_file_path] [book_id]",
	Short: "Download a book",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Echo: " + strings.Join(args, " "))

		curlFilePath := args[0]
		bookId := args[1]

		headers, cookies, err := parseCurlCommand(curlFilePath)
		if err != nil {
			log.Fatal("Error parsing curl command: ", err)
		}

		client := client.NewClient("https://author.today/", headers, cookies)
		e := client.GetChapters(bookId)
		err = e.Write(fmt.Sprintf("%s - %s.epub", e.Author(), e.Title()))
		if err != nil {
			log.Fatal("Error writing to file")
		}
	},
}
