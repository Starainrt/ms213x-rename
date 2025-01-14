package main

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	displayName       string
	serialNumber      uint32
	manufacturerID    string
	productCode       uint16
	weekOfManufacture uint8
	yearOfManufacture int
	attachEdidPath    string

	dumpEdidPath     string
	saveEdidPath     string
	saveFirmwarePath string
	showDetail       bool
)

func init() {
	cmdMain.Flags().StringVarP(&attachEdidPath, "attach-edid", "t", "", "Attach a EDID bin file to the firmware")
	cmdMain.Flags().StringVarP(&displayName, "display-name", "d", "", "The display name to set in the EDID(13 ascii characters max)")
	cmdMain.Flags().Uint32VarP(&serialNumber, "serial-number", "s", 0, "The serial number to set in the EDID")
	cmdMain.Flags().StringVarP(&manufacturerID, "manufacturer-id", "m", "", "The manufacturer ID to set in the EDID(3 uppercase ascii characters)")
	cmdMain.Flags().Uint16VarP(&productCode, "product-code", "p", 0, "The product code to set in the EDID")
	cmdMain.Flags().Uint8VarP(&weekOfManufacture, "week-of-manufacture", "w", 0, "The week of manufacture to set in the EDID(0-54)")
	cmdMain.Flags().IntVarP(&yearOfManufacture, "year-of-manufacture", "y", 0, "The year of manufacture to set in the EDID(1990-2245)")
	cmdMain.Flags().StringVarP(&dumpEdidPath, "dump-edid", "e", "", "Dump the EDID from the firmware to a file")
	cmdMain.Flags().StringVarP(&saveEdidPath, "save-edid", "a", "", "Save the modified EDID to a file")
	cmdMain.Flags().StringVarP(&saveFirmwarePath, "save-firmware", "o", "", "Save the modified firmware to a file")
	cmdMain.Flags().BoolVarP(&showDetail, "show-detail", "v", false, "Show the detail information of the EDID")
	cmdMain.SilenceUsage = true
	cmdMain.AddCommand(cmdCheck)
}

var cmdMain = &cobra.Command{
	Short: "MS213X Collector Display Name Modification Tool",
	Long:  "This tool can help you modify the basic EDID information in the MS213X collector firmware, such as the display name, serial number etc...",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Printf("Please specify a firmware or edid bin file.")
			return errors.New("no file specified")
		}
		if shouldModifiyEdid() && attachEdidPath != "" {
			log.Printf("You can not modify the EDID and attach a EDID file at the same time.")
			return errors.New("conflict options")
		}
		f, err := os.Open(args[0])
		if err != nil {
			log.Printf("Open file %s failed: %v", args[0], err)
			return err
		}
		defer f.Close()

		edid, err := loadEdidFromFile(f)
		if err != nil {
			log.Printf("could not find EDID: %v", err)
			return err
		}
		if showDetail {
			edid.PrettyPrint()
		} else {
			edid.PrettyPrintShort()
		}
		if dumpEdidPath != "" {
			err = os.WriteFile(dumpEdidPath, edid.ToBytes(), 0644)
			if err != nil {
				log.Printf("write EDID to file %s failed: %v", dumpEdidPath, err)
				return err
			}
			log.Printf("Dump EDID to %s Ok", dumpEdidPath)
		}
		if !shouldModifiyEdid() && attachEdidPath == "" {
			return nil
		}
		if shouldModifiyEdid() {
			if err = ChangeEdid(edid); err != nil {
				return err
			}
		} else {
			edidFile, err := os.Open(attachEdidPath)
			if err != nil {
				log.Printf("read EDID file %s failed: %v", attachEdidPath, err)
				return err
			}
			edid, err = loadEdidFromFile(edidFile)
			if err != nil {
				log.Printf("parse EDID file %s failed: %v", attachEdidPath, err)
				return err
			}
		}
		var lastErr error
		if saveEdidPath != "" {
			err = os.WriteFile(saveEdidPath, edid.ToBytes(), 0644)
			if err != nil {
				lastErr = err
				log.Printf("write modified EDID to file %s failed: %v", saveEdidPath, err)
			}
			log.Printf("Save modified EDID to %s Ok", saveEdidPath)
		}
		if saveFirmwarePath != "" {
			of, err := os.Create(saveFirmwarePath)
			if err != nil {
				log.Printf("create firmware file %s failed: %v", saveFirmwarePath, err)
				return err
			}
			err = applyFirstEDID2NewFile(f, of, edid)
			if err != nil {
				log.Printf("apply EDID to firmware failed: %v", err)
				of.Close()
				return err
			}
			of.Close()
			log.Printf("Save modified firmware to %s Ok", saveFirmwarePath)
			srcHash, err := hashWithoutEdid(args[0])
			if err != nil {
				log.Printf("Failed to calculate the hash of original firmware:%v; do not use the generated firmware.", err)
				return err
			}
			dstHash, err := hashWithoutEdid(saveFirmwarePath)
			if err != nil {
				log.Printf("Failed to calculate the hash of generated firmware:%v; do not use the generated firmware.", err)
				return err
			}
			if srcHash != dstHash {
				log.Printf("fireware hash verification failed,original hash:%s,generated hash:%s,not equal,!! do not use the generated firmware !!", srcHash, dstHash)
				return errors.New("verify hash failed")
			}
		}
		return lastErr
	},
}

var cmdCheck = &cobra.Command{
	Use:   "hash",
	Short: "removing the EDID part from the old and new firmware for hash verification",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcHash, err := hashWithoutEdid(args[0])
		if err != nil {
			log.Printf("Calc src firmware hash failed:%v", err)
			return err
		}
		dstHash, err := hashWithoutEdid(args[1])
		if err != nil {
			log.Printf("Calc saved firmware hash failed:%v", err)
			return err
		}
		log.Printf("Src firmware hash:%s", srcHash)
		log.Printf("Saved firmware hash:%s", dstHash)
		if srcHash == dstHash {
			log.Printf("Verify firmware hash OK")
		} else {
			log.Printf("Verify firmware hash failed")
			return errors.New("verify hash failed")
		}
		return nil
	},
}

func shouldModifiyEdid() bool {
	return displayName != "" || serialNumber != 0 || manufacturerID != "" || productCode != 0 || weekOfManufacture != 0 || yearOfManufacture != 0
}

func ChangeEdid(edid *Edid) error {
	var err error
	if displayName != "" {
		log.Printf("Change display name from %s to %s", edid.MonitorName, displayName)
		err = edid.SetMonitorName(displayName)
		if err != nil {
			log.Printf("set display name failed: %v", err)
			return err
		}
	}
	if serialNumber != 0 {
		log.Printf("Change serial number from %d to %d", edid.SerialNumber, serialNumber)
		err = edid.SetSerialNumber(serialNumber)
		if err != nil {
			log.Printf("set serial number failed: %v", err)
			return err
		}
	}
	if manufacturerID != "" {
		log.Printf("Change manufacturer id from %s to %s", edid.ManufacturerId, manufacturerID)
		err = edid.SetManufacturerId(manufacturerID)
		if err != nil {
			log.Printf("set manufacturer id failed: %v", err)
			return err
		}
	}
	if productCode != 0 {
		log.Printf("Change product code from %d to %d", edid.ProductCode, productCode)
		err = edid.SetProductCode(productCode)
		if err != nil {
			log.Printf("set product code failed: %v", err)
			return err
		}
	}
	if weekOfManufacture != 0 {
		log.Printf("Change week of manufacture from %d to %d", edid.WeekOfManufacture, weekOfManufacture)
		err = edid.SetWeekOfManufacture(weekOfManufacture)
		if err != nil {
			log.Printf("set week of manufacture failed: %v", err)
			return err
		}
	}
	if yearOfManufacture != 0 {
		log.Printf("Change year of manufacture from %d to %d", edid.YearOfManufacture, yearOfManufacture)
		err = edid.SetYearOfManufacture(yearOfManufacture)
		if err != nil {
			log.Printf("set year of manufacture failed: %v", err)
			return err
		}
	}
	return nil
}
