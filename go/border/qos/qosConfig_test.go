package qos

import (
	"fmt"
	"testing"
)

const configFileLocation = "/home/fischjoe/go/src/github.com/joelfischerr/scion/go/border/qos/sample-config.yaml"

func TestForMarc(t *testing.T) {

	fmt.Println("Hello test 2")

	qosConfig, _ := InitQueueing(configFileLocation, nil)

	fmt.Println("Config is", qosConfig)

	fmt.Println("Name is", qosConfig.config.Queues[0].GetPacketQueue().Name)
	fmt.Println("Name is", qosConfig.config.Queues[1].GetPacketQueue().Name)
	fmt.Println("Name is", qosConfig.config.Queues[2].GetPacketQueue().Name)

	fmt.Println("Profile is", qosConfig.config.Queues[0].GetPacketQueue().Profile)
	fmt.Println("Profile is", qosConfig.config.Queues[1].GetPacketQueue().Profile)
	fmt.Println("Profile is", qosConfig.config.Queues[2].GetPacketQueue().Profile)

	fmt.Println("CongWarning is", qosConfig.config.Queues[0].GetPacketQueue().CongWarning)
	fmt.Println("CongWarning is", qosConfig.config.Queues[1].GetPacketQueue().CongWarning)
	fmt.Println("CongWarning is", qosConfig.config.Queues[2].GetPacketQueue().CongWarning)

	t.Errorf("FUCK")
	t.Errorf("IntMin(2, -2) = %d; want -2", 17)
}
