package main

import "strings"

type diskSlice []string
type netIOIfaceSlice []string

func (d *diskSlice) String() string {
	return ""
}

func (d *diskSlice) Set(value string) error {
	*d = append(*d, strings.TrimSpace(value))
	return nil
}

func (n *netIOIfaceSlice) String() string {
	return ""
}

func (n *netIOIfaceSlice) Set(value string) error {
	*n = append(*n, strings.TrimSpace(value))
	return nil
}
