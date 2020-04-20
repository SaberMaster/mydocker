package network

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net"
	"github.com/vishvananda/netlink"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
)
type Network struct {
	Name string
	IpRange *net.IPNet
	Driver string
}

type Endpoint struct {
	ID string `json:"id"`
	Device netlink.Veth `json:"device"`
	IPAddress net.IP `json:"ip_address"`
	MacAddress net.HardwareAddr `json:"mac_address"`
	PortMapping []string `json:"port_mapping"`
	Network *Network
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
	drivers = map[string]NetworkDriver{}
	networks = map[string]*Network{}
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

func ListNetwork()  {
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

