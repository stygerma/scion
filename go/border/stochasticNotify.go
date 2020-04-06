//Implements the stochastic notification approach

package main

import (
	"math/rand"

	"github.com/scionproto/scion/go/border/qosqueues"
	"github.com/scionproto/scion/go/border/rpkt"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/layers"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scmp"
)

const (
	logEnabledStoch = false
)

func (r *Router) stochNotify() {
	for np := range r.notifications { //TODO: if we don't include the classrule we could just use QPkt instead of NPkt
		//if r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning().Approach == 2 { //TODO: remove when congestionWarning fields are read out correctly
		queueFullness := float64((r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel())
		switchingPoint, output := (r.config.Queues[np.Qpkt.QueueNo].GetPID()).NewControlUpdate(queueFullness)
		probs := r.calculateProbs(np.Qpkt, switchingPoint)
		random := rand.Intn(100)

		log.Debug("Stochastic Notify stats", "probs", probs, "random", random, "switching Point", switchingPoint, "PID output", output)
		if random <= probs {
			stochCW := r.createStochCongWarn(np)
			r.sendStochNotificationSCMP(np.Qpkt, stochCW)
			np.Qpkt.Rp.RefInc(-1)
		}
		//}
	}
}

func (r *Router) sendStochNotificationSCMP(qp *qosqueues.QPkt, info scmp.Info) {
	if logEnabledStoch {
		srcIA, _ := qp.Rp.SrcIA()
		srcHost, _ := qp.Rp.SrcHost()
		DstIA, _ := qp.Rp.DstIA()
		DstHost, _ := qp.Rp.DstHost()
		log.Debug("New queueing packet", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "QNo", qp.QueueNo, "Pkt ID",
			qp.Rp.Id)
	}
	notification, err, id := r.createStochSCMPNotification(qp, scmp.ClassType{Class: scmp.C_General, Type: scmp.T_G_StochasticCongWarn}, info)
	if err != nil {
		log.Error("unable to create notification SCMP", "err", err)
		return
	}
	if logEnabledStoch {
		srcIA, _ := notification.SrcIA()
		srcHost, _ := notification.SrcHost()
		DstIA, _ := notification.DstIA()
		DstHost, _ := notification.DstHost()
		pub := qp.Rp.Ctx.Conf.BR.InternalAddr
		routerAddr := addr.HostFromIP(pub.IP)
		pld, _ := notification.Payload(false)
		l4, _ := notification.L4Hdr(false)
		log.Debug("New SCMP Notification", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "\n RtrAddr", routerAddr,
			"CurrBW", r.config.Queues[qp.QueueNo].GetTokenBucket().CurrBW, "Pkt ID", id,
			"\n L4", l4,
			"\n Congestion Warning", pld)
	}
	notification.Route()

}

func (r *Router) createStochSCMPNotification(qp *qosqueues.QPkt,
	ct scmp.ClassType, info scmp.Info) (*rpkt.RtrPkt, error, string) {

	if logEnabledStoch {
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
	sp.L4 = scmp.NewHdr(scmp.ClassType{Class: scmp.C_General, Type: scmp.T_G_StochasticCongWarn}, sp.Pld.Len())
	log.Debug("Created SPkt reply", "sp", sp, "Pkt ID", id)
	reply, err := qp.Rp.CreateReply(sp)
	if logEnabledStoch {
		srcIA, _ := reply.SrcIA()
		srcHost, _ := reply.SrcHost()
		DstIA, _ := reply.DstIA()
		DstHost, _ := reply.DstHost()
		log.Debug("Created RPkt reply", "SrcIA", srcIA, "SrcHost",
			srcHost, "DstIA", DstIA, "DstHost", DstHost, "Pkt ID", id)
	}
	return reply, err, id
}

func (r *Router) createStochCongWarn(np *qosqueues.NPkt) *scmp.InfoStochCW {
	restrictionPrint := r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning()
	testing := r.config.Queues[np.Qpkt.QueueNo].GetMinBandwidth()
	restriction := 3
	if logEnabledStoch {
		log.Debug("restrictions on information content", "restriction", restrictionPrint, "MinBW", testing)
	}
	if restriction > 3 {
		log.Error("Unable to create congestion warning", "restriction on information content", restriction)
		return nil
	}
	stochCW := &scmp.InfoStochCW{}
	stochCW.ConsIngress = common.IFIDType(np.Qpkt.Rp.Ingress.IfID)
	if logEnabledStoch {
		log.Debug("InfoBscCW", "ConsIngress", common.IFIDType(np.Qpkt.Rp.Ingress.IfID),
			"QueueLength", (r.config.Queues[np.Qpkt.QueueNo]).GetLength(), "CurrBW",
			r.config.Queues[np.Qpkt.QueueNo].GetTokenBucket().CurrBW, "QueueFullness",
			(r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel(), "Violation", np.Qpkt.Act.Reason)
	}
	if restriction > 0 {
		stochCW.QueueLength = uint64((r.config.Queues[np.Qpkt.QueueNo]).GetLength())
	}
	if restriction > 1 {
		stochCW.CurrBW = r.config.Queues[np.Qpkt.QueueNo].GetTokenBucket().CurrBW
		stochCW.QueueFullness = uint64((r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel())
	}
	if restriction > 2 {
		stochCW.Violation = uint64(np.Qpkt.Act.Reason)
		//stochCW.ClassRule = np.Qpkt.Act.Rule
	}
	return stochCW
}

func (r *Router) calculateProbs(qp *qosqueues.QPkt, switchingPoint int) int {
	//queueFullness := (r.config.Queues[qp.QueueNo]).GetFillLevel() //TODO: uncomment when queues fill up more realistically
	queueFullness := rand.Intn(100)
	if queueFullness <= switchingPoint {
		return queueFullness
	}
	return (queueFullness - 1) / (queueFullness - switchingPoint)
}
