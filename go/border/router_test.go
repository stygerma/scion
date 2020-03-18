package main

import (
	"fmt"
	"testing"

	"github.com/scionproto/scion/go/lib/addr"
)

func TestLoadSampleConfig(t *testing.T) {
	r, _ := setupTestRouter(t)

	r.loadConfigFile("sample-config.yaml")

	fmt.Println("The config is: ", r.config)

	fmt.Println("The Queue is: ", r.config.Queues[0])
	fmt.Println("The Rule is: ", r.config.Rules[0])
	// t.Errorf("Output: %v", r.config)

}

func TestLoadSampleConfigQueues(t *testing.T) {
	r, _ := setupTestRouter(t)

	r.loadConfigFile("sample-config.yaml")

	fmt.Println("The Queue is: ", r.config.Queues[0])
	fmt.Println("The Rule is: ", r.config.Rules[0])
	fmt.Println("We have this number of queues: ", len(r.config.Queues))
	t.Errorf("Output: %v", r.config)

	if r.config.Queues[0].ID != 0 {
		t.Errorf("Incorrect Queue ID")
	}
	if r.config.Queues[1].ID != 1 {
		t.Errorf("Incorrect Queue ID")
	}

	if r.config.Queues[0].Name != "General Queue" {
		t.Errorf("Incorrect Queue Name is %v but should be %v", r.config.Queues[0].Name, "General Queue")
	}
	if r.config.Queues[1].Name != "Speedy Queue" {
		t.Errorf("Incorrect Queue Name is %v but should be %v", r.config.Queues[0].Name, "Speedy Queue")
	}
}

func TestMaps(t *testing.T) {
	m := make(map[addr.IA]*internalClassRule)

	IA1, _ := addr.IAFromString("1-ff00:0:110")
	IA2, _ := addr.IAFromString("2-ff00:0:110")
	IA3, _ := addr.IAFromString("3-ff00:0:110")
	IA4, _ := addr.IAFromString("4-ff00:0:110")

	rul1 := internalClassRule{Name: "Hello Test", SourceAs: matchRule{IA: IA1}}
	rul2 := internalClassRule{Name: "Hello World", SourceAs: matchRule{IA: IA2}}
	rul3 := internalClassRule{Name: "Hello SCION", SourceAs: matchRule{IA: IA3}}
	rul4 := internalClassRule{Name: "Hello Internet", SourceAs: matchRule{IA: IA4}}

	m[IA1] = &rul1
	m[IA2] = &rul2
	m[IA3] = &rul3
	m[IA4] = &rul4

	search, _ := addr.IAFromString("3-ff00:0:110")

	rule, found := m[search]
	fmt.Println("We have found", found, rule)

	search, _ = addr.IAFromString("5-ff00:0:110")

	rule, found = m[search]
	fmt.Println("We have found", found, rule)

	t.Errorf("See logs")

}
