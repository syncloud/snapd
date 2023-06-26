package model

import "strings"

type SnapId string

func (s SnapId) Name() string {
	if strings.Contains(string(s), ".") {
		parts := strings.Split(string(s), ".")
		return parts[0]
	} else {
		return string(s)
	}
}

func (s SnapId) Version() string {
	if strings.Contains(string(s), ".") {
		parts := strings.Split(string(s), ".")
		return parts[1]
	} else {
		return ""
	}
}
