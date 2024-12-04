package util

import "testing"

func TestParsePlatform(t *testing.T) {
	for _, x := range []struct{
		platform string
		notNil bool
	} {
		{"linux/amd64", true},
		{"linux/ia64", false},
	} {
		p := ParsePlatform(x.platform)
		notNil := p != nil
		if notNil != x.notNil {
			t.Errorf("%s's notNil != %v", x.platform, x.notNil)
		}
	}
}
