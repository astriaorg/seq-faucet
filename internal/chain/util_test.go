package chain

import (
	"testing"
)

func TestIsValidAddress(t *testing.T) {
	type args struct {
		address     string
		checksummed bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "valid address", args: args{address: "0xab5801a7d398351b8be11c439e05c5b3259aec9b", checksummed: false}, want: true},
		{name: "invalid address", args: args{address: "invalid address", checksummed: false}, want: false},
		{name: "address without 0x", args: args{address: "ab5801a7d398351b8be11c439e05c5b3259aec9b", checksummed: false}, want: true},
		{name: "valid checksum address", args: args{address: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B", checksummed: true}, want: true},
		{name: "invalid checksum address", args: args{address: "0xab5801a7d398351b8be11c439e05c5b3259aec9b", checksummed: true}, want: false},
		{name: "checksum address without 0x", args: args{address: "Ab5801a7D398351b8bE11C439e05C5B3259aeC9B", checksummed: true}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidAddress(tt.args.address, tt.args.checksummed); got != tt.want {
				t.Errorf("IsValidAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHas0xPrefix(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "has 0x prefix", args: args{str: "0xab5801a7d398351b8be11c439e05c5b3259aec9b"}, want: true},
		{name: "has no 0X prefix", args: args{str: "ab5801a7d398351b8be11c439e05c5b3259aec9b"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Has0xPrefix(tt.args.str); got != tt.want {
				t.Errorf("Has0xPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
