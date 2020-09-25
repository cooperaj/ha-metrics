package main

import "sync"

type collector interface {
	Monitor(*sync.WaitGroup)
}
