package cmd

import (
	"fmt"
	"os"

	"github.com/pasknel/lora-packet-forwarder-client/gateway"
	"github.com/pasknel/lorawan-mapper/decode"
	"github.com/pasknel/lorawan-mapper/pcap"
	"github.com/pasknel/lorawan-mapper/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func onNewGateway(gwMac gateway.Mac) error {
	return nil
}

func onDeleteGateway(gwMac gateway.Mac) error {
	return nil
}

func handleRxPackets(client *gateway.Client, capture *os.File) {
	defer capture.Close()

	for {
		select {
		case msg := <-client.RXPacketChan():
			if err := pcap.AddPacket(capture, msg); err != nil {
				log.Error(err)
				continue
			}

			if err := decode.DecodePacket(msg); err != nil {
				log.Error(err)
				continue
			}
		default:
			continue
		}
	}
}

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Capture LoRaWAN messages",
	Long:  `Capture LoRaWAN messages`,

	Run: func(cmd *cobra.Command, args []string) {
		util.PrintBanner()

		logFile, err := InitLogger(LOG_FILE)
		if err != nil {
			log.Fatalf("error creating log file - err: %v", err)
		}
		defer logFile.Close()

		client, err := gateway.NewClient(BIND_ADDRESS, onNewGateway, onDeleteGateway)
		if err != nil {
			log.Fatalf("error creating new gateway client - err: %v", err)
		}
		defer client.Close()

		fmt.Printf("[+] Gateway listener started: %s (UDP) \n", BIND_ADDRESS)

		capture, err := pcap.CreatePCAP(PCAP)
		if err != nil {
			log.Fatalf("error creating pcap file - err: %v", err)
		}

		if err := pcap.CreateNamedPipe(PIPE); err != nil {
			log.Fatal(err)
		}

		if err := pcap.RedirectToPipe(PCAP, PIPE); err != nil {
			log.Fatal(err)
		}

		fmt.Println("[*] Now open the named pipe in Wireshark")

		go handleRxPackets(client, capture)

		<-EXIT
	},
}

func init() {
	captureCmd.Flags().StringVarP(&BIND_ADDRESS, "bind", "b", "0.0.0.0:1700", "Bind address (ip:port)")
	captureCmd.Flags().StringVarP(&LOG_FILE, "log", "l", "gateway.log", "Log File Path")
	captureCmd.Flags().StringVarP(&PCAP, "write", "w", "capture.pcap", "PCAP File Path")
	captureCmd.Flags().StringVarP(&PIPE, "pipe", "p", "/tmp/lorawan", "Pipe Path")
}
