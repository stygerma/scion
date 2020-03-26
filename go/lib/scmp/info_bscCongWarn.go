package scmp

//IMPL: Defines the new SCMP type used for basic information dissemination

type BscCongWarn struct {
	//timestamp uint32 TBA: should not be needed as we can just look at the timestamp from the packet itself
	ifInfo ifCongState //MS: need to implement something such that the interface state is defined while the ISP can restrict what is shared
}
