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
	_ "github.com/zjkmxy/go-ndn/pkg/schema/demosec"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
	"github.com/zjkmxy/go-ndn/pkg/utils"
)

const SchemaJson = `{
  "nodes": {
    "/randomData/<v=time>": {
      "type": "LeafNode",
      "attrs": {
        "CanBePrefix": false,
        "MustBeFresh": true,
        "Lifetime": 6000,
        "Freshness": 1000,
        "ValidDuration": 3153600000000.0
      }
    },
    "/contentKey": {
      "type": "ContentKeyNode",
      "attrs": {
        "Lifetime": 6000,
        "Freshness": 1000,
        "ValidDuration": 3153600000000.0
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
    },
    {
      "type": "FixedHmacSigner",
      "path": "/contentKey/<contentKeyID>",
      "attrs": {
        "KeyValue": "$hmacKey"
      }
    },
    {
      "type": "MemStorage",
      "path": "/",
      "attrs": {}
    }
  ]
}`
const HmacKey = "Hello, World!"

func passAll(enc.Name, enc.Wire, ndn.Signature) bool {
	return true
}

func main() {
	log.SetLevel(log.InfoLevel)
	logger := log.WithField("module", "main")

	// Setup schema tree
	tree := schema.CreateFromJson(SchemaJson, map[string]any{
		"$isProducer": true,
		"$hmacKey":    HmacKey,
	})

	// Start engine
	timer := basic_engine.NewTimer()
	face := basic_engine.NewStreamFace("unix", "/var/run/nfd/nfd.sock", true)
	app := basic_engine.NewEngine(face, timer, sec.NewSha256IntSigner(timer), passAll)
	err := app.Start()
	if err != nil {
		logger.Fatalf("Unable to start engine: %+v", err)
		return
	}
	defer app.Shutdown()

	// Attach schema
	prefix, _ := enc.NameFromStr("/example/schema/encryptionApp")
	err = tree.Attach(prefix, app)
	if err != nil {
		logger.Fatalf("Unable to attach the schema to the engine: %+v", err)
		return
	}
	defer tree.Detach()

	// Produce key and encrypt
	path, _ := enc.NamePatternFromStr("/contentKey")
	ckMNode := tree.At(path).Apply(enc.Matching{})
	ck := ckMNode.Call("GenKey")
	cipherText := ckMNode.Call("Encrypt", ck, enc.Wire{[]byte("Hello, world!")}).(enc.Wire)

	// Produce data
	ver := utils.MakeTimestamp(timer.Now())
	path, _ = enc.NamePatternFromStr("/randomData/<v=time>")
	node := tree.At(path)
	mNode := node.Apply(enc.Matching{
		"time": enc.Nat(ver).Bytes(),
	})
	mNode.Call("Provide", cipherText)
	fmt.Printf("Generated packet with version= %d\n", ver)

	// Wait for keyboard quit signal
	sigChannel := make(chan os.Signal, 1)
	fmt.Print("Start serving ...\n")
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	receivedSig := <-sigChannel
	logger.Infof("Received signal %+v - exiting\n", receivedSig)
}
