package cmd

import (
	"os"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var (
	BIND_ADDRESS string
	LOG_FILE     string
	PCAP         string
	PIPE         string
	ROWS_LIMIT   = 30
	EXIT         = make(chan struct{})
	PKT_NUMBER   int
)

var rootCmd = &cobra.Command{
	Use:   "lorawan-mapper",
	Short: "Capture and visualize LoRaWAN messages",
	Long:  `Capture and visualize LoRaWAN messages`,

	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func InitLogger(logFile string) (*os.File, error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return file, err
	}

	log.SetOutput(file)
	log.SetFormatter(&log.JSONFormatter{})

	return file, nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(captureCmd)
}
