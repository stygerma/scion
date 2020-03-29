// Handles the notifications for the Basic approach, i.e. just send one SCMP message to each source of traffic.

package main

import (
	"github.com/scionproto/scion/go/border/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/layers"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scmp"
)

const logEnabled = true

//Sends notification for each
func (r *Router) bscNotify() {
	for np := range r.notifications {

		//TODO: uncomment later when all approaches are implemented
		//if r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning().Approach == 0 {
		if logEnabled {
			/*srcIA, _ := np.Qpkt.Rp.SrcIA()
			srcHost, _ := np.Qpkt.Rp.SrcHost()
			DstIA, _ := np.Qpkt.Rp.DstIA()
			DstHost, _ := np.Qpkt.Rp.DstHost()
			log.Debug("New notification packet", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost)*/
			log.Debug("New notification packet", "NPkt", np, "Pkt ID", np.Qpkt.Rp.Id)
		}
		bscCW := r.createCongWarn(np)
		if logEnabled {
			log.Debug("Created basic congestion warning", "bscCW", bscCW, "Pkt ID", np.Qpkt.Rp.Id)
		}
		r.sendNotificationSCMP(np.Qpkt, bscCW)
		//}
	}

}

//check creation of messages from local AS, such messages are not created in the doPKTError method
//Similar to PacketError method maybe could assign the approaches in here too
func (r *Router) sendNotificationSCMP(qp *qosqueues.QPkt, info scmp.Info) {
	if logEnabled {
		srcIA, _ := qp.Rp.SrcIA()
		srcHost, _ := qp.Rp.SrcHost()
		DstIA, _ := qp.Rp.DstIA()
		DstHost, _ := qp.Rp.DstHost()
		log.Debug("New queueing packet", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "QNo", qp.QueueNo, "Pkt ID",
			qp.Rp.Id)
	}
	notification, err, id := r.createSCMPNotification(qp, scmp.ClassType{Class: scmp.C_General, Type: scmp.T_G_BasicCongWarn}, info)
	if err != nil {
		log.Error("unable to create notification SCMP", "err", err)
		return
	}
	if logEnabled {
		srcIA, _ := notification.SrcIA()
		srcHost, _ := notification.SrcHost()
		DstIA, _ := notification.DstIA()
		DstHost, _ := notification.DstHost()
		pub := qp.Rp.Ctx.Conf.BR.InternalAddr
		routerAddr := addr.HostFromIP(pub.IP)
		log.Debug("New SCMP Notification", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "RtrAddr", routerAddr,
			"CurrBW", r.config.Queues[qp.QueueNo].GetTokenBucket().CurrBW, "Pkt ID", id)
	}
	notification.Route()

}

func (r *Router) createSCMPNotification(qp *qosqueues.QPkt,
	ct scmp.ClassType, info scmp.Info) (*rpkt.RtrPkt, error, string) {

	if logEnabled {
		srcIA, _ := qp.Rp.SrcIA()
		srcHost, _ := qp.Rp.SrcHost()
		DstIA, _ := qp.Rp.DstIA()
		DstHost, _ := qp.Rp.DstHost()
		log.Debug("New queueing packet", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "Pkt ID", qp.Rp.Id)
	}
	id := qp.Rp.Id
	sp, err := qp.Rp.CreateReplyScnPkt()
	if err != nil {
		return nil, err, ""
	}
	qp.Rp.RefInc(-1) //original packet is not needed anymore, possible to free it

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
	log.Debug("Created SPkt reply", "sp", sp, "Pkt ID", id)
	reply, err := qp.Rp.CreateReply(sp)
	if logEnabled {
		srcIA, _ := reply.SrcIA()
		srcHost, _ := reply.SrcHost()
		DstIA, _ := reply.DstIA()
		DstHost, _ := reply.DstHost()
		log.Debug("Created RPkt reply", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "Pkt ID", id)
	}
	return reply, err, id
}

func (r *Router) createCongWarn(np *qosqueues.NPkt) *scmp.InfoBscCW {
	restrictionPrint := r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning()
	testing := r.config.Queues[np.Qpkt.QueueNo].GetMinBandwidth()
	restriction := 3
	if logEnabled {
		log.Debug("restrictions on information content", "restriction", restrictionPrint, "MinBW", testing)
	}
	if restriction > 3 {
		log.Error("Unable to create congestion warning", "restriction on information content", restriction)
		return nil
	}
	bscCW := &scmp.InfoBscCW{}
	bscCW.ConsIngress = common.IFIDType(np.Qpkt.Rp.Ingress.IfID)
	if logEnabled {
		log.Debug("InfoBscCW", "ConsIngress", common.IFIDType(np.Qpkt.Rp.Ingress.IfID),
			"QueueLength", (r.config.Queues[np.Qpkt.QueueNo]).GetLength(), "CurrBW",
			r.config.Queues[np.Qpkt.QueueNo].GetTokenBucket().CurrBW, "QueueFullness",
			(r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel(), "Violation", np.Qpkt.Act.Reason)
	}
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
	return bscCW
}
