package main

import (
	"fmt"
	"testing"
)

func TestLoadSampleConfig(t *testing.T) {
	r, _ := setupTestRouter(t)

	r.loadConfigFile("sample-config.yaml")

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
