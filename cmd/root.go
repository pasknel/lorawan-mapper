package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/pasknel/lora-packet-forwarder-client/gateway"
	"github.com/pasknel/lorawan-mapper/cmd/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tm "github.com/buger/goterm"
	log "github.com/sirupsen/logrus"
)

var (
	BIND_ADDRESS string
	LOG_FILE     string
	ROWS_LIMIT   = 30
	EXIT         = make(chan struct{})
	SETUP_DB     bool
	PKT_NUMBER   int
)

func EncodeB64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func onNewGateway(gwMac gateway.Mac) error {
	return nil
}

func onDeleteGateway(gwMac gateway.Mac) error {
	return nil
}

func handleRxPackets(client *gateway.Client) {
	tm.Clear()

	PKT_NUMBER = 1
	header := table.Row{"NUMBER", "TIMESTAMP", "GATEWAY MAC", "CHANNEL", "FREQUENCY", "RSSI", "CODE RATE", "MESSAGE TYPE", "MESSAGE INFO"}
	rows := []table.Row{}

	for {
		select {
		case msg := <-client.RXPacketChan():
			log.WithFields(log.Fields{
				"channel":     msg.RXInfo.Channel,
				"frequency":   msg.RXInfo.Frequency,
				"code_rate":   msg.RXInfo.CodeRate,
				"gateway_mac": msg.RXInfo.MAC.String(),
				"payload_b64": EncodeB64(msg.PHYPayload),
			}).Println("received packet with phy payload")

			pkt, err := util.DecodePacket(msg)
			if err != nil {
				log.Errorf("error decoding packet - err: %v", err)
				continue
			}

			now := time.Now()

			rows = append(rows, table.Row{
				PKT_NUMBER,
				now.Format("2006-01-02 15:04:05.000000"),
				msg.RXInfo.MAC.String(),
				msg.RXInfo.Channel,
				msg.RXInfo.Frequency,
				msg.RXInfo.RSSI,
				msg.RXInfo.CodeRate,
				pkt.MessageType,
				pkt.Info,
			})

			PKT_NUMBER++

			tm.MoveCursor(1, 1)

			util.CreateTable(header, rows)
			if len(rows) == ROWS_LIMIT {
				rows = rows[1:]
			}

			tm.Flush()
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "lorawan-mapper",
	Short: "Capture and visualize LoRaWAN messages",
	Long:  `Capture and visualize LoRaWAN messages`,

	Run: func(cmd *cobra.Command, args []string) {
		logFile, err := InitLogger(LOG_FILE)
		if err != nil {
			log.Fatal("error creating log file - err: %v", err)
		}
		defer logFile.Close()

		if SETUP_DB {
			if err := SetupDB(); err != nil {
				log.Fatal(err)
			}
		}

		client, err := gateway.NewClient(BIND_ADDRESS, onNewGateway, onDeleteGateway)
		if err != nil {
			log.Fatalf("error creating new gateway client - err: %v", err)
		}
		defer client.Close()

		fmt.Printf("[*] Starting gateway UDP listener on %s \n", BIND_ADDRESS)

		go handleRxPackets(client)

		<-EXIT
	},
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

func SetupDB() error {
	viper.SetConfigFile("config.yaml")
	viper.ReadInConfig()

	if err := util.SetupNeo4j(); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.Flags().StringVarP(&BIND_ADDRESS, "bind", "b", "0.0.0.0:1700", "Bind address (ip:port)")
	rootCmd.Flags().StringVarP(&LOG_FILE, "log", "l", "gateway.log", "Log File Path")
	rootCmd.Flags().BoolVarP(&SETUP_DB, "db", "d", false, "Setup neo4j database (erases all previous data)")
}
