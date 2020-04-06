//Implements the hop-by-hop notification approach

package main

import (
	"github.com/scionproto/scion/go/border/qosqueues"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/scmp"
)

const (
	logEnabledHbh = false
)

func (r *Router) hbhNotify() {
	for np := range r.notifications {
		//if r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning().Approach == 2 { //TODO: remove when congestionWarning fields are read out correctly
		//Check if the packet came from local AS? don't include in RASCA and create BscNotifyCW for it: include in RASCA
		// if np.Qpkt.Rp.DirFrom == 1 || np.Qpkt.Rp.DirFrom == 2 { //Q: not sure if first case is possible
		// bscCW := r.createBscCongWarn(np)
		// r.sendBscNotificationSCMP(np.Qpkt, bscCW)
		// np.Qpkt.Rp.RefInc(-1)
		// }
		hbhCW := r.createHbhCongWarn(np)
		hbhPktInfo := np.Qpkt.Rp.CreateHbhPktInfo(hbhCW)
		if logEnabledHbh {
			revPath := hbhPktInfo.RevPath
			infoField, _ := hbhPktInfo.RevPath.GetInfoField(hbhPktInfo.RevPath.InfOff)
			hopField, _ := hbhPktInfo.RevPath.GetHopField(hbhPktInfo.RevPath.HopOff)
			_ = hbhPktInfo.RevPath.IncOffsets()
			newRevPath := hbhPktInfo
			nextinfoField, _ := hbhPktInfo.RevPath.GetInfoField(hbhPktInfo.RevPath.InfOff)
			nextHopField, _ := hbhPktInfo.RevPath.GetHopField(hbhPktInfo.RevPath.HopOff)
			log.Debug("Reversed path", "RevPath", revPath,
				"\n InfoField", infoField, "\n HopField", hopField, "\n newRevPath", newRevPath,
				"\n nextInfoField",
				nextinfoField, "\n nextHopField", nextHopField)
		}
		np.Qpkt.Rp.RefInc(-1)
		//}
	}

}

//Creates congestion warning for one packet, these then have to be cumulated to get a hbh congestion warning
func (r *Router) createHbhCongWarn(np *qosqueues.NPkt) *scmp.InfoHbhCW {
	restriction := 3
	if logEnabledHbh {
		restrictionPrint := r.config.Queues[np.Qpkt.QueueNo].GetCongestionWarning()
		testing := r.config.Queues[np.Qpkt.QueueNo].GetMinBandwidth()
		log.Debug("restrictions on information content", "restriction", restrictionPrint, "MinBW", testing)
	}
	if restriction > 3 {
		log.Error("Unable to create congestion warning", "restriction on information content", restriction)
		return nil
	}
	hbhCW := &scmp.InfoHbhCW{}
	hbhCW.ConsIngress = common.IFIDType(np.Qpkt.Rp.Ingress.IfID)
	if logEnabledHbh {
		log.Debug("InfoBscCW", "ConsIngress", common.IFIDType(np.Qpkt.Rp.Ingress.IfID),
			"QueueLength", (r.config.Queues[np.Qpkt.QueueNo]).GetLength(), "CurrBW",
			r.config.Queues[np.Qpkt.QueueNo].GetTokenBucket().CurrBW, "QueueFullness",
			(r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel(), "Violation", np.Qpkt.Act.Reason)
	}
	if restriction > 0 {
		hbhCW.QueueLength = uint64((r.config.Queues[np.Qpkt.QueueNo]).GetLength())
	}
	if restriction > 1 {
		hbhCW.CurrBW = r.config.Queues[np.Qpkt.QueueNo].GetTokenBucket().CurrBW
		hbhCW.QueueFullness = uint64((r.config.Queues[np.Qpkt.QueueNo]).GetFillLevel())
	}
	if restriction > 2 {
		hbhCW.Violation = uint64(np.Qpkt.Act.Reason)
		//bscCW.ClassRule = np.Qpkt.Act.Rule
	}
	return hbhCW
}
