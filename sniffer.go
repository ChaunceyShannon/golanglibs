//go:build pcap

package golanglibs

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func init() {
	Funcs.Sniffer = sniffer
	Funcs.ReadPcapFile = readPcapFile
}

func doPacketSource(packetSource *gopacket.PacketSource, pkgchan chan *networkPacketStruct, pcapFileHandler ...*pcap.Handle) {
	for packet := range packetSource.Packets() {
		//print("Packet found: ", packet)
		tranSrcPortLayer := packet.TransportLayer()
		if tranSrcPortLayer != nil {
			//print("tranSrcPortLayer found.")
			pkg := networkPacketStruct{}

			linuxSLLLayer := packet.Layer(layers.LayerTypeLinuxSLL)
			if linuxSLLLayer != nil {
				linuxSLLPacket, _ := linuxSLLLayer.(*layers.LinuxSLL)
				pkg.SrcMac = fmt.Sprintf("%s", linuxSLLPacket.Addr)
			}

			//print(packet)
			ethLayer := packet.Layer(layers.LayerTypeEthernet)
			if ethLayer != nil {
				//print("eth layer found")
				ethernetPacket, _ := ethLayer.(*layers.Ethernet)
				pkg.SrcMac = fmt.Sprintf("%s", ethernetPacket.SrcMAC)
				pkg.DstMac = fmt.Sprintf("%s", ethernetPacket.DstMAC)
				//fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
			}

			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			if ipLayer != nil {
				//print("ip layer found")
				ip, ok := ipLayer.(*layers.IPv4)
				if ok {
					pkg.IPVersion = 4
					pkg.SrcIP = fmt.Sprintf("%s", ip.SrcIP)
					pkg.DstIP = fmt.Sprintf("%s", ip.DstIP)
				} else {
					pkg.IPVersion = 6
					ip6, _ := ipLayer.(*layers.IPv6)
					pkg.SrcIP = fmt.Sprintf("%s", ip6.SrcIP)
					pkg.DstIP = fmt.Sprintf("%s", ip6.DstIP)
				}
			}

			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer != nil {
				//print("tcp layer found")
				pkg.Protocol = "tcp"
				tcp, _ := tcpLayer.(*layers.TCP)
				pkg.SrcPort = Int(fmt.Sprintf("%d", tcp.SrcPort))
				pkg.DstPort = Int(fmt.Sprintf("%d", tcp.DstPort))
			}

			udpLayer := packet.Layer(layers.LayerTypeUDP)
			if udpLayer != nil {
				//print("udp layer found")
				pkg.Protocol = "udp"
				udp, _ := udpLayer.(*layers.UDP)
				pkg.SrcPort = Int(fmt.Sprintf("%d", udp.SrcPort))
				pkg.DstPort = Int(fmt.Sprintf("%d", udp.DstPort))
			}

			applicationLayer := packet.TransportLayer()
			if applicationLayer != nil {
				pkg.Data = Str(applicationLayer.LayerPayload())
				//print("Data:", pkg.data)
			}

			// if strStartsWith(pkg.data, "GET /action") {
			// 	print(packet)
			// }

			if pkg.Data != "" {
				pkgchan <- &pkg
			}
		}
	}
	if len(pcapFileHandler) != 0 {
		pcapFileHandler[0].Close()
	}
	close(pkgchan)
}

func sniffer(interfaceName string, filterString string, promisc ...bool) chan *networkPacketStruct {
	// 4096????????????????????????buffer, mtu?????????1500, ??????4096??????????????????, ??????mtu?????????4096, ????????????
	// promisc??????????????????????????????
	// timeout???0.3???, ???kernel???0.3???????????????????????????pcap, ???????????????30???, ??????????????????????????????????????????????????????, 30?????????????????????
	var handle *pcap.Handle
	var err error
	if len(promisc) == 0 {
		handle, err = pcap.OpenLive(interfaceName, 4096, false, getTimeDuration(0.3))
	} else {
		handle, err = pcap.OpenLive(interfaceName, 4096, promisc[0], getTimeDuration(0.3))
	}

	Panicerr(err)
	//defer handle.Close()

	err = handle.SetBPFFilter(filterString)
	Panicerr(err)

	pkgchan := make(chan *networkPacketStruct)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go doPacketSource(packetSource, pkgchan)

	return pkgchan
}

func readPcapFile(pcapFile string) chan *networkPacketStruct {
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}

	pkgchan := make(chan *networkPacketStruct)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go doPacketSource(packetSource, pkgchan, handle)

	return pkgchan
}
