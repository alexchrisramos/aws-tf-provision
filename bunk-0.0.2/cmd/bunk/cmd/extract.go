/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract a compressed bundle",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("extract called")
		extractBundle(args)
	},
}

func extractBundle(filename []string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Could not locate user's home dir: %s\n", err)
	}
	ticketsDir := os.Getenv("BUNK_TICKETS_DIR")
	if ticketsDir == "" {
		ticketsDir = homeDir + "/Documents/logs/tickets"
	}
	// workDir, err := os.Getwd()
	// if err != nil {
	// 	log.Fatalf("Could not get work dir: %s\n", err)
	// }
	bundleFilename := strings.Join(filename, " ")

	// fmt.Println(bundleFilename)
	// fmt.Println(homeDir)
	// fmt.Println(workDir)
	// fmt.Println(ticketsDir)

	reader := bufio.NewReader(os.Stdin)
	var ticket string
	for {
		fmt.Print("Enter ticket: -> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Could not read string: %s\n", err)
		}
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if text == "" {
			log.Fatalf("Please specify a ticket to continue")
		}

		ticket = text
		break
	}

	bundleFilePath, bundleFilename := filepath.Split(bundleFilename)

	f, err := os.Open(bundleFilePath + bundleFilename)
	// f, err := os.OpenFile(bundleFilePath+bundleFilename, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatalf("Could not open file %v: %v\n", bundleFilePath+bundleFilename, err)
	}
	defer f.Close()

	contentType, err := GetFileContentType(f)
	if err != nil {
		log.Fatalf("Could not get content type for file %v: %v\n", bundleFilename, err)
	}

	if contentType != "application/x-gzip" {
		// fmt.Printf("File content type: %s\n", contentType)
		// fmt.Println("Extracting bundle!")
		log.Fatalf("File content type is %v; expected application/x-gzip\n", contentType)
	}

	var bundleDir string
	d := strings.Split(bundleFilename, ".")
	if len(d) > 0 {
		bundleDir = ticketsDir + "/" + ticket + "/bundle-" + d[0]
	}

	// Check if bundle dir exists
	// TODO: Revisit this... os.IsExist()
	if _, err := os.Stat(bundleDir); err == nil {
		log.Fatalf("Failed to create bundle dir %v: able to stat dir", bundleDir)
	}

	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		log.Fatalf("Failed to create directory %v: %v\n", bundleDir, err)
	}

	fmt.Printf("Extracting %v to %v\n", bundleFilePath+bundleFilename, bundleDir)

	if err := archiver.Unarchive(bundleFilePath+bundleFilename, bundleDir); err != nil {
		log.Fatalf("Failed to unarchive %v: %v\n", f, err)
	}

	var files []string
	if err := filepath.Walk(bundleDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				log.Fatalf("Could not open file %v: %v\n", bundleFilePath+bundleFilename, err)
			}
			defer f.Close()

			fct, err := GetFileContentType(f)
			if err != nil {
				log.Fatalf("Could not get %v content type: %v\n", path, err)
			}

			if fct == "application/x-gzip" {
				files = append(files, path)
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}

	for _, file := range files {
		s := strings.Split(strings.TrimSuffix(strings.TrimSuffix(file, ".gz"), ".tar"), "/")
		dirName := s[len(s)-1]

		gzFile, err := os.Open(file)
		if err != nil {
			log.Fatalf("Failed to open file %v: %v\n", file, err)
		}
		defer gzFile.Close()

		tarFile, err := os.OpenFile(strings.TrimSuffix(file, ".gz"), os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			log.Fatalf("failed to open %v: %v\n", strings.TrimSuffix(file, ".gz"), err)
		}
		defer tarFile.Close()

		err = archiver.NewGz().Decompress(gzFile, tarFile)
		if err != nil {
			log.Fatalf("failed to decommpress archive: %v\n", err)
		}

		tar := archiver.Tar{
			OverwriteExisting: true,
			MkdirAll:          true,
		}

		if err := tar.Unarchive(strings.TrimSuffix(file, ".gz"), bundleDir+"/"+dirName); err != nil {
			log.Fatalf("Failed to unarchive %v: %v\n", strings.TrimSuffix(file, ".gz"), err)
		}

		if err := tar.Close(); err != nil {
			log.Fatalf("Failed to close tar: %v", err)
		}
	}

	if err := os.RemoveAll(bundleDir + "/bundles"); err != nil {
		log.Fatalf("Failed to remove %v/bundles: %v\n", bundleDir, err)
	}

	fmt.Printf("Extracted bundle to %v\n", bundleDir)
}

// GetFileContentType : Gets file content type
func GetFileContentType(out *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// extractCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// extractCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
