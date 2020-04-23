package network

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

type BridgeNetworkDriver struct {
	NetworkDriver
}

func (driver *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name

	bridge, err := netlink.LinkByName(bridgeName)

	if nil != err {
		return err
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = endpoint.ID[:5]
	// bind another peer to the bridge
	linkAttrs.MasterIndex = bridge.Attrs().Index

	// peerName is "cif-{endpoint.Id[0:5]}"
	endpoint.Device = netlink.Veth{
		LinkAttrs: linkAttrs,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	if err = netlink.LinkAdd(&endpoint.Device); nil != err {
		return fmt.Errorf("Error Add Endpoint Device: %v", err)
	}

	// ip link set xxx up
	if err = netlink.LinkSetUp(&endpoint.Device); nil != err {
		return fmt.Errorf("Error set Endpoint Device up: %v", err)
	}
	return nil
}

func (driver *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (driver *BridgeNetworkDriver) initBridge(network *Network) error {
	// create bridge virtual dev
	// brctl addbr br0
	bridgeName := network.Name
	if err := createBridgeInterface(bridgeName); nil != err {
		return fmt.Errorf("Error add bridge: %s, Error: %v", bridgeName, err)
	}

	// set bridge virtual dev addr and route
	// route add -net 172.18.0.0/24 dev br0
	gatewayIp := *network.IpRange
	gatewayIp.IP = network.IpRange.IP

	if err := setInterfaceIP(bridgeName, gatewayIp.String()); nil != err {
		return fmt.Errorf("Error assigning address: %s on bridge: %s with an error of: %v", gatewayIp, bridgeName, err)
	}

	// start bridge virtual dev
	// ip link set br0 up
	if err := setInterfaceUp(bridgeName); nil != err {
		return fmt.Errorf("Error set bridge up: %s, Error: %v", gatewayIp, bridgeName, err)
	}

	// set iptables snat rule
	// iptables -t nat -A POSTROUTING -s <bridgeName> ! -o <bridgeName> -j MASQUERADE
	if err := setupIPTables(bridgeName, network.IpRange); nil != err {
		return fmt.Errorf("Error setting iptables for: %s, Error: %v", bridgeName, err)
	}

	return nil
}


func (driver *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip

	network := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver: driver.Name(),
	}

	err := driver.initBridge(network)
	if nil != err {
		logrus.Errorf("error init bridge: %v", err)
	}
	return network, err
}

func (bridge *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	iface, err := netlink.LinkByName(bridgeName)
	if nil != err {
		return err
	}
	// todo: remove snat iptables
	return netlink.LinkDel(iface)
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// iptables -t nat -A POSTROUTING -s <subnet> ! -o <bridgeName> -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	logrus.Infof("iptables cmd:%s", iptablesCmd)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if nil != err {
		logrus.Errorf("iptables Output, %v", output)
	}
	return nil
}

func setInterfaceUp(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if nil != err {
		return fmt.Errorf("Error retrieving a link named: %s, error: %v", interfaceName, err)
	}

	if err := netlink.LinkSetUp(iface); nil != err {
		return fmt.Errorf("Error enabling interface for: %s, error: %v", interfaceName, err)
	}
	return nil
}

//  setInterfaceIP("bridgeName", "192.168.0.1/24")
func setInterfaceIP(name string, rawIP string) error {
	iface, err := netlink.LinkByName(name)
	if nil != err {
		return fmt.Errorf("error get interface: %v", err)
	}

	ipNet, err := netlink.ParseIPNet(rawIP)

	if nil != err {
		return err
	}

	addr := &netlink.Addr{
		IPNet:       ipNet,
		Label:       "",
		Flags:       0,
		Scope:       0,
		Peer:        nil,
	}
	//ip addr add $addr dev $link
	// and will add route 192.168.0.0/24 to this interface
	return netlink.AddrAdd(iface, addr)
}

func createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)

	if nil == err || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = bridgeName

	br := &netlink.Bridge{
		LinkAttrs:         linkAttrs,
	}
	if err := netlink.LinkAdd(br); nil != err {
		return fmt.Errorf("Bridge creation failed for bridge: %s, %v", bridgeName, err)
	}
	return nil
}
