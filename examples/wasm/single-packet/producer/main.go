//go:build js && wasm

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	basic_engine "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	"github.com/zjkmxy/go-ndn/pkg/log"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
	"github.com/zjkmxy/go-ndn/pkg/schema"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
)

const SchemaJson = `{
  "nodes": {
    "/randomData/<v=time>": {
      "type": "LeafNode",
      "attrs": {
        "CanBePrefix": false,
        "MustBeFresh": true,
        "Lifetime": 6000
      },
      "events": {
        "OnInterest": ["$onInterest"]
      }
    }
  },
  "policies": [
    {
      "type": "RegisterPolicy",
      "path": "/",
      "attrs": {
        "RegisterIf": "$isProducer"
      }
    },
    {
      "type": "Sha256Signer",
      "path": "/randomData/<v=time>",
      "attrs": {}
    }
  ]
}`

func passAll(enc.Name, enc.Wire, ndn.Signature) bool {
	return true
}

func onInterest(event *schema.Event) any {
	mNode := event.Target
	timestamp, _, _ := enc.ParseNat(mNode.Matching["time"])
	fmt.Printf(">> I: timestamp: %d\n", timestamp)
	content := []byte("Hello, world!")
	dataWire := mNode.Call("Provide", enc.Wire{content}).(enc.Wire)
	err := event.Reply(dataWire)
	if err != nil {
		log.WithField("module", "main").Errorf("unable to reply with data: %+v", err)
		return true
	}
	fmt.Printf("<< D: %s\n", mNode.Name.String())
	fmt.Printf("Content: (size: %d)\n", len(content))
	fmt.Printf("\n")
	return nil
}

func main() {
	log.SetLevel(log.InfoLevel)
	logger := log.WithField("module", "main")

	// Setup schema tree
	tree := schema.CreateFromJson(SchemaJson, map[string]any{
		"$onInterest": onInterest,
		"$isProducer": true,
	})

	// Start engine
	timer := basic_engine.NewTimer()
	// face := basic_engine.NewStreamFace("unix", "/var/run/nfd/nfd.sock", true)
	face := basic_engine.NewWasmWsFace("ws", "127.0.0.1:9696", true)
	app := basic_engine.NewEngine(face, timer, sec.NewSha256IntSigner(timer), passAll)
	err := app.Start()
	if err != nil {
		logger.Fatalf("Unable to start engine: %+v", err)
		return
	}
	defer app.Shutdown()

	// Attach schema
	prefix, _ := enc.NameFromStr("/example/testApp")
	err = tree.Attach(prefix, app)
	if err != nil {
		logger.Fatalf("Unable to attach the schema to the engine: %+v", err)
		return
	}
	defer tree.Detach()

	fmt.Print("Start serving ...\n")
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	receivedSig := <-sigChannel
	logger.Infof("Received signal %+v - exiting\n", receivedSig)
}
