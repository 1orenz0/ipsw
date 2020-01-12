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
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/blacktop/ipsw/pkg/dyld"
	"github.com/spf13/cobra"
)

func init() {
	dyldCmd.AddCommand(dyldListCmd)

	dyldListCmd.MarkZshCompPositionalArgumentFile(1, "dyld_shared_cache*")
}

// dyldListCmd represents the list command
var dyldListCmd = &cobra.Command{
	Use:   "list <dyld_shared_cache>",
	Short: "List all dylibs in dyld_shared_cache",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		if _, err := os.Stat(args[0]); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", args[0])
		}

		// TODO: check for
		// if ( dylibInfo->isAlias )
		//   	printf("[alias] %s\n", dylibInfo->path);

		f, err := dyld.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()

		f.CacheHeader.Print()
		f.LocalSymInfo.Print()
		f.Mappings.Print()

		// for idx, img := range f.Images {
		// 	fmt.Printf("%4d:\t0x%0X\t%s\n", idx+1, img.Info.Address, img.Name)
		// }

		image := f.Image("/System/Library/Frameworks/WebKit.framework/WebKit")
		fmt.Println(image.Info.String())
		fmt.Println(image.UUID)
		fmt.Println("DylibOffset:", image.DylibOffset)
		fmt.Println("Calc DylibOffset:", image.Info.Address-f.Mappings[0].Address)

		// for _, sym := range image.LocalSymbols {
		// 	fmt.Printf("0x%0X:\t%s\n", sym.Value, sym.Name)
		// }

		return nil
	},
}
