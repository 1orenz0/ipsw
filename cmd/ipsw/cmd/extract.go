/*
Copyright © 2019 blacktop

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/apex/log"
	"github.com/blacktop/ipsw/devicetree"
	"github.com/blacktop/ipsw/dyld"
	"github.com/blacktop/ipsw/kernelcache"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"howett.net/ranger"
)

var (
	kernelFlag     bool
	dyldFlag       bool
	deviceTreeFlag bool
	remote         bool
)

func init() {
	rootCmd.AddCommand(extractCmd)

	extractCmd.Flags().BoolVarP(&remote, "remote", "r", false, "Extract from URL")
	extractCmd.Flags().BoolVarP(&kernelFlag, "kernel", "k", false, "Extract kernelcache")
	extractCmd.Flags().BoolVarP(&dyldFlag, "dyld", "d", false, "Extract dyld_shared_cache")
	extractCmd.Flags().BoolVarP(&deviceTreeFlag, "device-tree", "t", false, "Extract DeviceTree")
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [path to IPSW | URL]",
	Short: "Extract kernelcache, dyld_shared_cache or DeviceTree from IPSW",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		if remote {
			if !isURL(args[0]) {
				log.Fatal("must supply valid URL when using the remote flag")
			}

			if deviceTreeFlag {
				log.Error("unable to extract DeviceTree remotely (for now)")
				// err := devicetree.RemoteParse(args[0])
				// if err != nil {
				// 	return errors.Wrap(err, "failed to extract DeviceTree")
				// }
			}
			if dyldFlag {
				log.Error("unable to extract dyld_shared_cache remotely")
			}
			if kernelFlag {
				log.Info("Extracting Kernelcache")
				url, err := url.Parse(args[0])
				if err != nil {
					return errors.Wrap(err, "failed to parse url")
				}
				reader, err := ranger.NewReader(&ranger.HTTPRanger{URL: url})
				if err != nil {
					return errors.Wrap(err, "failed to create ranger reader")
				}
				length, err := reader.Length()
				if err != nil {
					return errors.Wrap(err, "failed to get reader length")
				}
				zr, err := zip.NewReader(reader, length)
				if err != nil {
					return errors.Wrap(err, "failed to create zip reader from ranger reader")
				}

				for _, f := range zr.File {
					if strings.Contains(f.Name, "kernel") {
						kdata := make([]byte, f.UncompressedSize64)
						rc, _ := f.Open()
						io.ReadFull(rc, kdata)
						rc.Close()
						kcomp, err := kernelcache.ParseImg4Data(kdata)
						if err != nil {
							return errors.Wrap(err, "failed parse compressed kernelcache")
						}
						err = ioutil.WriteFile(f.Name+".decompressed", kernelcache.DecompressData(kcomp), 0644)
						if err != nil {
							return errors.Wrap(err, "failed to decompress kernelcache")
						}
					}
				}
			}
		} else {
			if _, err := os.Stat(args[0]); os.IsNotExist(err) {
				return fmt.Errorf("file %s does not exist", args[0])
			}

			if kernelFlag {
				log.Info("Extracting kernelcaches")
				err := kernelcache.Extract(args[0])
				if err != nil {
					return errors.Wrap(err, "failed to extract kernelcaches")
				}
			}

			if dyldFlag {
				log.Info("Extracting dyld_shared_cache")
				if runtime.GOOS != "darwin" {
					log.Fatal("dyld_shared_cache extraction only works on macOS :(")
				}
				err := dyld.Extract(args[0])
				if err != nil {
					return errors.Wrap(err, "failed to extract dyld_shared_cache")
				}
			}

			if deviceTreeFlag {
				log.Info("Extracting DeviceTrees")
				err := devicetree.Extract(args[0])
				if err != nil {
					return errors.Wrap(err, "failed to extract DeviceTrees")
				}
			}
		}

		return nil
	},
}