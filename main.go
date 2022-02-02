package main

import (
	"log"
	"os"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/spf13/cobra"
)

var wifiinterface string

func main() {
	parse_flags()
}

func errors(err error, code int) {
	if err != nil {
		log.Println(err)
		os.Exit(code)
	}
}

func writeInFile(filename string, data string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	errors(err, 2)
	defer file.Close()
	_, err = file.WriteString(data)
	errors(err, 3)
}

func Macs(cmd *cobra.Command, arg []string) {
	log.Println("Scanning...")
	ticker := time.NewTicker(1 * time.Second)
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
