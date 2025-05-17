package loratap

const (
	LORA_HEADER_LENGTH = uint16(35)
	DLT_LORATAP        = uint32(270)
)

type SF uint8

const (
	SF5  SF = 5
	SF6  SF = 6
	SF7  SF = 7
	SF8  SF = 8
	SF9  SF = 9
	SF10 SF = 10
	SF11 SF = 11
	SF12 SF = 12
)

type CR uint8

const (
	CRNone CR = 0
	CR4_5  CR = 5
	CR4_6  CR = 6
	CR4_7  CR = 7
	CR4_8  CR = 8
)

type LoRaTapChannel struct {
	Frequency uint32
	Bandwidth uint8
	SF        SF
}

type LoRaTapRSSI struct {
	PacketRSSI  uint8
	MaxRSSI     uint8
	CurrentRSSI uint8
	SNR         uint8
}

type LoRaTapFlags struct {
	ModFSK      uint8
	IQInverted  uint8
	ImplicitHdr uint8
	CRCOk       uint8
	CRCInvalid  uint8
	NoCRC       uint8
	Padding     uint8
}

type LoRaTapHeader struct {
	Version   uint8
	Padding   uint8
	Length    uint16
	Channel   LoRaTapChannel
	RSSI      LoRaTapRSSI
	SyncWord  uint8
	SourceGW  uint64
	Timestamp uint32
	Flags     uint8
	CR        CR
	DataRate  uint16
	IfChannel uint8
	RFChain   uint8
	Tag       uint16
}

func (f LoRaTapFlags) Pack() uint8 {
	var flags uint8 = 0
	flags |= (f.ModFSK & 1) << 0
	flags |= (f.IQInverted & 1) << 1
	flags |= (f.ImplicitHdr & 1) << 2
	flags |= (f.CRCOk & 1) << 3
	flags |= (f.CRCInvalid & 1) << 4
	flags |= (f.NoCRC & 1) << 5
	flags |= (f.Padding & 1) << 6
	return flags
}

func UnpackLoRaTapFlags(flags uint8) LoRaTapFlags {
	return LoRaTapFlags{
		ModFSK:      (flags >> 0) & 1,
		IQInverted:  (flags >> 1) & 1,
		ImplicitHdr: (flags >> 2) & 1,
		CRCOk:       (flags >> 3) & 1,
		CRCInvalid:  (flags >> 4) & 1,
		NoCRC:       (flags >> 5) & 1,
		Padding:     (flags >> 6) & 1,
	}
}
