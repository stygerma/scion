//Implements the hop-by-hop notification approach

package main

import (
	"time"
)

type hbhSelection struct {
	timestamp time.Time     //starting time of the time interval
	interval  time.Duration //Length of the time interval
	pkts      []QPkt        //packets of sources that need to be notified
}

/*func (hbhS *hbhSelection) createNotificationSCMP() error {

}*/
