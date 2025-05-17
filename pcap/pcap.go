package pcap

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pasknel/lora-packet-forwarder-client/gateway"
	"github.com/pasknel/lorawan-mapper/loratap"
)

var (
	LORA_HEADER_LENGTH = uint16(35)
	DLT_LORATAP        = uint32(270)
)

type PcapHeader struct {
	MagicNumber  uint32
	VersionMajor uint16
	VersionMinor uint16
	ThisZone     int32
	SigFigs      uint32
	SnapLen      uint32
	Network      uint32
}

type FrameHeader struct {
	TsSec   uint32
	TsUsec  uint32
	InclLen uint32
	OrigLen uint32
}

func HexToUint64(hexStr string) uint64 {
	num := new(big.Int)
	num.SetString(hexStr, 16)
	return num.Uint64()
}

func ConvertDataRate(dr string) loratap.CR {
	switch dr {
	case "4/5":
		return loratap.CR4_5
	case "4/6":
		return loratap.CR4_6
	case "4/7":
		return loratap.CR4_7
	case "4/8":
		return loratap.CR4_8
	default:
		return loratap.CRNone
	}
}

func CreatePCAP(path string) (*os.File, error) {
	pcap, err := os.Create(path)
	if err != nil {
		return pcap, fmt.Errorf("error creating pcap file - err: %v", err)
	}

	header := PcapHeader{
		MagicNumber:  0xa1b2c3d4,
		VersionMajor: 2,
		VersionMinor: 4,
		ThisZone:     0,
		SigFigs:      0,
		SnapLen:      255,
		Network:      DLT_LORATAP,
	}

	binary.Write(pcap, binary.LittleEndian, header)

	fmt.Printf("[+] Pcap created: %s \n", path)

	return pcap, nil
}

func CreateNamedPipe(path string) error {
	os.Remove(path)

	err := syscall.Mkfifo(path, 0666)
	if err != nil {
		return fmt.Errorf("error creating named pipe: %s", err)
	}

	fmt.Printf("[+] Named pipe created: %s \n", path)

	return nil
}

func RedirectToPipe(pcap string, pipe string) error {
	args := []string{"-c", fmt.Sprintf("tail -f %s > %s", pcap, pipe)}

	cmd := exec.Command("sh", args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error running bash command - err: %v", err)
	}

	fmt.Println("[+] Redirecting pcap to named pipe")

	return nil
}

func AddPacket(pcap *os.File, msg gateway.RXPacketBytes) error {
	bandwidth := msg.RXInfo.DataRate.Bandwidth
	gateway := msg.RXInfo.MAC.String()
	spreading_factor := msg.RXInfo.DataRate.SpreadFactor
	channel := msg.RXInfo.Channel
	freq := msg.RXInfo.Frequency
	rssi := msg.RXInfo.RSSI
	code_rate := ConvertDataRate(msg.RXInfo.CodeRate)
	timestamp := time.Now()

	lora_tap_channel := loratap.LoRaTapChannel{
		Frequency: uint32(freq),
		Bandwidth: uint8(bandwidth / 125),
		SF:        loratap.SF(spreading_factor),
	}

	lora_tap_rssi := loratap.LoRaTapRSSI{
		PacketRSSI:  uint8(255),
		MaxRSSI:     uint8(255),
		CurrentRSSI: uint8(rssi),
		SNR:         uint8(0),
	}

	lora_tap_flags := loratap.LoRaTapFlags{
		ModFSK:      0,
		IQInverted:  0,
		ImplicitHdr: 0,
		CRCOk:       1,
		CRCInvalid:  0,
		NoCRC:       0,
		Padding:     2,
	}

	lora_tap_header := loratap.LoRaTapHeader{
		Version:   1,
		Padding:   0,
		Length:    loratap.LORA_HEADER_LENGTH,
		Channel:   lora_tap_channel,
		RSSI:      lora_tap_rssi,
		SyncWord:  0x34,
		SourceGW:  HexToUint64(gateway),
		Timestamp: uint32(timestamp.Unix()),
		Flags:     lora_tap_flags.Pack(),
		CR:        code_rate,
		DataRate:  0,
		IfChannel: uint8(channel),
		RFChain:   0,
		Tag:       0,
	}

	payload := msg.PHYPayload
	size := len(payload)

	frameHeader := FrameHeader{
		TsSec:   uint32(timestamp.Unix() + 1),
		TsUsec:  uint32(timestamp.Nanosecond() / 1000),
		InclLen: uint32(uint16(size) + loratap.LORA_HEADER_LENGTH),
		OrigLen: uint32(uint16(size) + loratap.LORA_HEADER_LENGTH),
	}

	binary.Write(pcap, binary.LittleEndian, frameHeader)
	binary.Write(pcap, binary.BigEndian, lora_tap_header)
	binary.Write(pcap, binary.LittleEndian, payload)

	return nil
}
