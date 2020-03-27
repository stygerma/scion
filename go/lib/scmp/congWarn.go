package scmp

//IMPL:

type congWarn interface {
	getCongWarn()
}

//MS: ================================= May be unnecessary
//type QueueCongState struct {
//	CurrBW uint64 /*TODO: find way to calculate this, has to be approximated over time interval, just calculate it in the refill function.
//	Add field to the tokenBucket struct that shows the latest approximation of the BW*/
//	QueueLength   uint64
//	QueueFullness uint16
// 	ConsIngress   common.IFIDType /*Q: Joel's policies should enforce some contracts between ISP and also be inmplemented such BRs are not overwhelmed, this means that we should
// 	somehow add all the IFs of this BR to this field or find a way to identify this BR such that path segments over this BR will be avoided in
// 	the future*/
// 	/*MS: just add the ingress interface that the packet took, this should enable an end host to find the path segments which use this interface. */
// 	Violation uint64 /* Include the Violation and not the whole classRule, saves space and should also give some useful info.
// 	classRules as defined by Joel, this would introduce big space overhead, however, there are no real incentives for an ISP to share this.
// 	TODO: should be exportable, i.e. write getter or tell Joel to define the type classrule by writing in uppercase */
// }
// */
