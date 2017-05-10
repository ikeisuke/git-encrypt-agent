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
	"fmt"
	"os"
	"path"
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/ikeisuke/git-encrypt-agent/config"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure git-encrypt options.",
	Long: `Configure git-encrypt options.
If this command is run, you wil be prompted for configuration values
such as your AWS Profile Name and AWS Region Name.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := config.Load(projectGitDir)
		profile := waitInput("AWS Profile Name", c.AWSProfileName)
		region := waitInput("AWS Region Name", c.AWSRegionName)
		c.AWSProfileName = profile;
		c.AWSRegionName = region;
		err := config.Save(projectGitDir, c)
		if err != nil {
			fmt.Printf("Failed to save config file, %v\n", err)
			os.Exit(-1)
		}
	},
}

func waitInput(text string, value string) string {
	var input string
	var displayValue string
	if (len(value) == 0) {
		displayValue = "None"
	} else {
		displayValue = value
	}
	fmt.Printf("%s [%s]: ", text, displayValue);
	fmt.Scanln(&input)
	if len(input) == 0 {
		return value
	}
	return input
}

func init() {
	RootCmd.AddCommand(configureCmd)
}
