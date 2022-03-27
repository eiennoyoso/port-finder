package main

type PortRange struct {
	minPort int
	maxPort int
}

func (r *PortRange) Size() int {
	return r.maxPort - r.minPort
}
