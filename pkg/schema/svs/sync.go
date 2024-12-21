package svs

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	"github.com/zjkmxy/go-ndn/pkg/log"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
	stlv "github.com/zjkmxy/go-ndn/pkg/ndn/svs_2024"
	"github.com/zjkmxy/go-ndn/pkg/schema"
	"github.com/zjkmxy/go-ndn/pkg/utils"
)

type SyncState int

type MissingData struct {
	NodeId   enc.Name
	StartSeq uint64
	EndSeq   uint64
}

const (
	SyncSteady SyncState = iota
	SyncSupression
)

// SvsNode implements the StateVectorSync but works for only one instance.
// Similar is RegisterPolicy. A better implementation is needed if there is
// a need that multiple producers under the same name pattern that runs on the same application instance.
// It would also be more natural if we make 1-1 mapping between MatchedNodes and SVS instances,
// instead of the Node and the SVS instance, which is against the philosophy of matching.
// Also, this sample always starts from sequence number 0.
type SvsNode struct {
	schema.BaseNodeImpl

	OnMissingData *schema.EventTarget

	SyncInterval        time.Duration
	SuppressionInterval time.Duration
	BaseMatching        enc.Matching
	ChannelSize         uint64
	SelfNodeId          enc.Name

	dataLock        sync.Mutex
	timer           ndn.Timer
	cancelSyncTimer func() error
	missChan        chan MissingData
	stopChan        chan struct{}

	localSv   stlv.StateVector
	aggSv     stlv.StateVector
	state     SyncState
	selfSeq   uint64
	ownPrefix enc.Name
	notifNode *schema.Node
}

func (n *SvsNode) NodeImplTrait() schema.NodeImpl {
	return n
}

func CreateSvsNode(node *schema.Node) schema.NodeImpl {
	ret := &SvsNode{
		BaseNodeImpl: schema.BaseNodeImpl{
			Node:        node,
			OnAttachEvt: &schema.EventTarget{},
			OnDetachEvt: &schema.EventTarget{},
		},
		OnMissingData:       &schema.EventTarget{},
		BaseMatching:        enc.Matching{},
		SyncInterval:        30 * time.Second,
		SuppressionInterval: 200 * time.Millisecond,
	}

	path, _ := enc.NamePatternFromStr("/<8=nodeId>/<seq=seqNo>")
	leafNode := node.PutNode(path, schema.LeafNodeDesc)
	leafNode.Set(schema.PropCanBePrefix, false)
	leafNode.Set(schema.PropMustBeFresh, false)
	leafNode.Set(schema.PropLifetime, 4*time.Second)
	leafNode.Set(schema.PropFreshness, 60*time.Second)
	leafNode.Set("ValidDuration", 876000*time.Hour)

	path, _ = enc.NamePatternFromStr("/32=notif")
	ret.notifNode = node.PutNode(path, schema.ExpressPointDesc)
	ret.notifNode.Set(schema.PropCanBePrefix, true)
	ret.notifNode.Set(schema.PropMustBeFresh, true)
	ret.notifNode.Set(schema.PropLifetime, 1*time.Second)
	ret.notifNode.AddEventListener(schema.PropOnInterest, utils.IdPtr(ret.onSyncInt))

	ret.BaseMatching = enc.Matching{}
	ret.OnAttachEvt.Add(utils.IdPtr(ret.onAttach))
	ret.OnDetachEvt.Add(utils.IdPtr(ret.onDetach))

	return ret
}

func findSvsEntry(v *stlv.StateVector, nodeId enc.Name) int {
	// This is less efficient but enough for a demo.
	for i, n := range v.Entries {
		if nodeId.Equal(n.NodeId) {
			return i
		}
	}
	return -1
}

func (n *SvsNode) onSyncInt(event *schema.Event) any {
	mNotifNode := event.Target
	logger := mNotifNode.Logger("SvsNode") // the path will be the subchild
	remoteSv, err := stlv.ParseStateVector(enc.NewWireReader(event.Content), true)
	if err != nil {
		logger.Error("Unable to parse state vector. Drop.")
	}

	// If append() is called on localSv slice, a lock is necessary
	n.dataLock.Lock()
	defer n.dataLock.Unlock()

	// Compare state vectors
	// needFetch := false
	needNotif := false
	for _, cur := range remoteSv.Entries {
		li := findSvsEntry(&n.localSv, cur.NodeId)
		if li == -1 {
			n.localSv.Entries = append(n.localSv.Entries, &stlv.StateVectorEntry{
				NodeId: cur.NodeId,
				SeqNo:  cur.SeqNo,
			})
			// needFetch = true
			n.missChan <- MissingData{
				NodeId:   cur.NodeId,
				StartSeq: 1,
				EndSeq:   cur.SeqNo + 1,
			}
		} else if n.localSv.Entries[li].SeqNo < cur.SeqNo {
			log.Debugf("Missing data for: [%d]: %d < %d", cur.NodeId, n.localSv.Entries[li].SeqNo, cur.SeqNo)
			n.missChan <- MissingData{
				NodeId:   cur.NodeId,
				StartSeq: n.localSv.Entries[li].SeqNo + 1,
				EndSeq:   cur.SeqNo + 1,
			}
			n.localSv.Entries[li].SeqNo = cur.SeqNo
			// needFetch = true
		} else if n.localSv.Entries[li].SeqNo > cur.SeqNo {
			log.Debugf("Outdated remote on: [%d]: %d < %d", cur.NodeId, cur.SeqNo, n.localSv.Entries[li].SeqNo)
			needNotif = true
		}
	}
	for _, cur := range n.localSv.Entries {
		li := findSvsEntry(remoteSv, cur.NodeId)
		if li == -1 {
			needNotif = true
		}
	}
	// Notify the callback coroutine if applicable
	// if needFetch {
	// 	select {
	// 	case n.sigChan <- struct{}{}:
	// 	default:
	// 	}
	// }
	// Set sync state if applicable
	// if needNotif {
	// 	n.aggregate(remoteSv)
	// 	if n.state == SyncSteady {
	// 		n.transitToSuppress(remoteSv)
	// 	}
	// }
	// TODO: Have trouble understanding this mechanism from the Spec.
	// From StateVectorSync Spec 4.4,
	// "Incoming Sync Interest is outdated: Node moves to Suppression State."
	// implies the state becomes Supression State when `remote any< local`
	// From StateVectorSync Spec 6, the box below
	// "local_state_vector any< x"
	// implies the state becomes Supression State when `local any< remote`
	// Contradiction. The wrong one should be the figure.
	// Since supression is an optimization that does not affect the demo, ignore for now.
	// Report this issue to the team when have time.

	if needNotif || n.state == SyncSupression {
		// Set the aggregation timer
		if n.state == SyncSteady {
			n.state = SyncSupression
			n.aggSv = stlv.StateVector{Entries: make([]*stlv.StateVectorEntry, len(remoteSv.Entries))}
			copy(n.aggSv.Entries, remoteSv.Entries)
			n.cancelSyncTimer()
			n.cancelSyncTimer = n.timer.Schedule(n.getAggIntv(), n.onSyncTimer)
		} else {
			// Should aggregate the incoming sv first, and only shoot after sync timer.
			n.aggregate(remoteSv)
		}
	} else {
		// Reset the sync timer (already in lock)
		n.cancelSyncTimer()
		n.cancelSyncTimer = n.timer.Schedule(n.getSyncIntv(), n.onSyncTimer)
	}

	return true
}

func (n *SvsNode) MissingDataChannel() chan MissingData {
	// Note: DO NOT use with OnMissingData
	return n.missChan
}

func (n *SvsNode) MySequence() uint64 {
	return n.selfSeq
}

func (n *SvsNode) aggregate(remoteSv *stlv.StateVector) {
	for _, cur := range remoteSv.Entries {
		li := findSvsEntry(&n.aggSv, cur.NodeId)
		if li == -1 {
			n.aggSv.Entries = append(n.aggSv.Entries, &stlv.StateVectorEntry{
				NodeId: cur.NodeId,
				SeqNo:  cur.SeqNo,
			})
		} else {
			n.aggSv.Entries[li].SeqNo = utils.Max(n.aggSv.Entries[li].SeqNo, cur.SeqNo)
		}
	}
}

func (n *SvsNode) onSyncTimer() {
	n.dataLock.Lock()
	defer n.dataLock.Unlock()
	// If in supression state, first test necessity
	notNecessary := false
	if n.state == SyncSupression {
		n.state = SyncSteady
		notNecessary = true
		for _, cur := range n.localSv.Entries {
			li := findSvsEntry(&n.aggSv, cur.NodeId)
			if li == -1 || n.aggSv.Entries[li].SeqNo < cur.SeqNo {
				notNecessary = false
				break
			}
		}
	}
	if !notNecessary {
		n.expressStateVec()
	}
	// In case a new one is just scheduled by the onInterest callback. No-op most of the case.
	n.cancelSyncTimer()
	n.cancelSyncTimer = n.timer.Schedule(n.getSyncIntv(), n.onSyncTimer)
}

func (n *SvsNode) expressStateVec() {
	n.notifNode.Apply(n.BaseMatching).Call("NeedChan", n.localSv.Encode())
}

func (n *SvsNode) getSyncIntv() time.Duration {
	dev := rand.Int63n(n.SyncInterval.Nanoseconds()/4) - n.SyncInterval.Nanoseconds()/8
	return n.SyncInterval + time.Duration(dev)*time.Nanosecond
}

func (n *SvsNode) getAggIntv() time.Duration {
	dev := rand.Int63n(n.SuppressionInterval.Nanoseconds()) - n.SuppressionInterval.Nanoseconds()/2
	return n.SuppressionInterval + time.Duration(dev)*time.Nanosecond
}

func (n *SvsNode) NewData(mNode schema.MatchedNode, content enc.Wire) enc.Wire {
	n.dataLock.Lock()
	defer n.dataLock.Unlock()

	logger := mNode.Logger("SvsNode")

	n.selfSeq++
	newDataName := make(enc.Name, len(n.ownPrefix)+1)
	copy(newDataName, n.ownPrefix)
	newDataName[len(n.ownPrefix)] = enc.NewSequenceNumComponent(n.selfSeq)
	mLeafNode := mNode.Refine(newDataName)
	ret := mLeafNode.Call("Provide", content).(enc.Wire)
	if len(ret) > 0 {
		li := findSvsEntry(&n.localSv, n.SelfNodeId)
		if li >= 0 {
			n.localSv.Entries[li].SeqNo = n.selfSeq
		}
		n.state = SyncSteady
		logger.Debugf("NewData generated w/ seq=%d", n.selfSeq)
		n.expressStateVec()
	} else {
		logger.Errorf("Failed to provide seq=%d", n.selfSeq)
		n.selfSeq--
	}
	return ret
}

func (n *SvsNode) onAttach(event *schema.Event) any {
	if n.ChannelSize == 0 || len(n.SelfNodeId) == 0 ||
		n.BaseMatching == nil || n.SyncInterval <= 0 || n.SuppressionInterval <= 0 {
		panic(errors.New("SvsNode: not configured before Init"))
	}

	n.timer = event.TargetNode.Engine().Timer()
	n.dataLock = sync.Mutex{}
	n.dataLock.Lock()
	defer n.dataLock.Unlock()

	n.ownPrefix = event.TargetNode.Apply(n.BaseMatching).Name
	n.ownPrefix = append(n.ownPrefix, n.SelfNodeId...)

	// OnMissingData callback

	n.localSv = stlv.StateVector{Entries: make([]*stlv.StateVectorEntry, 0)}
	n.aggSv = stlv.StateVector{Entries: make([]*stlv.StateVectorEntry, 0)}
	// n.onMiss = schema.NewEvent[*SvsOnMissingEvent]()
	n.state = SyncSteady
	n.missChan = make(chan MissingData, n.ChannelSize)
	// The first sync Interest should be sent out ASAP
	n.cancelSyncTimer = n.timer.Schedule(utils.Min(n.getSyncIntv(), 100*time.Millisecond), n.onSyncTimer)

	n.stopChan = make(chan struct{}, 1)
	if len(n.OnMissingData.Val()) > 0 {
		go n.callbackRoutine()
	}

	// initialize localSv
	// TODO: this demo does not consider recovery from off-line. Should be done via ENV and storage policy.
	n.localSv.Entries = append(n.localSv.Entries, &stlv.StateVectorEntry{
		NodeId: n.SelfNodeId,
		SeqNo:  0,
	})
	n.selfSeq = 0
	return nil
}

func (n *SvsNode) onDetach(event *schema.Event) any {
	n.dataLock.Lock()
	defer n.dataLock.Unlock()
	n.cancelSyncTimer()
	close(n.missChan)
	n.stopChan <- struct{}{}
	close(n.stopChan)
	return nil
}

func (n *SvsNode) callbackRoutine() {
	panic("TODO: TO BE DONE")
}

func (n *SvsNode) GetDataName(mNode schema.MatchedNode, nodeId []byte, seq uint64) enc.Name {
	ret := make(enc.Name, len(mNode.Name)+2)
	copy(ret, mNode.Name)
	ret[len(mNode.Name)] = enc.Component{Typ: enc.TypeGenericNameComponent, Val: nodeId}
	ret[len(mNode.Name)+1] = enc.NewSequenceNumComponent(seq)
	return ret
}

func (n *SvsNode) CastTo(ptr any) any {
	switch ptr.(type) {
	case (*SvsNode):
		return n
	case (*schema.BaseNodeImpl):
		return &(n.BaseNodeImpl)
	default:
		return nil
	}
}

var SvsNodeDesc *schema.NodeImplDesc

func init() {
	SvsNodeDesc = &schema.NodeImplDesc{
		ClassName: "SvsNode",
		Properties: map[schema.PropKey]schema.PropertyDesc{
			"SyncInterval":        schema.TimePropertyDesc("SyncInterval"),
			"SuppressionInterval": schema.TimePropertyDesc("SuppressionInterval"),
			"BaseMatching":        schema.MatchingPropertyDesc("BaseMatching"),
			"ChannelSize":         schema.DefaultPropertyDesc("ChannelSize"),
			"SelfNodeId":          schema.DefaultPropertyDesc("SelfNodeId"),
			"ContentType":         schema.SubNodePropertyDesc("/<8=nodeId>/<seq=seqNo>", "ContentType"),
			"Lifetime":            schema.SubNodePropertyDesc("/<8=nodeId>/<seq=seqNo>", "Lifetime"),
			"Freshness":           schema.SubNodePropertyDesc("/<8=nodeId>/<seq=seqNo>", "Freshness"),
			"ValidDuration":       schema.SubNodePropertyDesc("/<8=nodeId>/<seq=seqNo>", "ValidDuration"),
			"MustBeFresh":         schema.SubNodePropertyDesc("/<8=nodeId>/<seq=seqNo>", "MustBeFresh"),
		},
		Events: map[schema.PropKey]schema.EventGetter{
			schema.PropOnAttach: schema.DefaultEventTarget(schema.PropOnAttach),
			schema.PropOnDetach: schema.DefaultEventTarget(schema.PropOnDetach),
			"OnMissingData":     schema.DefaultEventTarget("OnMissingData"),
		},
		Functions: map[string]schema.NodeFunc{
			"NewData": func(mNode schema.MatchedNode, args ...any) any {
				if len(args) != 1 {
					err := fmt.Errorf("SvsNode.NewData requires 1 arguments but got %d", len(args))
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				content, ok := args[0].(enc.Wire)
				if !ok && args[0] != nil {
					err := ndn.ErrInvalidValue{Item: "content", Value: args[0]}
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				return schema.QueryInterface[*SvsNode](mNode.Node).NewData(mNode, content)
			},
			"MissingDataChannel": func(mNode schema.MatchedNode, args ...any) any {
				if len(args) != 0 {
					err := fmt.Errorf("SvsNode.MissingDataChannel requires 0 arguments but got %d", len(args))
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				return schema.QueryInterface[*SvsNode](mNode.Node).MissingDataChannel()
			},
			"MySequence": func(mNode schema.MatchedNode, args ...any) any {
				if len(args) != 0 {
					err := fmt.Errorf("SvsNode.MySequence requires 0 arguments but got %d", len(args))
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				return schema.QueryInterface[*SvsNode](mNode.Node).MySequence()
			},
			"GetDataName": func(mNode schema.MatchedNode, args ...any) any {
				if len(args) != 2 {
					err := fmt.Errorf("SvsNode.GetDataName requires 2 arguments but got %d", len(args))
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				nodeId, ok := args[0].([]byte)
				if !ok && args[0] != nil {
					err := ndn.ErrInvalidValue{Item: "nodeId", Value: args[0]}
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				seq, ok := args[1].(uint64)
				if !ok && args[1] != nil {
					err := ndn.ErrInvalidValue{Item: "seq", Value: args[1]}
					mNode.Logger("SvsNode").Error(err.Error())
					return err
				}
				return schema.QueryInterface[*SvsNode](mNode.Node).GetDataName(mNode, nodeId, seq)
			},
		},
		Create: CreateSvsNode,
	}
	schema.RegisterNodeImpl(SvsNodeDesc)
}
