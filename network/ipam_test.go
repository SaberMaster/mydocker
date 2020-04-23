package network

import (
	"net"
	"reflect"
	"testing"
)

func TestIPAM_Allocate(t *testing.T) {
	type fields struct {
		SubnetAllocatorPath string
		Subnets             *map[string]string
	}
	type args struct {
		subnet *net.IPNet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantIp  net.IP
		wantErr bool
	}{
		{
			name: "simple",
			fields: fields{
				Subnets:             nil,
				SubnetAllocatorPath: "/tmp/subnet.json",
			},
			args: args{
				subnet: &net.IPNet{
					IP:   net.IPv4(192, 168, 0, 0),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			wantIp:  net.IPv4(192, 168, 0, 1),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := &IPAM{
				SubnetAllocatorPath: tt.fields.SubnetAllocatorPath,
				Subnets:             tt.fields.Subnets,
			}
			gotIp, err := ipam.Allocate(tt.args.subnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Allocate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIp.To4(), tt.wantIp.To4()) {
				t.Errorf("Allocate() gotIp = %v, want %v", gotIp, tt.wantIp)
			}
		})
	}
}

func TestIPAM_Release(t *testing.T) {
	type fields struct {
		SubnetAllocatorPath string
		Subnets             *map[string]string
	}
	type args struct {
		subnet *net.IPNet
		ipaddr *net.IP
	}
	pv4 := net.IPv4(192, 168, 0, 1)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "simple",
			fields: fields{
				Subnets:             nil,
				SubnetAllocatorPath: "/tmp/subnet.json",
			},
			args: args{
				subnet: &net.IPNet{
					IP:   net.IPv4(192, 168, 0, 0),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
				ipaddr: &pv4,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := &IPAM{
				SubnetAllocatorPath: tt.fields.SubnetAllocatorPath,
				Subnets:             tt.fields.Subnets,
			}
			if err := ipam.Release(tt.args.subnet, tt.args.ipaddr); (err != nil) != tt.wantErr {
				t.Errorf("Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPAM_dump(t *testing.T) {
	type fields struct {
		SubnetAllocatorPath string
		Subnets             *map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := &IPAM{
				SubnetAllocatorPath: tt.fields.SubnetAllocatorPath,
				Subnets:             tt.fields.Subnets,
			}
			if err := ipam.dump(); (err != nil) != tt.wantErr {
				t.Errorf("dump() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPAM_load(t *testing.T) {
	type fields struct {
		SubnetAllocatorPath string
		Subnets             *map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := &IPAM{
				SubnetAllocatorPath: tt.fields.SubnetAllocatorPath,
				Subnets:             tt.fields.Subnets,
			}
			if err := ipam.load(); (err != nil) != tt.wantErr {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getNewIp(t *testing.T) {
	type args struct {
		subnet *net.IPNet
		offset int
	}
	tests := []struct {
		name   string
		args   args
		wantIp net.IP
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIp := getNewIp(tt.args.subnet, tt.args.offset); !reflect.DeepEqual(gotIp, tt.wantIp) {
				t.Errorf("getNewIp() = %v, want %v", gotIp, tt.wantIp)
			}
		})
	}
}

func Test_getOffset(t *testing.T) {
	type args struct {
		subnet *net.IPNet
		ipaddr *net.IP
	}
	tests := []struct {
		name string
		args args
		want int
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOffset(tt.args.subnet, tt.args.ipaddr); got != tt.want {
				t.Errorf("getOffset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setValue(t *testing.T) {
	type args struct {
		ipam   *IPAM
		subnet *net.IPNet
		offset int
		value  byte
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
