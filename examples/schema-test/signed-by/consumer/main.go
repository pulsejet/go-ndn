package main

import (
	"encoding/base64"
	"fmt"
	"os"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	basic_engine "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	"github.com/zjkmxy/go-ndn/pkg/log"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
	"github.com/zjkmxy/go-ndn/pkg/schema"
	"github.com/zjkmxy/go-ndn/pkg/schema/demosec"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
)

const (
	TrustAnchorPktB64 = ("BsMHKAgHZXhhbXBsZQgGc2NoZW1hCAhzaWduZWRCeQgLdHJ1c3RBbmNob3IUCRgBAhkEADbugBUP" +
		"SG1hY1RydXN0QW5jaG9yFlkbAQQcKgcoCAdleGFtcGxlCAZzY2hlbWEICHNpZ25lZEJ5CAt0cnVz" +
		"dEFuY2hvcv0A/Sb9AP4PMjAyMzAyMTNUMDU1NzU0/QD/DzIwNDMwMjA4VDA1NTc1NBcg0FmaCWVb" +
		"U7ei6w5fTNGS5KOklhhMfA9eLvRaUCfYpLw=")

	SchemaJson = `{
		"nodes": {
			"/<nodeId>/data": {
				"type": "LeafNode",
				"attrs": {
					"CanBePrefix": false,
					"MustBeFresh": true,
					"Lifetime": 6000,
					"Freshness": 1000,
					"ValidDuration": 3153600000000.0
				}
			},
			"/<nodeId>/key": {
				"type": "LeafNode",
				"attrs": {
					"CanBePrefix": false,
					"MustBeFresh": true,
					"Lifetime": 6000,
					"Freshness": 3600000
				}
			},
			"/trustAnchor": {
				"type": "LeafNode",
				"attrs": {
					"CanBePrefix": false,
					"MustBeFresh": false,
					"SupressInt": true
				}
			}
		},
		"policies": [
			{
				"type": "RegisterPolicy",
				"path": "/<nodeId>",
				"attrs": {
					"RegisterIf": "$isProducer",
					"Patterns": {
						"nodeId": "$nodeId"
					}
				}
			},
			{
				"type": "SignedBy",
				"path": "/<nodeId>/data",
				"attrs": {
					"KeyNodePath": "/<nodeId>/key",
					"Mapping": {
						"nodeId": "$nodeId"
					},
					"KeyStore": "$keyStore"
				}
			},
			{
				"type": "SignedBy",
				"path": "/<nodeId>/key",
				"attrs": {
					"KeyNodePath": "/trustAnchor",
					"Mapping": {
						"nodeId": "$nodeId"
					},
					"KeyStore": "$keyStore"
				}
			},
			{
				"type": "MemStorage",
				"path": "/<nodeId>/data",
				"attrs": {}
			},
			{
				"type": "KeyStoragePolicy",
				"path": "/<nodeId>/key",
				"attrs": {
					"KeyStore": "$keyStore"
				}
			},
			{
				"type": "KeyStoragePolicy",
				"path": "/trustAnchor",
				"attrs": {
					"KeyStore": "$keyStore"
				}
			}
		]
	}`
)

func passAll(enc.Name, enc.Wire, ndn.Signature) bool {
	return true
}

func main() {
	log.SetLevel(log.InfoLevel)
	logger := log.WithField("module", "main")

	if len(os.Args) < 2 {
		logger.Fatal("Insufficient argument. Please input the version number given by the producer.")
		return
	}
	nodeId := os.Args[1]
	prefix, _ := enc.NameFromStr("/example/schema/signedBy")

	// Create key store
	keyStore := demosec.NewDemoHmacKeyStore()

	// Enroll trust anchor
	trustAnchorBuf, err := base64.StdEncoding.DecodeString(TrustAnchorPktB64)
	if err != nil {
		logger.Fatalf("Invalid trust anchor: %+v", err)
		return
	}
	err = keyStore.AddTrustAnchor(trustAnchorBuf)
	if err != nil {
		logger.Fatalf("Unable to add trust anchor: %+v", err)
		return
	}

	// Setup schema tree
	tree := schema.CreateFromJson(SchemaJson, map[string]any{
		"$isProducer": false,
		"$nodeId":     nodeId,
		"$keyStore":   keyStore,
	})

	// Start engine
	timer := basic_engine.NewTimer()
	face := basic_engine.NewStreamFace("unix", "/var/run/nfd/nfd.sock", true)
	app := basic_engine.NewEngine(face, timer, sec.NewSha256IntSigner(timer), passAll)
	err = app.Start()
	if err != nil {
		logger.Fatalf("Unable to start engine: %+v", err)
		return
	}
	defer app.Shutdown()

	// Attach schema
	err = tree.Attach(prefix, app)
	if err != nil {
		logger.Fatalf("Unable to attach the schema to the engine: %+v", err)
		return
	}
	defer tree.Detach()

	// Fetch the data
	path, _ := enc.NamePatternFromStr("/<nodeId>/data")
	node := tree.At(path)
	mNode := node.Apply(enc.Matching{
		"nodeId": []byte(nodeId),
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
}
