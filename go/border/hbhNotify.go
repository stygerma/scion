//Implements the hop-by-hop notification approach

package main

import (
	"time"
)

<<<<<<< HEAD
type hbhSelection struct {
=======
type hbpSelection struct {
>>>>>>> e9ea163d9f4c7bbb1d1a067fd6cd04bb881e6b81
	timestamp time.Time     //starting time of the time interval
	interval  time.Duration //Length of the time interval
	pkts      []qPkt        //packets of sources that need to be notified
}
<<<<<<< HEAD

func (hbhS *hbhSelection) createNotificationSCMP() error {

}
=======
>>>>>>> e9ea163d9f4c7bbb1d1a067fd6cd04bb881e6b81
