# ms213x-rename

本工具能够提取，修改MS2130/MS2131采集卡EDID信息

## 安装

`go install github.com/starainrt/ms213x-rename`或者下载relese中的预编译版本


## 使用

```
ms213x-rename.exe -h
This tool can help you modify the basic EDID information in the MS213X collector firmware, such as the display name, serial number etc...

Usage:
   [flags]

Flags:
  -t, --attach-edid string          Attach a EDID bin file to the firmware
  -d, --display-name string         The display name to set in the EDID(13 ascii characters max)
  -e, --dump-edid string            Dump the EDID from the firmware to a file
  -h, --help                        help for this command
  -m, --manufacturer-id string      The manufacturer ID to set in the EDID(3 uppercase ascii characters)
  -p, --product-code uint16         The product code to set in the EDID
  -a, --save-edid string            Save the modified EDID to a file
  -o, --save-firmware string        Save the modified firmware to a file
  -s, --serial-number uint32        The serial number to set in the EDID
  -v, --show-detail                 Show the detail information of the EDID
  -w, --week-of-manufacture uint8   The week of manufacture to set in the EDID(0-54)
  -y, --year-of-manufacture int     The year of manufacture to set in the EDID(1990-2245)
```

## 使用方法
1. 使用 MS21XX&91XXDownloadTool 工具连接采集卡并备份固件到本地
2. 使用此工具修改EDID相应的信息
3. 使用 MS21XX&91XXDownloadTool 工具刷写修改后的固件到采集卡
4. 重新插拔采集卡生效

参考：[https://www.b612.me/default/301.html](https://www.b612.me/default/301.html)