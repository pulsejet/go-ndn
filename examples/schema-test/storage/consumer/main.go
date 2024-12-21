package main

import (
	"fmt"
	"os"
	"strconv"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	"github.com/zjkmxy/go-ndn/pkg/engine"
	"github.com/zjkmxy/go-ndn/pkg/log"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
	"github.com/zjkmxy/go-ndn/pkg/schema"
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
      "type": "MemStorage",
      "path": "/",
      "attrs": {}
    }
  ]
}`

func main() {
	log.SetLevel(log.InfoLevel)
	logger := log.WithField("module", "main")

	if len(os.Args) < 2 {
		logger.Fatal("Insufficient argument. Please input the version number given by the producer.")
		return
	}
	ver, err := strconv.Atoi(os.Args[1])
	if err != nil {
		logger.Fatal("Invalid argument")
		return
	}

	// Setup schema tree
	tree := schema.CreateFromJson(SchemaJson, map[string]any{
		"$isProducer": false,
	})

	// Start engine
	face := engine.NewUnixFace("/var/run/nfd/nfd.sock")
	app := engine.NewBasicEngine(face)
	err = app.Start()
	if err != nil {
		logger.Fatalf("Unable to start engine: %+v", err)
		return
	}
	defer app.Stop()

	// Attach schema
	prefix, _ := enc.NameFromStr("/example/schema/storageApp")
	err = tree.Attach(prefix, app)
	if err != nil {
		logger.Fatalf("Unable to attach the schema to the engine: %+v", err)
		return
	}
	defer tree.Detach()

	// Fetch the data
	path, _ := enc.NamePatternFromStr("/randomData/<v=time>")
	node := tree.At(path)
	mNode := node.Apply(enc.Matching{
		"time": enc.Nat(ver).Bytes(),
	})
	result := <-mNode.Call("NeedChan").(chan schema.NeedResult)
	switch result.Status {
	case ndn.InterestResultNack:
		fmt.Printf("Nacked with reason=%d\n", *result.NackReason)
	case ndn.InterestResultTimeout:
		fmt.Printf("Timeout\n")
	case ndn.InterestCancelled:
		fmt.Printf("Canceled\n")
	case ndn.InterestResultData:
		fmt.Printf("Received Data: %+v\n", string(result.Content.Join()))
	}

	// Fetch the data again. No Interest should be sent
	result = <-mNode.Call("NeedChan").(chan schema.NeedResult)
	switch result.Status {
	case ndn.InterestResultNack:
		fmt.Printf("Nacked with reason=%d\n", *result.NackReason)
	case ndn.InterestResultTimeout:
		fmt.Printf("Timeout\n")
	case ndn.InterestCancelled:
		fmt.Printf("Canceled\n")
	case ndn.InterestResultData:
		fmt.Printf("Received Data: %+v\n", string(result.Content.Join()))
	}
}
