package decode

import (
	"fmt"

	"github.com/brocaar/lorawan"
	"github.com/pasknel/lora-packet-forwarder-client/gateway"
	"github.com/pasknel/lorawan-mapper/util"
)

func UnmarshalPHYPayload(data []byte) (lorawan.PHYPayload, error) {
	var p lorawan.PHYPayload

	err := p.UnmarshalBinary(data)
	if err != nil {
		return p, fmt.Errorf("error during lorawan phypayload unmarshal")
	}

	return p, nil
}

func HandleJoinRequest(phy lorawan.PHYPayload) error {
	joinReq, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return fmt.Errorf("error decoding join request message")
	}

	if err := util.GraphJoinRequest(*joinReq); err != nil {
		return fmt.Errorf("error in neo4j - err: %v", err)
	}

	return nil
}

func HandleDataUp(phy lorawan.PHYPayload) error {
	data, err := phy.MACPayload.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error during mac payload marshal binary - err: %v", err)
	}

	var mac lorawan.MACPayload
	if err := mac.UnmarshalBinary(true, data); err != nil {
		return fmt.Errorf("error during lorawan mac payload unmarshal")
	}

	if err := util.GraphDataUp(mac); err != nil {
		return fmt.Errorf("error in neo4j - err: %v", err)
	}

	return nil
}

func DecodePacket(msg gateway.RXPacketBytes) error {
	phy, err := UnmarshalPHYPayload(msg.PHYPayload)
	if err != nil {
		return fmt.Errorf("error during packet unmarshal - err: %v", err)
	}

	switch phy.MHDR.MType {
	case lorawan.JoinRequest:
		if err := HandleJoinRequest(phy); err != nil {
			return err
		}
	case lorawan.UnconfirmedDataUp, lorawan.ConfirmedDataUp:
		if err := HandleDataUp(phy); err != nil {
			return err
		}
	}

	return nil
}
