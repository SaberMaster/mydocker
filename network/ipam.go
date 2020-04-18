package network

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets *map[string]string
}


var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err :=os.Stat(ipam.SubnetAllocatorPath); nil != err {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()

	if nil != err {
		return err
	}
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if nil != err {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if nil != err {
		logrus.Errorf("Error dump allocation info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); nil != err {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}

	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if nil != err {
		return err
	}

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if nil != err {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if nil != err {
		return err
	}
	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ipam.Subnets = &map[string]string{}
	err = ipam.load()
	if nil != err {
		logrus.Errorf("Error load allocation info, %v", err)
	}
	ones, size := subnet.Mask.Size()

	if _, exists := (*ipam.Subnets)[subnet.String()]; !exists {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1 << uint((size - ones)))
	}

	for offset := range((*ipam.Subnets)[subnet.String()]) {
		if '0' == (*ipam.Subnets)[subnet.String()][offset] {
			setValue(ipam, subnet, offset, '1')

			ip = getNewIp(subnet, offset)
			break
		}
	}

	ipam.dump()
	return
}

func getNewIp(subnet *net.IPNet, offset int) (ip net.IP) {
	// ip is a uint array, [1,2,3,4]
	ip = subnet.IP.To4()
	for t := uint(4); t > 0; t-- {
		[]byte(ip)[4 - t] += uint8(offset >> ((t - 1) * 8))
	}
	// ip is allocated from 1, not 0
	ip[3]++
	return ip
}

func getOffset(subnet *net.IPNet, ipaddr *net.IP) int {
	ip := ipaddr.To4()
	offset := 0
	ip[3]--
	for t := uint(4); t > 0; t-- {
		offset += int(ip[t - 1] - subnet.IP.To4()[t - 1]) << ((4 - t) * 8)
	}
	return offset
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if nil != err {
		logrus.Errorf("Error load allocation info, %v", err)
	}

	offset := getOffset(subnet, ipaddr)

	setValue(ipam, subnet, offset, '0')

	ipam.dump()
	return nil
}

func setValue(ipam *IPAM, subnet *net.IPNet, offset int, value byte) {
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[offset] = value
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)
}
