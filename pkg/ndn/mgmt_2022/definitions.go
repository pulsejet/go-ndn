//go:generate gondn_tlv_gen
package mgmt_2022

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

const (
	FaceScopeNonLocal = uint64(0)
	FaceScopeLocal    = uint64(1)
)

const (
	FacePersPersistent = uint64(0)
	FacePersOnDemand   = uint64(1)
	FacePersPermanent  = uint64(2)
)

const (
	FaceLinkPointToPoint = uint64(0)
	FaceLinkMultiAccess  = uint64(1)
	FaceLinkAdHoc        = uint64(2)
)

const (
	FaceFlagNoFlag                   = uint64(0)
	FaceFlagLocalFieldsEnabled       = uint64(1)
	FaceFlagLpReliabilityEnabled     = uint64(2)
	FaceFlagCongestionMarkingEnabled = uint64(4)
)

const (
	RouteFlagNoFlag       = uint64(0)
	RouteFlagChildInherit = uint64(1)
	RouteFlagCapture      = uint64(2)
)

const (
	FaceEventCreated   = uint64(1)
	FaceEventDestroyed = uint64(2)
	FaceEventUp        = uint64(3)
	FaceEventDown      = uint64(4)
)

const (
	CsFlagNone    = uint64(0)
	CsEnableAdmit = uint64(1)
	CsEnableServe = uint64(2)
)

// +tlv-model:dict
type Strategy struct {
	//+field:name
	Name enc.Name `tlv:"0x07"`
}

// +tlv-model:dict
type ControlArgs struct {
	// Note: go-ndn generator does not support inheritance yet.

	//+field:name
	Name enc.Name `tlv:"0x07"`
	//+field:natural:optional
	FaceId *uint64 `tlv:"0x69"`
	//+field:string:optional
	Uri *string `tlv:"0x72"`
	//+field:string:optional
	LocalUri *string `tlv:"0x81"`
	//+field:natural:optional
	Origin *uint64 `tlv:"0x6f"`
	//+field:natural:optional
	Cost *uint64 `tlv:"0x6a"`
	//+field:natural:optional
	Capacity *uint64 `tlv:"0x83"`
	//+field:natural:optional
	Count *uint64 `tlv:"0x84"`
	//+field:natural:optional
	BaseCongestionMarkInterval *uint64 `tlv:"0x87"`
	//+field:natural:optional
	DefaultCongestionThreshold *uint64 `tlv:"0x88"`
	//+field:natural:optional
	Mtu *uint64 `tlv:"0x89"`
	//+field:natural:optional
	Flags *uint64 `tlv:"0x6c"`
	//+field:natural:optional
	Mask *uint64 `tlv:"0x70"`
	//+field:struct:Strategy
	Strategy *Strategy `tlv:"0x6b"`
	//+field:natural:optional
	ExpirationPeriod *uint64 `tlv:"0x6d"`
	//+field:natural:optional
	FacePersistency *uint64 `tlv:"0x85"`
}

// +tlv-model:dict
type ControlResponseVal struct {
	//+field:natural
	StatusCode uint64 `tlv:"0x66"`
	//+field:string
	StatusText string `tlv:"0x67"`
	//+field:struct:ControlArgs
	Params *ControlArgs `tlv:"0x68"`
}

type ControlParameters struct {
	//+field:struct:ControlArgs
	Val *ControlArgs `tlv:"0x68"`
}

type ControlResponse struct {
	//+field:struct:ControlResponseVal
	Val *ControlResponseVal `tlv:"0x65"`
}

type FaceEventNotificationValue struct {
	//+field:natural
	FaceEventKind uint64 `tlv:"0xc1"`
	//+field:natural
	FaceId uint64 `tlv:"0x69"`
	//+field:string
	Uri string `tlv:"0x72"`
	//+field:string
	LocalUri string `tlv:"0x81"`
	//+field:natural
	FaceScope uint64 `tlv:"0x84"`
	//+field:natural
	FacePersistency uint64 `tlv:"0x85"`
	//+field:natural
	LinkType uint64 `tlv:"0x86"`
	//+field:natural
	Flags uint64 `tlv:"0x6c"`
}

type FaceEventNotification struct {
	//+field:struct:FaceEventNotificationValue
	Val *FaceEventNotificationValue `tlv:"0xc0"`
}

type GeneralStatus struct {
	//+field:string
	NfdVersion string `tlv:"0x80"`
	//+field:natural
	StartTimestamp uint64 `tlv:"0x81"`
	//+field:natural
	CurrentTimestamp uint64 `tlv:"0x82"`
	//+field:natural
	NNameTreeEntries uint64 `tlv:"0x83"`
	//+field:natural
	NFibEntries uint64 `tlv:"0x84"`
	//+field:natural
	NPitEntries uint64 `tlv:"0x85"`
	//+field:natural
	NMeasurementsEntries uint64 `tlv:"0x86"`
	//+field:natural
	NCsEntries uint64 `tlv:"0x87"`
	//+field:natural
	NInInterests uint64 `tlv:"0x90"`
	//+field:natural
	NInData uint64 `tlv:"0x91"`
	//+field:natural
	NInNacks uint64 `tlv:"0x97"`
	//+field:natural
	NOutInterests uint64 `tlv:"0x92"`
	//+field:natural
	NOutData uint64 `tlv:"0x93"`
	//+field:natural
	NOutNacks uint64 `tlv:"0x98"`
	//+field:natural
	NSatisfiedInterests uint64 `tlv:"0x99"`
	//+field:natural
	NUnsatisfiedInterests uint64 `tlv:"0x9a"`

	//+field:natural:optional
	NFragmentationError *uint64 `tlv:"0xc8"`
	//+field:natural:optional
	NOutOverMtu *uint64 `tlv:"0xc9"`
	//+field:natural:optional
	NInLpInvalid *uint64 `tlv:"0xca"`
	//+field:natural:optional
	NReassemblyTimeouts *uint64 `tlv:"0xcb"`
	//+field:natural:optional
	NInNetInvalid *uint64 `tlv:"0xcc"`
	//+field:natural:optional
	NAcknowledged *uint64 `tlv:"0xcd"`
	//+field:natural:optional
	NRetransmitted *uint64 `tlv:"0xce"`
	//+field:natural:optional
	NRetxExhausted *uint64 `tlv:"0xcf"`
	//+field:natural:optional
	NConngestionMarked *uint64 `tlv:"0xd0"`
}

type FaceStatus struct {
	//+field:natural
	FaceId uint64 `tlv:"0x69"`
	//+field:string
	Uri string `tlv:"0x72"`
	//+field:string
	LocalUri string `tlv:"0x81"`
	//+field:natural:optional
	ExpirationPeriod *uint64 `tlv:"0x6d"`
	//+field:natural
	FaceScope uint64 `tlv:"0x84"`
	//+field:natural
	FacePersistency uint64 `tlv:"0x85"`
	//+field:natural
	LinkType uint64 `tlv:"0x86"`
	//+field:natural:optional
	BaseCongestionMarkInterval *uint64 `tlv:"0x87"`
	//+field:natural:optional
	DefaultCongestionThreshold *uint64 `tlv:"0x88"`
	//+field:natural:optional
	Mtu *uint64 `tlv:"0x89"`

	//+field:natural
	NInInterests uint64 `tlv:"0x90"`
	//+field:natural
	NInData uint64 `tlv:"0x91"`
	//+field:natural
	NInNacks uint64 `tlv:"0x97"`
	//+field:natural
	NOutInterests uint64 `tlv:"0x92"`
	//+field:natural
	NOutData uint64 `tlv:"0x93"`
	//+field:natural
	NOutNacks uint64 `tlv:"0x98"`
	//+field:natural
	NInBytes uint64 `tlv:"0x94"`
	//+field:natural
	NOutBytes uint64 `tlv:"0x95"`

	//+field:natural
	Flags uint64 `tlv:"0x6c"`
}

type FaceStatusMsg struct {
	//+field:sequence:*FaceStatus:struct:FaceStatus
	Vals []*FaceStatus `tlv:"0x80"`
}

type FaceQueryFilterValue struct {
	//+field:natural:optional
	FaceId *uint64 `tlv:"0x69"`
	//+field:string:optional
	UriScheme *string `tlv:"0x83"`
	//+field:string:optional
	Uri *string `tlv:"0x72"`
	//+field:string:optional
	LocalUri *string `tlv:"0x81"`
	//+field:natural:optional
	FaceScope *uint64 `tlv:"0x84"`
	//+field:natural:optional
	FacePersistency *uint64 `tlv:"0x85"`
	//+field:natural:optional
	LinkType *uint64 `tlv:"0x86"`
}

type FaceQueryFilter struct {
	//+field:struct:FaceQueryFilterValue
	Val *FaceQueryFilterValue `tlv:"0x96"`
}

type Route struct {
	//+field:natural
	FaceId uint64 `tlv:"0x69"`
	//+field:natural
	Origin uint64 `tlv:"0x6f"`
	//+field:natural
	Cost uint64 `tlv:"0x6a"`
	//+field:natural
	Flags uint64 `tlv:"0x6c"`
	//+field:natural:optional
	ExpirationPeriod *uint64 `tlv:"0x6d"`
}

type RibEntry struct {
	//+field:name
	Name enc.Name `tlv:"0x07"`
	//+field:sequence:*Route:struct:Route
	Routes []*Route `tlv:"0x81"`
}

type RibStatus struct {
	//+field:sequence:*RibEntry:struct:RibEntry
	Entries []*RibEntry `tlv:"0x80"`
}

type NextHopRecord struct {
	//+field:natural
	FaceId uint64 `tlv:"0x69"`
	//+field:natural
	Cost uint64 `tlv:"0x6a"`
}

type FibEntry struct {
	//+field:name
	Name enc.Name `tlv:"0x07"`
	//+field:sequence:*NextHopRecord:struct:NextHopRecord
	NextHopRecords []*NextHopRecord `tlv:"0x81"`
}

type FibStatus struct {
	//+field:sequence:*FibEntry:struct:FibEntry
	Entries []*FibEntry `tlv:"0x80"`
}

type StrategyChoice struct {
	//+field:name
	Name enc.Name `tlv:"0x07"`
	//+field:struct:Strategy
	Strategy *Strategy `tlv:"0x6b"`
}

type StrategyChoiceMsg struct {
	//+field:sequence:*StrategyChoice:struct:StrategyChoice
	StrategyChoices []*StrategyChoice `tlv:"0x80"`
}

// Not supported by NFD yet
type CsInfo struct {
	//+field:natural
	Capacity uint64 `tlv:"0x83"`
	//+field:natural
	Flags uint64 `tlv:"0x6c"`
	//+field:natural
	NCsEntries uint64 `tlv:"0x87"`
	//+field:natural
	NHits uint64 `tlv:"0x81"`
	//+field:natural
	NMisses uint64 `tlv:"0x82"`
}

// No Tlv numbers assigned yet
type CsQuery struct {
	Name            enc.Name
	PacketSize      uint64
	FreshnessPeriod uint64
}
