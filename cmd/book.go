package cmd

import (
	"fmt"
	"github.com/kuzaxak/author-converter/client"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var downloadBookCmd = &cobra.Command{
	Use:   "download [book_id]",
	Short: "Download a book",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Echo: " + strings.Join(args, " "))

		client := client.NewClient("https://author.today/")
		e := client.GetChapters(args[0])
		err := e.Write(fmt.Sprintf("%s - %s.epub", e.Author(), e.Title()))
		if err != nil {
			log.Fatal("Error writing to file")
		}
	},
}
