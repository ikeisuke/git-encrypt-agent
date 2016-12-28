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
	"path"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		keyId := cmd.Flags().Lookup("key-id").Value.String()
		if len(keyId) > 0 {
			// TODO: 上書きチェック
			sess, err := session.NewSessionWithOptions(session.Options{
			     Config: aws.Config{Region: aws.String("ap-northeast-1")},
			     Profile: "gitencrypt",
			})
			k := kms.New(sess)
			r := kms.GenerateDataKeyWithoutPlaintextInput{}
			r.SetKeyId(keyId)
			r.SetKeySpec("AES_256")
			data, err := k.GenerateDataKeyWithoutPlaintext(&r);
			if err != nil {
				fmt.Printf("Failed to generate data key, %v\n", err)
				os.Exit(-1)
			}
			datakeyfile := path.Join(projectRootDir, ".gitdatakey")
			err = ioutil.WriteFile(datakeyfile, data.CiphertextBlob, 0644)
			if err != nil {
				fmt.Printf("Failed to save config file, %v\n", err)
				os.Exit(-1)
			}
		}
		status, err := exec.Command("git", "config", "--get", "--local", "merge.renormalize").Output()
		if err != nil {
			err := exec.Command("git", "config", "--add", "--local", "merge.renormalize").Run()
			if err != nil {
				fmt.Printf("Failed to save merge.renormalize, %v\n", err)
				os.Exit(-1)
			}
		}
		statusString := string(status[:len(status)-1])
		if "true" != statusString {
			var input string
			for {
				fmt.Printf("Overwrite merge.renormalize to true from %s [Y/n]: ", statusString);
				fmt.Scanln(&input)
				if len(input) == 0 {
					exec.Command("git", "config", "--replace-all", "--local", "merge.renormalize", "true").Run()
					break;
				}
				first := strings.ToLower(input[:1])
				if first == "y" {
					exec.Command("git", "config", "--replace-all", "--local", "merge.renormalize", "true").Run()
					break;
				} else if first == "n" {
					fmt.Printf("Install failed, git-encrypt required merge.renormalize to be true\n")
					os.Exit(-1)
				}
			}
		}
		exec.Command("git", "config", "--replace-all", "--local", "filter.encrypt.clean", "git encrypt clean").Run()
		exec.Command("git", "config", "--replace-all", "--local", "filter.encrypt.smudge", "git encrypt smudge").Run()
		exec.Command("git", "config", "--replace-all", "--local", "diff.encrypt.textconv", "git encrypt smudge").Run()
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
	installCmd.Flags().String("key-id", "", "Generate encrypted data key by specified Key ID. A valid indentifier is Key ID, CMK ARN, Alias and Alias ARN")
}
