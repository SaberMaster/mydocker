package network

import (
	"encoding/json"
	"fmt"
	"github.com/3i2bgod/mydocker/container"
	"github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"device"`
	IPAddress   net.IP           `json:"ip_address"`
	MacAddress  net.HardwareAddr `json:"mac_address"`
	PortMapping []string         `json:"port_mapping"`
	Network     *Network
}

type NetworkDriver interface {
	Name() string

	Create(subnet string, name string) (*Network, error)

	Delete(network Network) error

	Connect(network *Network, endpoint *Endpoint) error

	Disconnect(network *Network, endpoint *Endpoint) error
}

func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); nil != err {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)

	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer nwFile.Close()
	if nil != err {
		logrus.Errorf("network error:", err)
		return err
	}

	ipamConfigJson, err := json.Marshal(nw)
	if nil != err {
		logrus.Errorf("network error:", err)
		return err
	}

	_, err = nwFile.Write(ipamConfigJson)
	if nil != err {
		logrus.Errorf("network error:", err)
		return err
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); nil != err {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if nil != err {
		return err
	}

	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if nil != err {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if nil != err {
		logrus.Errorf("Error load network info, %v", err)
		return err
	}
	return nil
}

var (
	defaultNetworkPath = "/var/run/mydocker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); nil != err {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		network := &Network{
			Name: nwName,
		}

		if err := network.load(nwPath); nil != err {
			logrus.Errorf("error load network: %s", err)
		}

		networks[nwName] = network
		return nil
	})

	return nil
}

func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)

	ip, err := ipAllocator.Allocate(cidr)
	if nil != err {
		return err
	}
	cidr.IP = ip

	network, err := drivers[driver].Create(cidr.String(), name)

	if nil != err {
		return err
	}

	return network.dump(defaultNetworkPath)
}

func ListNetwork() {
	writer := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(writer, "Name\tIpRange\tDriver\n")

	for _, network := range networks {
		fmt.Fprintf(writer, "%s\t%s\t%s\n",
			network.Name,
			network.IpRange.String(),
			network.Driver,
		)
	}

	if err := writer.Flush(); nil != err {
		logrus.Errorf("Flush error: %v", err)
	}
}

func DeleteNetwork(networkName string) error {
	network, ok := networks[networkName]

	if !ok {
		return fmt.Errorf("No such network %s", networkName)
	}

	if err := ipAllocator.Release(network.IpRange, &network.IpRange.IP); nil != err {
		return fmt.Errorf("Remove gateway ip error: %s", err)
	}

	if err := drivers[network.Driver].Delete(*network); nil != err {
		return fmt.Errorf("Remove network driver error: %s", err)
	}
	return network.remove(defaultNetworkPath)
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	network, ok := networks[networkName]

	if !ok {
		return fmt.Errorf("No such network %s", networkName)
	}

	ip, err := ipAllocator.Allocate(network.IpRange)
	if nil != err {
		return err
	}

	endpoint := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	if err := drivers[network.Driver].Connect(network, endpoint); nil != err {
		return err
	}

	// config dev ip and route in the netns
	if err = configEndpointIpAddressAndRoute(endpoint, cinfo); nil != err {
		return err
	}

	return configPortMapping(endpoint, cinfo)
}

func configPortMapping(endpoint *Endpoint, cInfo *container.ContainerInfo) error {
	for _, pm := range endpoint.PortMapping {
		portMapping := strings.Split(pm, ":")

		if 2 != len(portMapping) {
			logrus.Errorf("port mapping format error. %v", pm)
			continue
		}

		// iptables -t nat -A PREROUTING -p tcp -m tcp --dport host_port -j DNAT --to-destination container_ip:container_port
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], endpoint.IPAddress.String(), portMapping[1])
		logrus.Infof("iptables cmd:%s", iptablesCmd)

		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if nil != err {
			logrus.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}

func configEndpointIpAddressAndRoute(endpoint *Endpoint, cInfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(endpoint.Device.PeerName)

	if nil != err {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	// add peer virtual endpoint to container namespace
	defer enterContainerNetns(&peerLink, cInfo)()

	interfaceIP := *endpoint.Network.IpRange
	interfaceIP.IP = endpoint.IPAddress.To4()

	// set veth ip in the container
	if err = setInterfaceIP(endpoint.Device.PeerName, interfaceIP.String()); nil != err {
		return fmt.Errorf("set up interface:%v ip error: %s", endpoint.Network, err)
	}

	// start veth in the container
	if err = setInterfaceUp(endpoint.Device.PeerName); nil != err {
		return err
	}

	// enable `lo` 127.0.0.1 interface
	if err = setInterfaceUp("lo"); nil != err {
		return err
	}

	// route add -net 0.0.0.0/0 gw {bridge} dev {veth}
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        endpoint.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); nil != err {
		return err
	}
	return nil
}

func enterContainerNetns(link *netlink.Link, cInfo *container.ContainerInfo) func() {

	file, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cInfo.Pid), os.O_RDONLY, 0)
	if nil != err {
		logrus.Errorf("get container net namespace error:%v.", err)
	}

	nsFD := file.Fd()

	//lock current thread
	//otherwise, goroutine will be dispatch to another thread

	runtime.LockOSThread()

	// move link to container's network namespace
	// ip link set $link netns $ns
	if err = netlink.LinkSetNsFd(*link, int(nsFD)); nil != err {
		logrus.Errorf("set link netns error: %v", err)
	}

	// get net namespace of current network
	origins, err := netns.Get()

	if nil != err {
		logrus.Errorf("get current netns error: %v", err)
	}

	if err = netns.Set(netns.NsHandle(nsFD)); nil != err {
		logrus.Errorf("set netns error: %v", err)
	}

	return func() {
		// reset to origin net namespace
		netns.Set(origins)
		// close origin namespace file
		origins.Close()
		runtime.UnlockOSThread()
		// close namespace file
		file.Close()
	}
}

func Disconnect(networkName string, cinfo *container.ContainerInfo) error {
	panic("not implement")
}
