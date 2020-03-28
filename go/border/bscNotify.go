// Handles the notifications for the Basic approach, i.e. just send one SCMP message to each source of traffic.

package main

import (
	"github.com/scionproto/scion/go/border/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/layers"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scmp"
)

//Sends notification for each
func (r *Router) bscNotify() {
	for np := range r.notifications {
		//TODO:	create SCMP message with createNotificationSCMP()
		bscCW := r.createCongWarn(np)
		r.sendNotificationSCMP(np.Qpkt, bscCW)
	}

}

//Similar to PacketError method maybe could assign the approaches in here too
func (r *Router) sendNotificationSCMP(qp *qosqueues.QPkt, info scmp.Info) {

	notification, err := r.createSCMPNotification(qp, scmp.ClassType{Class: scmp.C_General, Type: scmp.T_G_BasicCongWarn}, info)
	if err != nil {
		log.Error("unable to create notification SCMP", "err", err)
		return
	}
	notification.Route()

}

func (r *Router) createSCMPNotification(qp *qosqueues.QPkt,
	ct scmp.ClassType, info scmp.Info) (*rpkt.RtrPkt, error) {

	sp, err := qp.Rp.CreateReplyScnPkt()
	if err != nil {
		return nil, err
	}
	sp.HBHExt = make([]common.Extension, 0, common.ExtnMaxHBH+1)
	/*MS: We classify the congestion warning as not erroneous and don't need the
	Basic congestion warning to be HBH*/
	ext := &layers.ExtnSCMP{Error: false, HopByHop: false}
	sp.HBHExt = append(sp.HBHExt, ext)
	//TODO: SCMP authentication needs to be added as extension, but I'm not sure if it is completely implemented
	/*drkeyExt := scmp_auth.NewDRKeyExtn() //this does not seem to be used yet for scmps
	if err := drkeyExt.SetDirection(1); err := nil { //Q: not sure if this is the right direction (parameter of the function)
		return nil, err
	}
	if err := drkeyExt.SetMAC()
	*/

	sp.Pld = scmp.NotifyPld(info)
	sp.L4 = scmp.NewHdr(scmp.ClassType{Class: scmp.C_General, Type: scmp.T_G_BasicCongWarn}, sp.Pld.Len())
	return qp.Rp.CreateReply(sp)
}

//TODO: include information according to restrictions
func (r *Router) createCongWarn(np *qosqueues.NPkt) *scmp.InfoBscCW {
	restriction := r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning().InfoContent
	if restriction > 3 {
		log.Error("Unable to create congestion warning", "restriction on information content", restriction)
		return nil
	}
	bscCW := &scmp.InfoBscCW{}
	bscCW.ConsIngress = common.IFIDType(np.Qpkt.Rp.Ingress.IfID)
	/*EndHost, err := np.Qpkt.Rp.SrcHost()
	if err != nil {
		log.Error("Unable to create congestion warning", "restriction on information content", restriction)
		return nil
	}
	bscCW.EndHost = EndHost*/
	if restriction > 0 {
		bscCW.QueueLength = uint64((r.config.Queues[np.Qpkt.QueueNo]).GetLength())
	}
	if restriction > 1 {
		bscCW.CurrBW = r.config.Queues[np.Qpkt.QueueNo].GetTokenBucket().CurrBW
		bscCW.QueueFullness = uint64((r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel())
	}
	if restriction > 2 {
		bscCW.Violation = uint64(np.Qpkt.Act.Reason)
		bscCW.ClassRule = np.Qpkt.Act.Rule
	}
	//bscCW.ClassRule = np.Rule
	return bscCW
}
