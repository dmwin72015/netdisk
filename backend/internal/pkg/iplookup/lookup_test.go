package iplookup

import "testing"

func TestRegionString(t *testing.T) {
	tests := []struct {
		name   string
		region Region
		want   string
	}{
		{name: "all fields", region: Region{Country: "中国", Province: "北京市", City: "北京", ISP: "电信"}, want: "中国|北京市|北京|电信"},
		{name: "country only", region: Region{Country: "中国"}, want: "中国"},
		{name: "country and city", region: Region{Country: "中国", City: "上海"}, want: "中国|上海"},
		{name: "province and ISP", region: Region{Province: "广东省", ISP: "联通"}, want: "广东省|联通"},
		{name: "empty region", region: Region{}, want: ""},
		{name: "ISP only", region: Region{ISP: "移动"}, want: "移动"},
		{name: "country and province", region: Region{Country: "美国", Province: "California"}, want: "美国|California"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.region.String()
			if got != tt.want {
				t.Errorf("Region.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNoopLookup(t *testing.T) {
	lookup := NewNoopLookup()

	tests := []struct {
		name string
		ip   string
	}{
		{name: "valid IPv4", ip: "8.8.8.8"},
		{name: "valid IPv6", ip: "2001:4860:4860::8888"},
		{name: "empty string", ip: ""},
		{name: "invalid IP", ip: "not-an-ip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, err := lookup.Lookup(tt.ip)
			if err != nil {
				t.Errorf("noopLookup.Lookup(%q) returned unexpected error: %v", tt.ip, err)
			}
			if region.String() != "" {
				t.Errorf("noopLookup.Lookup(%q) = %q, want empty", tt.ip, region.String())
			}
		})
	}
}
