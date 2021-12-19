---
date: 2021-12-17T21:43:37-07:00
title: "ipsw download tss"
slug: ipsw_download_tss
url: /commands/ipsw_download_tss/
---
## ipsw download tss

🚧 Download SHSH Blobs

```
ipsw download tss [flags]
```

### Options

```
  -h, --help            help for tss
  -s, --signed string   Check if iOS version is still being signed
```

### Options inherited from parent commands

```
      --black-list stringArray   iOS device black list
  -b, --build string             iOS BuildID (i.e. 16F203)
      --config string            config file (default is $HOME/.ipsw.yaml)
  -y, --confirm                  do not prompt user for confirmation
  -d, --device string            iOS Device (i.e. iPhone11,2)
      --insecure                 do not verify ssl certs
  -m, --model string             iOS Model (i.e. D321AP)
      --proxy string             HTTP/HTTPS proxy
  -_, --remove-commas            replace commas in IPSW filename with underscores
      --restart-all              always restart resumable IPSWs
      --resume-all               always resume resumable IPSWs
      --skip-all                 always skip resumable IPSWs
  -V, --verbose                  verbose output
  -v, --version string           iOS Version (i.e. 12.3.1)
      --white-list stringArray   iOS device white list
```

### SEE ALSO

* [ipsw download](/cmd/ipsw_download/)	 - Download Apple Firmware files (and more)

