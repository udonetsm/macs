package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/spf13/cobra"
	"github.com/udonetsm/handlesignals"
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
	log.Printf("Can't chmod %s. Sure you run as sudo\nErr: %s with code %v\n", args[0], args[1], args[2])
	os.Exit(0)
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
	count                  int
	fullLog                bool
)

/* Writting data in target file */
func writeInFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)
	errors(err, func() { err_openfile(filename, err, ERR_OPEN_FILE) })
	defer file.Close()
	_, err = file.WriteString(data)
	errors(err, func() { err_writefile(filename, err, ERR_WRITE_FILE) })
}

func IfSigint(args ...interface{}) {
	err := os.Chmod(args[0].(string), 0777)
	errors(err, func() { err_chmod(args[0], err, ERR_CHMOD) })
	fmt.Println("\nChanged mode for ", args[0])
	os.Exit(0)
}

func IfNoInt() {
	fmt.Print()
}

/* Makes scanning within a minute and write macs in macs.txt for analizing */
func Do(cmd *cobra.Command, arg []string) {
	var strmacs, filename string
	filename = address + ".txt"
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go scanning(ticker, strmacs, filename)
	time.Sleep(time.Duration(scanningtime) * time.Second)
	ticker.Stop()
	fmt.Printf("\nWrote %v macs in %s\n", count, filename)
	err := os.Chmod(address+".txt", 0777)
	errors(err, func() { err_chmod(filename, err, ERR_CHMOD) })
	wifiinterface, address, filename = "", "", ""
	scanningtime, interval = 0, 0
}

func scanning(ticker *time.Ticker, strmacs, filename string) {
	sigchan := make(chan os.Signal)
	fmt.Printf("Passed\t | \tLeft\t | \t Last\t | \t Scanned\n")
	start := time.Now().Unix()
	for sec := range ticker.C {
		macs, err := wifiscan.Scan(wifiinterface)
		errors(err, func() { err_scanning(err, ERR_WIFI_SCAN) })
		for _, mac := range macs {
			go handlesignals.Capture_signals(syscall.SIGINT, sigchan, func() { IfSigint(filename) }, IfNoInt)
			count++
			strmacs += mac.SSID + "\n"
		}
		now := sec.Unix()
		now -= start
		writeInFile(filename, strmacs+"\n")
		out := fmt.Sprintf("%v(sec)\t | %v(sec)\t | %v:%v:%v\t | %v(macs)", now, scanningtime-now, sec.Local().Hour(), sec.Local().Minute(), sec.Local().Second(), len(macs))
		if fullLog {
			fmt.Print(out + "\n")
		} else {
			fmt.Print("\033[2K\r" + out)
		}
		strmacs = ""
	}
}

/* Get flag and show help message if it's not exist */
func parse_flags() {
	rootCmd := &cobra.Command{
		Use:     "macs",
		Version: "1.0",
		Example: `This example scan wifis within 120 seconds 
		every 5 seconds, write scanned macs in Ленина_5-2.txt and show full log` + "\n\n" +
			"\t" + `sudo macs -a Ленина_5-2 -t 120 -s 5 -l`,
		Run: Do,
	}

	rootCmd.Flags().StringVarP(&wifiinterface, "iface", "i", "", "set outgoing interface")
	rootCmd.Flags().StringVarP(&address, "addr", "a", "", "set address")
	rootCmd.Flags().Int64VarP(&scanningtime, "time", "t", 30, "set scanning time")
	rootCmd.Flags().Int64VarP(&interval, "sleep", "s", 13, "set scanning interval")
	rootCmd.Flags().BoolVarP(&fullLog, "log", "l", false, "see full log")
	rootCmd.MarkFlagRequired("addr")
	err := rootCmd.Execute()
	errors(err, func() { err_parseflags(err, ERR_PARSE_FLAGS) })
}

func main() {
	parse_flags()
}
