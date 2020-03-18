//Implements the hop-by-hop notification approach

package main

import (
	"time"
)

type hbpSelection struct {
	timestamp time.Time     //starting time of the time interval
	interval  time.Duration //Length of the time interval
	pkts      []qPkt        //packets of sources that need to be notified
}
