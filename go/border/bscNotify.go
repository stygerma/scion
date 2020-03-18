//Handles the notifications for the Basic approach, i.e. just send one SCMP message to each source of traffic.

package main

//Sends notification for each
func (r *Router) bscNotify() {
	for {
		select {
		case qp := <-r.notifications:
			qp.sendNotify()
		}

	}

}

func (qp *qPkt) sendNotify() error {

}
