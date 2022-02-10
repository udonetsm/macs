package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	ERR_FILE_CREATE   = 300
	ERR_FILE_WRITE    = 301
	ERR_GET_MACS      = 302
	ERR_FLAGS_EXECUTE = 303
	EXEC_ERR          = 304
	OUT_ERR           = 305
	CONN_ERR          = 306
)

var (
	wifiinterface, address string
	scaningtime            int
)

func main() {
	parse_flags()
}

/*Cathing errors*/
func errors(err error, code int) {
	if err != nil {
		logrus.Info(err)
		os.Exit(code)
	}
}

/*Writes data into a file*/
func write(filename string, data string) {
	file, err := os.OpenFile(address+".txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	errors(err, ERR_FILE_CREATE)
	defer file.Close()
	_, err = file.WriteString(data + "\n")
	errors(err, ERR_FILE_WRITE)
}

/*Disconnect wifi, set encoding code in comand rompt for all be fine*/
func execs() {
	ex := exec.Command("chcp", "437")
	err := ex.Run()
	errors(err, EXEC_ERR)
	ex = exec.Command("netsh", "wlan", "disconnect")
	err = ex.Run()
	errors(err, EXEC_ERR)
}

/*All actions together*/
func Macs(cmd *cobra.Command, arg []string) {
	execs()
	var count, times, counMacs int
	var mac wifiscan.Wifi
	ticker := time.NewTicker(15 * time.Second)
	logrus.Info("Scanning start...")
	go func() {
		for _ = range ticker.C {
			times += 1
			macs, err := wifiscan.Scan(wifiinterface)
			errors(err, ERR_GET_MACS)
			for counMacs, mac = range macs {
				write(address+".txt", mac.SSID)
				count++
				counMacs++
			}
			write(address+".txt", "\n")
			logrus.Infof("Scanned %v times %s and found %v macs", times, wifiinterface, counMacs)
			counMacs = 0
		}
	}()
	time.Sleep(time.Duration(scaningtime) * time.Second)
	ticker.Stop()
	logrus.Infof("Scanning stopped. Wrote %v strings in %s.txt", count, address)
	err := connect()
	errors(err, CONN_ERR)
}

/*parses flags*/
func parse_flags() {
	rootCmd := &cobra.Command{
		Use:     "macs",
		Version: "1.0",
		Example: `   macs.exe -a "Ленина 5" -t 60`,
		Run:     Macs
	}
	rootCmd.Flags().StringVarP(&wifiinterface, "iface", "i", "", "set outgoing interface")
	rootCmd.Flags().StringVarP(&address, "addr", "a", "", "set address")
	rootCmd.Flags().IntVarP(&scaningtime, "time", "t", 120, "set scaning seconds")
	rootCmd.MarkFlagRequired("addr")
	err := rootCmd.Execute()
	errors(err, ERR_FLAGS_EXECUTE)
}

/*Connects to wifi after scaning*/
func connect() error {
	getProfs := exec.Command("netsh", "wlan", "show", "profiles")
	out, err := getProfs.Output()
	errors(err, OUT_ERR)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var profileName []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "All User Profile") {
			profileName = strings.Fields(line)
			break
		}
	}
	connect := exec.Command("netsh", "wlan", "connect", profileName[4])
	err = connect.Run()
	return err
}

/*Parses strings with ":" to yaml format
Read words before ":" and includes it as key
and
Read words after ":" and includes it as value*/
func ParseStringToYamlFormat(text []byte) []byte {
	scanner := bufio.NewScanner(strings.NewReader(string(text)))
	var fields []string
	var yaml_format, key, value string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if strings.Contains(line, ":") {
			fields = strings.Fields(line)
			lenfields := len(fields)
			for field_indx := 0; field_indx < lenfields; field_indx++ {
				if fields[field_indx] == ":" {
					for key_indx := 0; key_indx < field_indx+1; key_indx++ {
						key += fields[key_indx]
					}
					for val_indx := field_indx + 1; val_indx < lenfields; val_indx++ {
						value += fields[val_indx] + " "
					}
				}
			}
		}
		if len(key) > 1 || len(value) > 1 {
			yaml_format += key + ` "` + strings.TrimSpace(value) + `"` + "\n"
			key, value = "", ""
		} else {
			defer func() {
				fmt.Printf("Several strings skipped. It seems like empty. %s %s", key, value)
				fmt.Println()
			}()
		}
	}
	return []byte(yaml_format)
}
