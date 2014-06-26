package main

import (
	"fmt"
	"github.com/armon/consul-api"
	"log"
	"strings"
)

func stringInSlice(s string, list []string) bool {
	for _, elem := range list {
		if elem == s {
			return true
		}
	}
	return false
}

type ConsulLookUp struct {
	client       *consulapi.Client
	RemoveWithKV string
}

func (l *ConsulLookUp) checkKVDowned() ([]string, error) {
	down := []string{}
	if l.RemoveWithKV != "" {
		pairs, meta, err := l.client.KV().List(l.RemoveWithKV, nil)
		if err != nil {
			return nil, err
		}
		if !meta.KnownLeader {
			return nil, fmt.Errorf("Error discovering nodes. No leader in the cluster")
		}
		for _, p := range pairs {
			path := strings.Split(p.Key, "/")
			node := path[len(path)-1]
			down = append(down, node)
		}
	}
	return down, nil
}

func (l *ConsulLookUp) LookupService(name string) ([]*consulapi.ServiceEntry, error) {
	health := l.client.Health()
	// TODO allow just dependent checks instead of fully healthy node?
	entries, meta, err := health.Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}
	if !meta.KnownLeader {
		log.Printf("Error discovering nodes. No leader in the cluster")
		return nil, fmt.Errorf("Error discovering nodes. No leader in the cluster")
	}
	nodes, err := l.checkKVDowned()
	if err != nil {
		log.Printf("Error trying to find down nodes %v", err)
		return nil, err
	}
	if len(nodes) > 0 {
		prunedEntries := make([]*consulapi.ServiceEntry, 0, len(nodes))
		for _, entry := range entries {
			if stringInSlice(entry.Node.Node, nodes) {
				log.Printf("Node '%s' has been manually downed", entry.Node.Node)
			} else {
				prunedEntries = append(prunedEntries, entry)
			}
		}
		entries = prunedEntries
	}
	return entries, nil
}
