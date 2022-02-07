package main

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	ERR_OPEN_FILE   = 1000
	ERR_WRITE_FILE  = 1001
	ERR_WIFI_SCAN   = 1002
	ERR_PARSE_FLAGS = 1003
	ERR_CHMOD       = 1004
)

/* For parse flag */
var (
	wifiinterface, address string
	scanningtime, interval int
)

func errors(err error, code int, fn func(err error, code int, args ...interface{})) {
	if err != nil {
		fn(err, code)
	}
}

func err_open_file(err error, code int, args ...interface{}) {
	logrus.Error("Something wrong while open file\n", err, args)
	os.Exit(code)
}

func err_write_file(err error, code int, args ...interface{}) {
	logrus.Error("Something wrong while write file\n", err, args)
	os.Exit(code)
}

func err_wifi_scan(err error, code int, args ...interface{}) {
	logrus.Error("Something wrong while scanning wifi\n.", err, args)
}

func err_parse_flags(err error, code int, args ...interface{}) {
	logrus.Error("Something wrong while parsing flags\n", err, args)
	os.Exit(code)
}

func err_chmod(err error, code int, args ...interface{}) {
	logrus.Error("Sumthing wrong while chmod\n", err, code, args)
	os.Exit(code)
}

/* Writting data in target file */
func writeInFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)
	errors(err, ERR_OPEN_FILE, err_open_file)
	defer file.Close()
	_, err = file.WriteString(data)
	errors(err, ERR_WRITE_FILE, err_write_file)
}

/* Makes scanning within a minute and write macs in macs.txt for analizing */
func Macs(cmd *cobra.Command, arg []string) {
	logrus.Info("Scanning...")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	var count, countmacs int
	go func() {
		for sec := range ticker.C {
			countmacs = 0
			macs, err := wifiscan.Scan(wifiinterface)
			errors(err, ERR_WIFI_SCAN, err_wifi_scan)
			for _, mac := range macs {
				writeInFile(address+".txt", mac.SSID+"\n")
				count++
				countmacs++
			}
			fmt.Printf("Scanned %v macs in %v:%v:%v\n", countmacs, sec.Hour(), sec.Minute(), sec.Second())
		}
	}()
	time.Sleep(time.Duration(scanningtime) * time.Second)
	ticker.Stop()
	err := os.Chmod(address+".txt", 0777)
	errors(err, ERR_CHMOD, err_chmod)
	fmt.Printf("Wrote %v strings in %s.txt\n", count, address)
	wifiinterface, address = "", ""
	count, scanningtime, interval = 0, 0, 0
}

/* Get network interface name as flag and show help message if it's not exist */
func parse_flags() {
	rootCmd := &cobra.Command{
		Use:     "macs",
		Version: "1.0",
		Example: "This example will scan wifis within 120 seconds every 5 seconds and write scanned macs in Ленина_5-2.txt\n\n" +
			"\t" + `sudo macs -a Ленина_5-2 -t 120 -s 5`,
		Run: Macs,
	}

	rootCmd.Flags().StringVarP(&wifiinterface, "iface", "i", "", "set outgoing interface")
	rootCmd.Flags().StringVarP(&address, "addr", "a", "", "set address")
	rootCmd.Flags().IntVarP(&scanningtime, "time", "t", 30, "set scanning time")
	rootCmd.Flags().IntVarP(&interval, "sleep", "s", 13, "set scanning interval")
	rootCmd.MarkFlagRequired("addr")
	err := rootCmd.Execute()
	errors(err, ERR_PARSE_FLAGS, err_parse_flags)
}

func main() {
	parse_flags()
}
