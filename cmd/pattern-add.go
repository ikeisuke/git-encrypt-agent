// Copyright © 2016 Keisuke Isono <reirou.k@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var patternAddCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		pattern := cmd.Flags().Lookup("pattern").Value.String()
		if len(pattern) == 0 {
			fmt.Println("Error: argument --pattern is required")
			os.Exit(-1)
		}
		attributeFile := path.Join(projectRootDir, ".gitattributes")
		_, err := os.Stat(attributeFile)
		if err != nil {
			if os.IsNotExist(err) {
				ioutil.WriteFile(attributeFile, []byte{}, 0644)
			} else {
				fmt.Printf("Failed to open .gitattributes, %v\n", err)
				os.Exit(-1)
			}
		}
		//attrtext := fmt.Sprintf("%s filter=git-encrypt diff=git-encrypt", pattern)
		f, err := os.Open(attributeFile)
		if err != nil {
			fmt.Printf("Failed to read file: %v, %v\n", attributeFile, err)
			os.Exit(-1)
		}
		filename := fmt.Sprintf("%v.tmp", attributeFile)
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
		file.Truncate(0)
		if err != nil {
			fmt.Printf("Failed to read file: %v, %v\n", attributeFile, err)
			os.Exit(-1)
		}
		err = file.Truncate(0)
		if err != nil {
			fmt.Printf("Failed to read file: %v, %v\n", attributeFile, err)
			os.Exit(-1)
		}
		defer func() {
			f.Close()
			file.Close()
			os.Rename(filename, attributeFile)
		}()
		scanner := bufio.NewScanner(f)
		writer := bufio.NewWriter(file)
		var exist bool
		append := fmt.Sprintf("%s filter=git-encrypt diff=git-encrypt", pattern)
		for scanner.Scan() {
			line := scanner.Text()
			writer.WriteString(line + "\n")
			if append == line {
				exist = true
			}
		}
		if !exist {
			writer.WriteString(append)
		}
		writer.Flush()
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "File %s scan error: %v\n", attributeFile, err)
		}
	},
}

func init() {
	patternCmd.AddCommand(patternAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	patternAddCmd.Flags().String("pattern", "", "Help message for pattern")

}
