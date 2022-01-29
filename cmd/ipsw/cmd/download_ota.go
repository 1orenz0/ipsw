/*
Copyright © 2021 blacktop

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
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/apex/log"
	"github.com/blacktop/ipsw/internal/download"
	"github.com/blacktop/ipsw/internal/utils"
	"github.com/blacktop/ipsw/pkg/kernelcache"
	"github.com/blacktop/ipsw/pkg/ota"
	semver "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	downloadCmd.AddCommand(otaDLCmd)

	otaDLCmd.Flags().StringP("platform", "p", "ios", "Platform to download (ios, macos, watchos, tvos, audioos)")
	otaDLCmd.Flags().Bool("beta", false, "Download Beta OTAs")
	otaDLCmd.Flags().Bool("dyld", false, "Extract dyld_shared_cache from remote OTA zip")
	otaDLCmd.Flags().Bool("kernel", false, "Extract kernelcache from remote OTA zip")
	otaDLCmd.Flags().Bool("info", false, "Show all the latest OTAs available")
	otaDLCmd.Flags().String("info-type", "", "OS type to show OTAs for")
	otaDLCmd.Flags().StringP("output", "o", "", "Folder to download files to")
	viper.BindPFlag("download.ota.platform", otaDLCmd.Flags().Lookup("platform"))
	viper.BindPFlag("download.ota.beta", otaDLCmd.Flags().Lookup("beta"))
	viper.BindPFlag("download.ota.dyld", otaDLCmd.Flags().Lookup("dyld"))
	viper.BindPFlag("download.ota.kernel", otaDLCmd.Flags().Lookup("kernel"))
	viper.BindPFlag("download.ota.info", otaDLCmd.Flags().Lookup("info"))
	viper.BindPFlag("download.ota.info-type", otaDLCmd.Flags().Lookup("info-type"))
	viper.BindPFlag("download.ota.output", otaDLCmd.Flags().Lookup("output"))
}

// otaDLCmd represents the ota download command
var otaDLCmd = &cobra.Command{
	Use:           "ota [options]",
	Short:         "Download OTAs",
	SilenceUsage:  false,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		var err error
		var ver *semver.Version

		if Verbose {
			log.SetLevel(log.DebugLevel)
		}

		viper.BindPFlag("download.proxy", cmd.Flags().Lookup("proxy"))
		viper.BindPFlag("download.insecure", cmd.Flags().Lookup("insecure"))
		viper.BindPFlag("download.confirm", cmd.Flags().Lookup("confirm"))
		viper.BindPFlag("download.skip-all", cmd.Flags().Lookup("skip-all"))
		viper.BindPFlag("download.resume-all", cmd.Flags().Lookup("resume-all"))
		viper.BindPFlag("download.restart-all", cmd.Flags().Lookup("restart-all"))
		viper.BindPFlag("download.remove-commas", cmd.Flags().Lookup("remove-commas"))
		viper.BindPFlag("download.white-list", cmd.Flags().Lookup("white-list"))
		viper.BindPFlag("download.black-list", cmd.Flags().Lookup("black-list"))
		viper.BindPFlag("download.device", cmd.Flags().Lookup("device"))
		viper.BindPFlag("download.model", cmd.Flags().Lookup("model"))
		viper.BindPFlag("download.version", cmd.Flags().Lookup("version"))
		viper.BindPFlag("download.build", cmd.Flags().Lookup("build"))

		// settings
		proxy := viper.GetString("download.proxy")
		insecure := viper.GetBool("download.insecure")
		confirm := viper.GetBool("download.confirm")
		skipAll := viper.GetBool("download.skip-all")
		resumeAll := viper.GetBool("download.resume-all")
		restartAll := viper.GetBool("download.restart-all")
		removeCommas := viper.GetBool("download.remove-commas")
		// filters
		device := viper.GetString("download.device")
		model := viper.GetString("download.model")
		version := viper.GetString("download.version")
		build := viper.GetString("download.build")
		doDownload := viper.GetStringSlice("download.white-list")
		doNotDownload := viper.GetStringSlice("download.black-list")
		// flags
		platform := viper.GetString("download.ota.platform")
		getBeta := viper.GetBool("download.ota.beta")
		remoteDyld := viper.GetBool("download.ota.dyld")
		remoteKernel := viper.GetBool("download.ota.kernel")
		otaInfo := viper.GetBool("download.ota.info")
		otaInfoType := viper.GetString("download.ota.info-type")
		output := viper.GetString("download.ota.output")

		// verify args
		// if kernel && len(pattern) > 0 {
		// 	return fmt.Errorf("you cannot supply a --kernel AND a --pattern (they are mutually exclusive)")
		// }
		// if len(version) > 0 && len(build) > 0 {
		// 	log.Fatal("you cannot supply a --version AND a --build (they are mutually exclusive)")
		// }

		// Query for asset sets
		as, err := download.GetAssetSets(proxy, insecure)
		if err != nil {
			log.Fatal(err.Error())
		}

		/****************
		 * GET OTA INFO *
		 ****************/
		if otaInfo {
			if len(device) > 0 {
				log.WithField("device", device).Info("OTAs")
				for _, asset := range as.ForDevice(device) {
					utils.Indent(log.WithFields(log.Fields{
						"posting_date":    asset.PostingDate,
						"expiration_date": asset.ExpirationDate,
					}).Info, 2)(asset.ProductVersion)
				}
			} else {
				if len(otaInfoType) == 0 {
					prompt := &survey.Select{
						Message: "Choose an OS type:",
						Options: []string{"iOS", "macOS"},
					}
					survey.AskOne(prompt, &otaInfoType)
				} else {
					if !utils.StrSliceHas([]string{"iOS", "macOS"}, otaInfoType) {
						log.Fatal("you must supply a valid --info-type flag: (iOS, macOS)")
					}
				}
				log.WithField("type", otaInfoType).Info("OTAs")
				if otaInfoType == "iOS" {
					log.Warn("⚠️  This includes: iOS, iPadOS, watchOS, tvOS and audioOS (you can filter by adding the --device flag)")
				}
				for _, asset := range as.AssetSets[otaInfoType] {
					utils.Indent(log.WithFields(log.Fields{
						"posting_date":    asset.PostingDate,
						"expiration_date": asset.ExpirationDate,
					}).Info, 2)(asset.ProductVersion)
				}
			}
			return nil
		}

		if !utils.StrSliceHas(
			[]string{"ios", "macos", "watchos", "tvos", "audioos"}, strings.ToLower(platform)) {
			log.Fatal("you must supply a valid --platform flag. Choices are: ios, macos, watchos, tvos and audioos")
		}

		var destPath string
		if len(output) > 0 {
			destPath = filepath.Clean(output)
		}

		if len(version) > 0 {
			ver, err = semver.NewVersion(version)
			if err != nil {
				log.Fatal("failed to convert version into semver object")
			}
		}

		otaXML, err := download.NewOTA(as, download.OtaConf{
			Platform:        strings.ToLower(platform),
			Beta:            getBeta,
			Device:          device,
			Model:           model,
			Version:         ver,
			Build:           build,
			DeviceWhiteList: doDownload,
			DeviceBlackList: doNotDownload,
			Proxy:           proxy,
			Insecure:        insecure,
		})
		if err != nil {
			return fmt.Errorf("failed to parse remote OTA XML: %v", err)
		}
		// otas := otaXML.FilterOtaAssets(doDownload, doNotDownload) FIXME: integrate the white-list into the filter AND as a device list (if no device is given)
		// if len(otas) == 0 {
		// 	log.Fatal(fmt.Sprintf("no OTAs match device %s %s", device, doDownload))
		// }
		otas, err := otaXML.GetPallasOTAs()
		if err != nil {
			return err
		}

		if Verbose {
			for _, o := range otas {
				log.WithFields(log.Fields{
					"device":  strings.Join(o.SupportedDevices, " "),
					"build":   o.Build,
					"version": o.OSVersion,
					// "url":     o.RelativePath,
				}).Info("OTA")
			}
			// return nil
		}

		log.Debug("URLs to Download:")
		for _, o := range otas {
			utils.Indent(log.Debug, 2)(o.BaseURL + o.RelativePath)
		}

		cont := true
		if !confirm {
			// if filtered to a single device skip the prompt
			if len(otas) > 1 {
				cont = false
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("You are about to download %d OTA files. Continue?", len(otas)),
				}
				survey.AskOne(prompt, &cont)
			}
		}

		if cont {
			if remoteDyld || remoteKernel {
				for _, o := range otas {
					log.WithFields(log.Fields{
						"device": strings.Join(o.SupportedDevices, " "),
						"model":  strings.Join(o.SupportedDeviceModels, " "),
						"build":  o.Build,
						"type":   o.DocumentationID,
					}).Info(fmt.Sprintf("Getting %s %s remote OTA", o.ProductSystemName, strings.TrimPrefix(o.OSVersion, "9.9.")))
					zr, err := download.NewRemoteZipReader(o.BaseURL+o.RelativePath, &download.RemoteConfig{
						Proxy:    proxy,
						Insecure: insecure,
					})
					if err != nil {
						return fmt.Errorf("failed to open remote zip to OTA: %v", err)
					}
					if remoteKernel {
						log.Info("Extracting remote kernelcache")
						err = kernelcache.RemoteParse(zr, destPath)
						if err != nil {
							return fmt.Errorf("failed to download kernelcache from remote ota: %v", err)
						}
					}
					if remoteDyld {
						log.Info("Extracting remote dyld_shared_cache (can be a bit CPU intensive)")
						err = ota.RemoteExtract(zr, "^System/Library/.*/dyld_shared_cache.*$", destPath)
						if err != nil {
							return fmt.Errorf("failed to download dyld_shared_cache from remote ota: %v", err)
						}
					}
				}
			} else {
				downloader := download.NewDownload(proxy, insecure, skipAll, resumeAll, restartAll, Verbose)
				for _, o := range otas {
					folder := filepath.Join(destPath, fmt.Sprintf("%s%s_OTAs", o.ProductSystemName, strings.TrimPrefix(o.OSVersion, "9.9.")))
					os.MkdirAll(folder, os.ModePerm)
					var devices string
					if len(o.SupportedDevices) > 0 {
						devices = strings.Join(o.SupportedDevices, "_")
					} else {
						devices = strings.Join(o.SupportedDeviceModels, "_")
					}
					url := o.BaseURL + o.RelativePath
					destName := filepath.Join(folder, fmt.Sprintf("%s_%s", devices, getDestName(url, removeCommas)))
					if _, err := os.Stat(destName); os.IsNotExist(err) {
						log.WithFields(log.Fields{
							"device": strings.Join(o.SupportedDevices, " "),
							"model":  strings.Join(o.SupportedDeviceModels, " "),
							"build":  o.Build,
							"type":   o.DocumentationID,
						}).Info(fmt.Sprintf("Getting %s %s OTA", o.ProductSystemName, strings.TrimPrefix(o.OSVersion, "9.9.")))
						// download file
						downloader.URL = url
						downloader.DestName = destName
						err = downloader.Do()
						if err != nil {
							return fmt.Errorf("failed to download file: %v", err)
						}
					} else {
						log.Warnf("ota already exists: %s", destName)
					}
				}
			}
		}

		return nil
	},
}
