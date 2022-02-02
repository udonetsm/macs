package main

import (
	"log"
	"os"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/spf13/cobra"
)

/* For parse flag */
var wifiinterface string

func main() {
	parse_flags()
}

/* Capture errors */
func errors(err error, code int) {
	if err != nil {
		log.Println(err)
		os.Exit(code)
	}
}

/* Creates new file even if it exists */
func create(filename string) {
	file, err := os.Create(filename)
	errors(err, 12)
	defer file.Close()
}

/* Writting data in target file */
func writeInFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0777)
	errors(err, 2)
	defer file.Close()
	_, err = file.WriteString(data)
	errors(err, 3)
}

/* Makes scanning within a minute and write macs in macs.txt for analizing */
func Macs(cmd *cobra.Command, arg []string) {
	log.Println("Scanning...")
	ticker := time.NewTicker(1 * time.Second)
	create("macs.txt") //it necessary for rewrite file always
	go func() {
		for _ = range ticker.C {
			macs, err := wifiscan.Scan(wifiinterface)
			errors(err, 4)
			for _, mac := range macs {
				writeInFile("macs.txt", mac.SSID+"\n")
			}
		}
	}()
	time.Sleep(60 * time.Second)
	ticker.Stop()
	log.Println("Scanning finished, check macs.txt")
}

/* Get network interface name as flag and show help message if it's not exist */
func parse_flags() {
	rootCmd := &cobra.Command{
		Use:     "macs",
		Version: "1.0",
		Example: "\tsudo macs -i wlp3s0",
		Run:     Macs,
	}

	rootCmd.Flags().StringVarP(&wifiinterface, "iface", "i", "", "mark outgoing interface")
	rootCmd.MarkFlagRequired("iface")

	err := rootCmd.Execute()
	errors(err, 5)
}
