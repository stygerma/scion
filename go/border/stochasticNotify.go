//Implements the stochastic notification approach

package main

type stochastic struct {
	SwitchingPoint float64
	//Controller     PID
}

// func (r *Router) stochNotify() {
// 	for np := range r.notifications {

// 		//TODO: uncomment later when all approaches are implemented
// 		//if r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning().Approach == 2 {
// 		//TODO: calculate probability and only create the cong warn with that probability

// 		//}
// 	}
// }
