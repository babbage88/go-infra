package node_networking

import (
	"encoding/binary"
	"net"
)

// Ethernet frame
type ethernetFrame struct {
	DstMAC  net.HardwareAddr
	SrcMAC  net.HardwareAddr
	EthType uint16
	Payload []byte
}

func (f *ethernetFrame) Marshal() []byte {
	buf := make([]byte, 14+len(f.Payload))
	copy(buf[0:6], f.DstMAC)
	copy(buf[6:12], f.SrcMAC)
	binary.BigEndian.PutUint16(buf[12:14], f.EthType)
	copy(buf[14:], f.Payload)
	return buf
}

// ARP packet (IPv4 only)
type arpPacket struct {
	HType  uint16
	PType  uint16
	HLen   uint8
	PLen   uint8
	OpCode uint16
	SrcMAC net.HardwareAddr
	SrcIP  net.IP
	DstMAC net.HardwareAddr
	DstIP  net.IP
}

func (a *arpPacket) Marshal() []byte {
	buf := make([]byte, 28)
	binary.BigEndian.PutUint16(buf[0:2], a.HType)
	binary.BigEndian.PutUint16(buf[2:4], a.PType)
	buf[4] = a.HLen
	buf[5] = a.PLen
	binary.BigEndian.PutUint16(buf[6:8], a.OpCode)
	copy(buf[8:14], a.SrcMAC)
	copy(buf[14:18], a.SrcIP.To4())
	copy(buf[18:24], a.DstMAC)
	copy(buf[24:28], a.DstIP.To4())
	return buf
}
