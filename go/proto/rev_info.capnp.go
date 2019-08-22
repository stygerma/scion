// Code generated by capnpc-go. DO NOT EDIT.

package proto

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type RevInfo struct{ capnp.Struct }

// RevInfo_TypeID is the unique identifier for the type RevInfo.
const RevInfo_TypeID = 0xe40561cf10a34bc8

func NewRevInfo(s *capnp.Segment) (RevInfo, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 0})
	return RevInfo{st}, err
}

func NewRootRevInfo(s *capnp.Segment) (RevInfo, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 0})
	return RevInfo{st}, err
}

func ReadRootRevInfo(msg *capnp.Message) (RevInfo, error) {
	root, err := msg.RootPtr()
	return RevInfo{root.Struct()}, err
}

func (s RevInfo) String() string {
	str, _ := text.Marshal(0xe40561cf10a34bc8, s.Struct)
	return str
}

func (s RevInfo) IfID() uint64 {
	return s.Struct.Uint64(0)
}

func (s RevInfo) SetIfID(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s RevInfo) Isdas() uint64 {
	return s.Struct.Uint64(8)
}

func (s RevInfo) SetIsdas(v uint64) {
	s.Struct.SetUint64(8, v)
}

func (s RevInfo) LinkType() LinkType {
	return LinkType(s.Struct.Uint16(16))
}

func (s RevInfo) SetLinkType(v LinkType) {
	s.Struct.SetUint16(16, uint16(v))
}

func (s RevInfo) Timestamp() uint32 {
	return s.Struct.Uint32(20)
}

func (s RevInfo) SetTimestamp(v uint32) {
	s.Struct.SetUint32(20, v)
}

func (s RevInfo) Ttl() uint32 {
	return s.Struct.Uint32(24)
}

func (s RevInfo) SetTtl(v uint32) {
	s.Struct.SetUint32(24, v)
}

// RevInfo_List is a list of RevInfo.
type RevInfo_List struct{ capnp.List }

// NewRevInfo creates a new list of RevInfo.
func NewRevInfo_List(s *capnp.Segment, sz int32) (RevInfo_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 32, PointerCount: 0}, sz)
	return RevInfo_List{l}, err
}

func (s RevInfo_List) At(i int) RevInfo { return RevInfo{s.List.Struct(i)} }

func (s RevInfo_List) Set(i int, v RevInfo) error { return s.List.SetStruct(i, v.Struct) }

func (s RevInfo_List) String() string {
	str, _ := text.MarshalList(0xe40561cf10a34bc8, s.List)
	return str
}

// RevInfo_Promise is a wrapper for a RevInfo promised by a client call.
type RevInfo_Promise struct{ *capnp.Pipeline }

func (p RevInfo_Promise) Struct() (RevInfo, error) {
	s, err := p.Pipeline.Struct()
	return RevInfo{s}, err
}

const schema_c434abcc856ab808 = "x\xda4\xc81K\xc3@\x1c\x86\xf1\xf7\xbd\xbb4\xe9" +
	"\x94\x1e\xc4\xc1A\x0a\x8e\x0e\x82\xe2\xe4\xa2\x83Ku\xf1" +
	"_\xdcK\xb0\x09D\xd34\x98P\x14t\x10Zp\xa8" +
	"P\x07\xa1\xa0\x9b\xa3_@tup\x10\xfc\x12~\x04" +
	"\xf7\x93\x0e\xdd\x9e\xe7\xd7\x8a\xf7\x95\xf5z\x80\x18\xaf\xe1" +
	"\xbe\x8e^Z?\xb1\xf7\x0b\x09i\\\xf0v6\xf9~" +
	"\xdd\xf9\x84\xf1\x01\xbb\xf2n\xd7|`k\xb5M\xd0]" +
	"$\xa3^V\xa4Cn\x9e\xc6eQ\xeev\xf7\x92Q" +
	"\xa7H\x87\xc7\xa4D\xda\x00\x86\x80\xbd\xd9\x00\xe4RS" +
	"\xc6\x8a\x96\x8c\xb8\xc0\xdbm@\xae5\xe5N\xd1\xaa " +
	"\xa2\x02\xec\xe4\x10\x90\xb1\xa6\xcc\x14\xad\xf6\"j\xc0\xde" +
	"w\x01\x99j\xca\\\xd1\x9aFD\x03\xd8\xc7u@f" +
	"\x9a\xf2\xac\x18fi\xe7\x80M(6\xc1vV\xf5\xe3" +
	"jy.\xcf\x8a\xf3\x93\xab2\x01\xc0\xd0\xf5?\x9e\xa6" +
	"\x7f\xf3\xfc\x01 C\xd0\xd5\xd9 \xa9\xeax\x00\x96\x0c" +
	"\xa0\x18\x80~]\xe7\xcb\xfe\x0f\x00\x00\xff\xff s9" +
	"\xea"

func init() {
	schemas.Register(schema_c434abcc856ab808,
		0xe40561cf10a34bc8)
}
