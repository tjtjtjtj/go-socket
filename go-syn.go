package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// get the local ip and port based on our destination ip
func localIPPort(dstip net.IP) (net.IP, int) {
	serverAddr, err := net.ResolveUDPAddr("udp", dstip.String()+":12345")
	if err != nil {
		log.Fatal(err)
	}

	// We don't actually connect to anything, but we can determine
	// based on our destination ip what source ip we should use.
	if con, err := net.DialUDP("udp", nil, serverAddr); err == nil {
		if udpaddr, ok := con.LocalAddr().(*net.UDPAddr); ok {
			return udpaddr.IP, udpaddr.Port
		}
	}
	log.Fatal("could not get local ip: " + err.Error())
	return nil, -1
}

func main() {
	if len(os.Args) != 3 {
		log.Printf("Usage: %s <host/ip> <port>\n", os.Args[0])
		os.Exit(-1)
	}
	log.Println("starting")

	dstaddrs, err := net.LookupIP(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// parse the destination host and port from the command line os.Args
	dstip := dstaddrs[0].To4()
	var dstport layers.TCPPort
	if d, err := strconv.ParseUint(os.Args[2], 10, 16); err != nil {
		log.Fatal(err)
	} else {
		dstport = layers.TCPPort(d)
	}

	srcip, sport := localIPPort(dstip)
	srcport := layers.TCPPort(sport)
	log.Printf("using srcip: %v", srcip.String())

	// Our IP header... not used, but necessary for TCP checksumming.
	ip := &layers.IPv4{
		SrcIP:    srcip,
		DstIP:    dstip,
		Protocol: layers.IPProtocolTCP,
	}
	// Our TCP header
	tcp := &layers.TCP{
		SrcPort: srcport,
		DstPort: dstport,
		Seq:     1105024978,
		SYN:     true,
		Window:  14600,
	}
	tcp.SetNetworkLayerForChecksum(ip)

	// Serialize.  Note:  we only serialize the TCP layer, because the
	// socket we get with net.ListenPacket wraps our data in IPv4 packets
	// already.  We do still need the IP layer to compute checksums
	// correctly, though.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, tcp); err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("writing request")
	if _, err := conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstip}); err != nil {
		log.Fatal(err)
	}

	// Set deadline so we don't wait forever.
	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Fatal(err)
	}

	for {
		b := make([]byte, 4096)
		log.Println("reading from conn")
		n, addr, err := conn.ReadFrom(b)
		if err != nil {
			log.Println("error reading packet: ", err)
			return
		} else if addr.String() == dstip.String() {
			// Decode a packet
			log.Printf("addr:%s", addr.String())
			packet := gopacket.NewPacket(b[:n], layers.LayerTypeTCP, gopacket.Default)
			packet2 := gopacket.NewPacket(b[:n], layers.LayerTypeIPv4, gopacket.Default)
			log.Printf("packet:%s", packet.Dump())
			log.Printf("packet:%s", packet2.Dump())
			// Get the TCP layer from this packet
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)

				if tcp.DstPort == srcport {
					if tcp.SYN && tcp.ACK {
						log.Printf("Port %d is OPEN\n", dstport)
					} else {
						log.Printf("Port %d is CLOSED\n", dstport)
					}
					return
				}
			}
		} else {
			//packet := gopacket.NewPacket(b[:n], layers.LayerTypeTCP, gopacket.Default)
			packet := gopacket.NewPacket(b[:n], layers.LayerTypeIPv4, gopacket.Default)
			log.Printf("Got packet not matching addr\n")
			log.Printf("packet:%s", packet.Dump())

		}
	}
}
