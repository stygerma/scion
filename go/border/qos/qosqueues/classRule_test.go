package qosqueues

// TODO: Add tests for MatchModes as soon as you have decided which thing

// func TestGetEqualQueueNumbers(t *testing.T) {

// 	r, _ := setupTestRouter(t)

// 	r.initQueueing("sample-config.yaml")

// 	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

// 	queueNo1 := getQueueNumberIterativeForInternal(r, pkt)
// 	queueNo2 := getQueueNumberIterativeForInternal(r, pkt)
// 	queueNo3 := getQueueNumberIterativeFor(r, pkt)

// 	fmt.Println("Queue Numbers", queueNo1, queueNo2, queueNo3)

// 	if !(queueNo1 == queueNo2 && queueNo2 == queueNo3) {
// 		t.Errorf("Queue Numbers are incorrect %d %d %d should all be equal", queueNo1, queueNo2, queueNo3)
// 	}

// }

// func TestMultipleRuleMatchesHigh(t *testing.T) {

// 	r, _ := setupTestRouter(t)

// 	r.initQueueing("testdata/testConfig1.yaml")

// 	pkt := rpkt.JFPrepareRtrPacketWithSrings("1-ff00:0:110", "1-ff00:0:111", 1)

// 	queueNo1 := getQueueNumberIterativeForInternal(r, pkt)

// 	if queueNo1 != 15 {
// 		t.Errorf("Wrong queuenumber should be %d but is %d ", 15, queueNo1)
// 	}
// }

// func TestMultipleRuleMatchesLow(t *testing.T) {

// 	r, _ := setupTestRouter(t)

// 	r.initQueueing("testdata/testConfig2.yaml")

// 	pkt := rpkt.JFPrepareRtrPacketWithSrings("1-ff00:0:110", "1-ff00:0:111", 1)

// 	queueNo1 := getQueueNumberIterativeForInternal(r, pkt)

// 	if queueNo1 != 2 {
// 		t.Errorf("Wrong queuenumber should be %d but is %d ", 2, queueNo1)
// 	}
// }

// func BenchmarkIterativeBasic(b *testing.B) {

// 	r = &Router{Id: "TestRouter"}
// 	r.initQueueing("sample-config.yaml")
// 	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)

// 	queueNo1 := getQueueNumberIterativeForInternal(r, pkt)
// 	fmt.Println("Queue Number is", queueNo1)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		queueNo1 = getQueueNumberIterativeForInternal(r, pkt)
// 	}
// }

// func setupInterm(numberOfPackets int, configPath string) (*Router, []*rpkt.RtrPkt) {

// 	r = &Router{Id: "TestRouter"}
// 	r.initQueueing(configPath)
// 	pkt := rpkt.PrepareRtrPacketWithStrings("1-ff00:0:110", "1-ff00:0:111", 1)
// 	pkt1 := rpkt.PrepareRtrPacketWithStrings("2-ff00:0:212", "1-ff00:0:111", 1)
// 	pkt2 := rpkt.PrepareRtrPacketWithStrings("3-ff00:0:212", "1-ff00:0:111", 1)
// 	pkt3 := rpkt.PrepareRtrPacketWithStrings("4-ff00:0:212", "1-ff00:0:111", 1)
// 	pkt4 := rpkt.PrepareRtrPacketWithStrings("5-ff00:0:212", "1-ff00:0:111", 1)
// 	pkt5 := rpkt.PrepareRtrPacketWithStrings("6-ff00:0:212", "1-ff00:0:111", 1)

// 	arr := make([]*rpkt.RtrPkt, numberOfPackets)

// 	for i := 0; i < numberOfPackets-5; i++ {
// 		arr[i+0] = pkt
// 		arr[i+1] = pkt1
// 		arr[i+2] = pkt2
// 		arr[i+3] = pkt3
// 		arr[i+4] = pkt4
// 		arr[i+5] = pkt5
// 	}

// 	return r, arr

// }

// func BenchmarkBig(b *testing.B) {
// 	const noPackets = 10

// 	benchmarks := []struct {
// 		name          string
// 		funct         func(r *Router, rp *rpkt.RtrPkt) int
// 		configuration string
// 	}{
// 		{"Naive implementation with string comparison 600 packets", getQueueNumberIterativeFor, "sample-config.yaml"},
// 		{"Naive implementation with int comparison 600 packets", getQueueNumberIterativeForInternal, "sample-config.yaml"},
// 		{"Map based implementation 600 packets", GetQueueNumberWithHashFor, "sample-config.yaml"},
// 		{"Naive implementation with string comparison 600 packets", getQueueNumberIterativeFor, "bench-config-medium.yaml"},
// 		{"Naive implementation with int comparison 600 packets", getQueueNumberIterativeForInternal, "bench-config-medium.yaml"},
// 		{"Map based implementation 600 packets", GetQueueNumberWithHashFor, "bench-config-medium.yaml"},
// 		{"Naive implementation with string comparison 600 packets", getQueueNumberIterativeFor, "bench-config-large.yaml"},
// 		{"Naive implementation with int comparison 600 packets", getQueueNumberIterativeForInternal, "bench-config-large.yaml"},
// 		{"Map based implementation 600 packets", GetQueueNumberWithHashFor, "bench-config-large.yaml"},
// 	}

// 	for _, bm := range benchmarks {
// 		b.Run(bm.name, func(b *testing.B) {
// 			r, arr := setupInterm(noPackets, bm.configuration)
// 			b.ResetTimer()
// 			for i := 0; i < b.N; i++ {
// 				// fmt.Println("Rules", len(r.legacyConfig.Rules))
// 				for i := 0; i < noPackets; i++ {
// 					bm.funct(r, arr[i])
// 				}
// 			}
// 		})
// 	}
// }
