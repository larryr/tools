package main

// a test client for testing the dhcp server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
)

// program flags
var (
	discover = flag.Bool("discover", false, "Enable discover mode")
	inform   = flag.Bool("inform", false, "Enable inform mode")
	//verbose  = flag.Bool("verbose", false, "Enable verbose mode")
)

func main() {
	flag.Parse()

	args := os.Args
	if len(args) != 2 {
		log.Fatalf("expected interface argument: dhcptest <interface-string>")
	}
	intf := args[1]

	if *discover {
		// send dhcp4 request
		err := discover4(intf)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *inform {
		// send dhcp4 inform
		err := inform4(intf)
		if err != nil {

		}
	}
}

// send a DHCP4 Discover out the indicated named interface
func discover4(ifname string) error {

	var err error

	// create client
	cli, err := nclient4.New(ifname)

	// Discover
	offer, err := cli.DiscoverOffer(nil)

	log.Printf("offer:\n%v\n", offer)

	return err
}

func inform4(ifname string) error {

	var err error

	// get address assigned to given interface
	curIP, err := getInterfaceIP(ifname)
	if err != nil {
		return err
	}

	// create client
	cli, err := nclient4.New(ifname)
	if err != nil {
		return err
	}

	// Inform
	ack, err := cli.Inform(nil)

	log.Printf("inform ack:\n%v\n", ack)

	return err

}

func getInterfaceIP(intf string) (net.IP, error) {
	interfaceAddr, err := net.InterfaceByName(intf)
	if err != nil {
		return nil, err
	}

	addrs, err := interfaceAddr.Addrs()
	if err != nil {
		return nil, err
	}

	// return the first IP address found
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if len(ip) == 0 {
			continue
		}
		return ip, nil
	}

	return nil, fmt.Errorf("no IP address found for interface %s", intf)
}

// partialExchange will not perform the full exchange.  The DHCPOFFER will not be requested.
// a DHCPDISCOVER will be sent.
// the set of DHCPOFFERS will be collected and returned.
// no DHCPREQUEST will be sent. and therefore the server will not hold a lease.
// the 'exchange' will be returned to the caller.

/*
func partialExchange(ifname string, modifiers ...dhcpv4.Modifier) ([]*dhcpv4.DHCPv4, error) {

	client := client4.NewClient()
	exchange := make([]*dhcpv4.DHCPv4, 0)
	raddr, err := client.getRemoteUDPAddr()
	if err != nil {
		return nil, err
	}
	laddr, err := c.getLocalUDPAddr()
	if err != nil {
		return nil, err
	}
	// Get our file descriptor for the raw socket we need.
	var sfd int
	// If the address is not net.IPV4bcast, use a unicast socket. This should
	// cover the majority of use cases, but we're essentially ignoring the fact
	// that the IP could be the broadcast address of a specific subnet.
	if raddr.IP.Equal(net.IPv4bcast) {
		sfd, err = MakeBroadcastSocket(ifname)
	} else {
		//sfd, err = makeRawSocket(ifname)
		log.Fatalf("Remote address is not the broadcast address: %v ", raddr)
	}
	if err != nil {
		return conversation, err
	}
	rfd, err := makeListeningSocketWithCustomPort(ifname, laddr.Port)
	if err != nil {
		return conversation, err
	}

	defer func() {
		// close the sockets
		if err := unix.Close(sfd); err != nil {
			log.Printf("unix.Close(sendFd) failed: %v", err)
		}
		if sfd != rfd {
			if err := unix.Close(rfd); err != nil {
				log.Printf("unix.Close(recvFd) failed: %v", err)
			}
		}
	}()

	// Discover
	discover, err := dhcpv4.NewDiscoveryForInterface(ifname, modifiers...)
	if err != nil {
		return conversation, err
	}
	conversation = append(conversation, discover)

	// Offer
	offer, err := c.SendReceive(sfd, rfd, discover, dhcpv4.MessageTypeOffer)
	if err != nil {
		return conversation, err
	}
	conversation = append(conversation, offer)

	return conversation, nil
}
*/
