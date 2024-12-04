package util

import (
	"fmt"
	"slices"
	"strings"
)

var (
	allowedOS = []string{
		"linux",
		"windows",
		"darwin",
	}
	allowedArch = []string{
		"amd64",
		"arm64",
	}
	DefaultPlatform = Platform{
		OS: "linux",
		Arch: "amd64",
	}
)

type Platform struct {
	OS string
	Arch string
}

func (p *Platform) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

func GetAllAllowedPlatform() []Platform {
	var allowed []Platform
	for _, os := range allowedOS {
		for _, arch := range allowedArch {
			allowed = append(allowed, Platform{OS: os, Arch: arch})
		}
	}
	return allowed
}

func ParsePlatform(platformText string) *Platform {
	platformText = strings.TrimSpace(platformText)
	infos := strings.Split(platformText, "/")
	if len(infos) != 2 {
		return nil
	}
	if slices.Contains(allowedOS, infos[0]) && slices.Contains(allowedArch, infos[1]) {
		return &Platform{
			OS: infos[0],
			Arch: infos[1],
		}
	}
	return nil
}
