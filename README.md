## network

should use the following cmd in the host
```shell script
sysctl -w net.ipv4.conf.all.forwarding=1
```

### cmd

build a bridge
```shell script
ip netns add ns1
ip link add veth0 type veth peer name veth1
ip link set veth1 netns ns1
brctl addbr br0
brctl addif br0 ens3
brctl addif br0 veth0
ip link set veth0 up
ip link set br0 up
ip netns exec ns1 ifconfig veth1 192.18.0.2/24 up
ip netns exec ns1 route add default dev veth1
route add -net 192.18.0.0/24 dev br0
```

show info
```shell script
route -n
iptables -t nat -nvL
```