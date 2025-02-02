package schema

import (
	"errors"
	"sync"
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	"github.com/zjkmxy/go-ndn/pkg/log"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
)

// Tree represents an NTSchema tree.
// It is supposed to be a static knowledge and shared by all parties in the system at compile time.
// The execution order: construct the tree -> apply policies & env setup -> attach to engine
type Tree struct {
	Root NTNode
	lock sync.RWMutex

	Engine ndn.Engine
}

// Attach the tree to the engine at prefix
func (t *Tree) Attach(prefix enc.Name, engine ndn.Engine) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	err := t.Root.SetAttachedPrefix(prefix)
	if err != nil {
		return err
	}
	path := make(enc.NamePattern, len(prefix))
	for i, c := range prefix {
		path[i] = c
	}
	err = t.Root.OnAttach(path, engine)
	if err != nil {
		return err
	}
	err = engine.AttachHandler(prefix, t.intHandler)
	if err != nil {
		return err
	}
	log.WithField("module", "schema").Info("Attached to engine.")
	t.Engine = engine
	return nil
}

// Detach the schema tree from the engine
func (t *Tree) Detach() {
	if t.Engine == nil {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	log.WithField("module", "schema").Info("Detached from engine")
	t.Engine.DetachHandler(t.Root.AttachedPrefix())
	t.Root.OnDetach()
}

// Match an NDN name to a (variable) matching
func (t *Tree) Match(name enc.Name) (NTNode, enc.Matching) {
	return t.Root.Match(name)
}

func (t *Tree) intHandler(
	interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire,
	reply ndn.ReplyFunc, deadline time.Time,
) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	matchName := interest.Name()
	extraComp := enc.Component{}
	if matchName[len(matchName)-1].Typ == enc.TypeParametersSha256DigestComponent ||
		matchName[len(matchName)-1].Typ == enc.TypeImplicitSha256DigestComponent {
		extraComp = matchName[len(matchName)-1]
		matchName = matchName[:len(matchName)-1]
	}
	node, matching := t.Root.Match(matchName)
	if node == nil {
		log.WithField("module", "schema").WithField("name", interest.Name().String()).Warn("Unexpected Interest. Drop.")
		return
	}
	if extraComp.Typ != enc.TypeInvalidComponent {
		switch extraComp.Typ {
		case enc.TypeParametersSha256DigestComponent:
			matching["params-sha256"] = extraComp.Val
		case enc.TypeImplicitSha256DigestComponent:
			matching["sha256digest"] = extraComp.Val
		}
	}
	node.OnInterest(interest, rawInterest, sigCovered, reply, deadline, matching)
}

// At the path return the node. Path does not include the attached prefix.
func (t *Tree) At(path enc.NamePattern) NTNode {
	return t.Root.At(path)
}

// PutNode puts the specified node at the specified path. Path does not include the attached prefix.
func (t *Tree) PutNode(path enc.NamePattern, node NTNode) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if len(path) == 0 {
		if t.Root == nil {
			t.Root = node
			return nil
		} else {
			return errors.New("schema node already exists")
		}
	} else {
		if t.Root == nil {
			t.Root = &BaseNode{}
			t.Root.Init(nil, nil)
		}
		return t.Root.PutNode(path, node)
	}
}

// RLock locks the tree for read use
func (t *Tree) RLock() {
	t.lock.RLock()
}

// RUnlock unlocks the tree locked by RLock
func (t *Tree) RUnlock() {
	t.lock.RUnlock()
}
