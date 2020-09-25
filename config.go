package main

import "strings"

type diskSlice []string

func (d *diskSlice) String() string {
	return ""
}

func (d *diskSlice) Set(value string) error {
	*d = append(*d, strings.TrimSpace(value))
	return nil
}
