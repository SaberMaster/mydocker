package network

import (
	"net"
	"reflect"
	"testing"
)

func TestBridgeNetworkDriver_Create(t *testing.T) {
	type args struct {
		subnet string
		name   string
	}
	tests := []struct {
		name    string
		args    args
		want    *Network
		wantErr bool
	}{
		{
			name:    "simple",
			args:    args{
				subnet: "192.168.0.1/24",
				name:   "test_bridge",
			},
			want:    &Network{
				Name:    "test_bridge",
				IpRange: &net.IPNet{
					IP:   net.IPv4(192, 168, 0, 1),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
				Driver:  "bridge",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &BridgeNetworkDriver{}
			got, err := driver.Create(tt.args.subnet, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBridgeNetworkDriver_Delete(t *testing.T) {
	type args struct {
		network Network
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "simple",
			args:    args{
				network: Network{
					Name:    "test_bridge",
					IpRange: &net.IPNet{
						IP:   net.IPv4(192, 168, 0, 1),
						Mask: net.IPv4Mask(255, 255, 255, 0),
					},
					Driver:  "bridge",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := &BridgeNetworkDriver{}
			if err := bridge.Delete(tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridgeNetworkDriver_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &BridgeNetworkDriver{}
			if got := driver.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBridgeNetworkDriver_initBridge(t *testing.T) {
	type args struct {
		network *Network
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &BridgeNetworkDriver{}
			if err := driver.initBridge(tt.args.network); (err != nil) != tt.wantErr {
				t.Errorf("initBridge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createBridgeInterface(t *testing.T) {
	type args struct {
		bridgeName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createBridgeInterface(tt.args.bridgeName); (err != nil) != tt.wantErr {
				t.Errorf("createBridgeInterface() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setInterfaceIP(t *testing.T) {
	type args struct {
		name  string
		rawIP string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setInterfaceIP(tt.args.name, tt.args.rawIP); (err != nil) != tt.wantErr {
				t.Errorf("setInterfaceIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setInterfaceUp(t *testing.T) {
	type args struct {
		interfaceName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setInterfaceUp(tt.args.interfaceName); (err != nil) != tt.wantErr {
				t.Errorf("setInterfaceUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setupIPTables(t *testing.T) {
	type args struct {
		bridgeName string
		subnet     *net.IPNet
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setupIPTables(tt.args.bridgeName, tt.args.subnet); (err != nil) != tt.wantErr {
				t.Errorf("setupIPTables() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}