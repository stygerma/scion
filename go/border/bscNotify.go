//Handles the notifications for the Basic approach, i.e. just send one SCMP message to each source of traffic.

package main

import (
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/scmp"
	"github.com/scionproto/scion/go/lib/spse/scmp_auth"
)

//Sends notification for each
func (r *Router) bscNotify() {
	for {
		select {
		case qp := <-r.notifications:
			//TODO: create scmp.BscCongWarn according to qp, the restrictions imposed by the ISP and the current IF state
			//TODO:	create SCMP message with createNotificationSCMP()
			//if bscCW, err := do this in createNotificationSCMP
			if rpkt, err := r.createNotificationSCMP(qp); err := nil {
				log.Debug("unable to create notification SCMP")
			}
			rpkt.Route()
		}

	}

}


func (r *Router) createNotificationSCMP(qp *qPkt) (*rpkt.RtrPkt, error) {
	//BC: basically just modification from the createSCMPErrorReply method from border/error.go


	sp, err := qp.rp.CreateReplyScnPkt()
	if err != nil {
		return nil, err
	}
	sp.HBHExt = make([]common.Extension, 0, common.ExtnMaxHBH+1)
	//MS: We classify the congestion warning as not erroneous and don't need the Basic congestion warning to be HBH
	ext := &layers.ExtnSCMP{Error: false, HopByHop: false}
	sp.HBHExt = append(sp.HBHExt, ext)
	//TODO: SCMP authentication needs to be added as extension, but I'm not sure if it is completely implemented  
	/*drkeyExt := scmp_auth.NewDRKeyExtn() //this does not seem to be used yet for scmps 
	if err := drkeyExt.SetDirection(1); err := nil { //Q: not sure if this is the right direction (parameter of the function)
		return nil, err
	} 
	if err := drkeyExt.SetMAC()
	*/
	bscCongWarn := createCongWarn(qp)
	sp.Pld = createPld(bscCongWarn)
	sp.L4 = scmp.NewHdr(&scmp.ClassType{scmp.Class : scmp.C_General, scmp.Type : scmp.T_G_BasicCongWarn}, sp.Pld.Len())
	return rp.CreateReply(sp)
}

func createPld() *scmp.Payload {
	
}


func (r *Router) createCongWarn(qp *qPkt) *scmp.BscCongWarn {
	ifCongState := &scmp.IfCongState{}
	ifCongState.CurrBW = r.config.Queues[qp.queueNo].tb.currBW
	ifCongState.QueueLength = r.config.Queues[qp.queueNo].length
	ifCongState.QueueFullness = r.config.Queues[qp.queueNo].queue.len()
	ifCongState.violation = qp.act.reason
	//TODO find way to include something that identifies the BR or its interface 
	return &scmp.BscCongWarn{scmp.IfCongState: ifCongState}
}

