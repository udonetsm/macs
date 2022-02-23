package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/spf13/cobra"
)

const (
	ERR_OPEN_FILE   = 1000
	ERR_WRITE_FILE  = 1001
	ERR_WIFI_SCAN   = 1002
	ERR_PARSE_FLAGS = 1003
	ERR_CHMOD       = 1004
)

func errors(err error, caller func()) {
	if err != nil {
		caller()
	}
}

func err_chmod(args ...interface{}) {
	log.Printf("Can't chmod %s. Sure you are sudo?\nErr: %s with code %v\n", args[0], args[1], args[2])
}

func err_openfile(args ...interface{}) {
	log.Printf("Can't open %s\nErr: %s\n", args[0], args[1])
	os.Exit(args[2].(int))
}

func err_writefile(args ...interface{}) {
	log.Printf("Can't write in %s\nErr: %s\n", args[0], args[1])
	os.Exit(args[2].(int))
}

func err_scanning(args ...interface{}) {
	log.Printf("Can't scanning wifis\nErr: %s with code %v\n", args[0], args[1])
}

func err_parseflags(args ...interface{}) {
	log.Printf("Can't parse flags\nErr: %s\n", args[0])
	os.Exit(args[1].(int))
}

/* For parse flag */
var (
	wifiinterface, address string
	scanningtime, interval int64
)

/* Writting data in target file */
func writeInFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)
	errors(err, func() { err_openfile(filename, err, ERR_OPEN_FILE) })
	defer file.Close()
	_, err = file.WriteString(data)
	errors(err, func() { err_writefile(filename, err, ERR_WRITE_FILE) })
}

/* Makes scanning within a minute and write macs in macs.txt for analizing */
func Macs(cmd *cobra.Command, arg []string) {
	var count int
	var now int64
	var strmacs, filename string
	filename = address + ".txt"
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	start := time.Now().Unix()
	go func() {
		fmt.Printf("Passed\t | \tScanned\t | \tLeft\n")
		for sec := range ticker.C {
			macs, err := wifiscan.Scan(wifiinterface)
			errors(err, func() { err_scanning(err, ERR_WIFI_SCAN) })
			for _, mac := range macs {
				count++
				strmacs += mac.SSID + "\n"
			}
			now = sec.Unix()
			now -= start
			writeInFile(filename, strmacs+"\n")
			fmt.Printf("%v(sec)\t | \t%v(macs)\t | \t%v(sec)\n", now, len(macs), scanningtime-now)
			strmacs = ""
		}
	}()
	time.Sleep(time.Duration(scanningtime) * time.Second)
	ticker.Stop()
	fmt.Printf("Wrote %v strings in %s\n", count, filename)
	err := os.Chmod(address+".txt", 0777)
	errors(err, func() { err_chmod(filename, err, ERR_CHMOD) })
	wifiinterface, address, filename = "", "", ""
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
	rootCmd.Flags().Int64VarP(&scanningtime, "time", "t", 30, "set scanning time")
	rootCmd.Flags().Int64VarP(&interval, "sleep", "s", 13, "set scanning interval")
	rootCmd.MarkFlagRequired("addr")
	err := rootCmd.Execute()
	errors(err, func() { err_parseflags(err, ERR_PARSE_FLAGS) })
}

func main() {
	parse_flags()
}
