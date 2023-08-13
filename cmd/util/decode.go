package util

import (
	"encoding/json"
	"fmt"

	"github.com/brocaar/lorawan"
	"github.com/fatih/color"
	"github.com/pasknel/lora-packet-forwarder-client/gateway"
	"github.com/tidwall/pretty"

	log "github.com/sirupsen/logrus"
)

type Packet struct {
	MessageType string
	Info        string
}

func prettyPrint(data interface{}) {
	b, _ := json.Marshal(data)
	result := pretty.Pretty(b)
	result = pretty.Color(result, nil)
	fmt.Println(string(result))
}

func UnmarshalPHYPayload(data []byte) (lorawan.PHYPayload, error) {
	var p lorawan.PHYPayload

	err := p.UnmarshalBinary(data)
	if err != nil {
		return p, fmt.Errorf("error during lorawan phypayload unmarshal")
	}

	return p, nil
}

func HandleJoinRequest(phy lorawan.PHYPayload) (Packet, error) {
	var pkt Packet

	joinReq, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		return pkt, fmt.Errorf("error decoding join request message")
	}

	log.WithFields(log.Fields{
		"mhdr":        phy.MHDR,
		"mac_payload": phy.MACPayload,
		"dev_eui":     joinReq.DevEUI,
		"dev_nonce":   joinReq.DevNonce,
		"join_eui":    joinReq.JoinEUI,
	}).Info("join request")

	c := color.New(color.FgRed)
	pkt.MessageType = c.Sprint("Join Request")

	pkt.Info = fmt.Sprintf("dev_eui: %s | join_eui: %s | dev_nonce: %x",
		joinReq.DevEUI.String(),
		joinReq.JoinEUI,
		joinReq.DevNonce,
	)

	GraphJoinRequest(*joinReq)

	return pkt, nil
}

func HandleUnconfirmedDataUp(phy lorawan.PHYPayload) (Packet, error) {
	var pkt Packet

	data, err := phy.MACPayload.MarshalBinary()
	if err != nil {
		return pkt, fmt.Errorf("error during mac payload marshal binary - err: %v", err)
	}

	var mac lorawan.MACPayload
	if err := mac.UnmarshalBinary(true, data); err != nil {
		return pkt, fmt.Errorf("error during lorawan mac payload unmarshal")
	}

	log.WithFields(log.Fields{
		"mhdr":        phy.MHDR,
		"mac_payload": phy.MACPayload,
		"dev_addr":    mac.FHDR.DevAddr,
		"nwk_id":      mac.FHDR.DevAddr.NwkID(),
		"fport":       mac.FPort,
		"fcnt":        mac.FHDR.FCnt,
	}).Info("unconfirmed data up")

	var dataBytes []byte
	for _, payload := range mac.FRMPayload {
		data, err := payload.MarshalBinary()
		if err != nil {
			log.Errorf("error during frm payload marshal binary - err: %v", err)
			continue
		}
		dataBytes = append(dataBytes, data...)
	}

	c := color.New(color.FgGreen)
	pkt.MessageType = c.Sprint("Unconfirmed Data Up")

	pkt.Info = fmt.Sprintf("dev_addr: %s | nwk_id: %x | fport: %03d | fcnt: %d | len: %d",
		mac.FHDR.DevAddr.String(),
		mac.FHDR.DevAddr.NwkID(),
		*mac.FPort,
		mac.FHDR.FCnt,
		len(dataBytes),
	)

	GraphUnconfirmedDataUp(mac)

	return pkt, nil
}

func HandleConfirmedDataUp(phy lorawan.PHYPayload) (Packet, error) {
	var pkt Packet

	data, err := phy.MACPayload.MarshalBinary()
	if err != nil {
		return pkt, fmt.Errorf("error during mac payload marshal binary - err: %v", err)
	}

	var mac lorawan.MACPayload
	if err := mac.UnmarshalBinary(true, data); err != nil {
		return pkt, fmt.Errorf("error during lorawan mac payload unmarshal")
	}

	log.WithFields(log.Fields{
		"mhdr":        phy.MHDR,
		"mac_payload": phy.MACPayload,
		"dev_addr":    mac.FHDR.DevAddr,
		"nwk_id":      mac.FHDR.DevAddr.NwkID(),
		"fport":       mac.FPort,
		"fcnt":        mac.FHDR.FCnt,
	}).Info("confirmed data up")

	var dataBytes []byte
	for _, payload := range mac.FRMPayload {
		data, err := payload.MarshalBinary()
		if err != nil {
			log.Errorf("error during frm payload marshal binary - err: %v", err)
			continue
		}
		dataBytes = append(dataBytes, data...)
	}

	c := color.New(color.FgBlue)
	pkt.MessageType = c.Sprint("Confirmed Data Up")

	pkt.Info = fmt.Sprintf("dev_addr: %s | nwk_id: %x | fport: %03d | fcnt: %d | len: %d",
		mac.FHDR.DevAddr.String(),
		mac.FHDR.DevAddr.NwkID(),
		*mac.FPort,
		mac.FHDR.FCnt,
		len(dataBytes),
	)

	return pkt, nil
}

func DecodePacket(msg gateway.RXPacketBytes) (Packet, error) {
	var pkt Packet

	phy, err := UnmarshalPHYPayload(msg.PHYPayload)
	if err != nil {
		return pkt, fmt.Errorf("error during packet unmarshal - err: %v", err)
	}

	switch phy.MHDR.MType {
	case lorawan.JoinRequest:
		pkt, err = HandleJoinRequest(phy)
		if err != nil {
			return pkt, fmt.Errorf("error decoding join request")
		}
	case lorawan.UnconfirmedDataUp:
		pkt, err = HandleUnconfirmedDataUp(phy)
		if err != nil {
			return pkt, fmt.Errorf("error decoding unconfirmed data up")
		}
	case lorawan.ConfirmedDataUp:
		pkt, err = HandleConfirmedDataUp(phy)
		if err != nil {
			return pkt, fmt.Errorf("error decoding confirmed data up")
		}
	default:
		c := color.New(color.FgYellow)
		pkt.MessageType = c.Sprint(phy.MHDR.MType.String())
		log.Printf("received packet - mtype: %v", phy.MHDR.MType)
	}

	return pkt, nil
}
