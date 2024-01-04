package main

import (
	"fmt"
	"time"

	"github.com/apex/log"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	basic_engine "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
	"github.com/zjkmxy/go-ndn/pkg/utils"
)

var app *basic_engine.Engine

func passAll(enc.Name, enc.Wire, ndn.Signature) bool {
	return true
}

func main() {
	timer := basic_engine.NewTimer()
	// face := basic_engine.NewWebSocketFace("ws", "localhost:9696", true)
	face := basic_engine.NewStreamFace("unix", "/var/run/nfd/nfd.sock", true)
	app = basic_engine.NewEngine(face, timer, sec.NewSha256IntSigner(timer), passAll)
	log.SetLevel(log.InfoLevel)
	logger := log.WithField("module", "main")
	err := app.Start()
	if err != nil {
		logger.Errorf("Unable to start engine: %+v", err)
		return
	}
	defer app.Shutdown()

	name, _ := enc.NameFromStr("/example/testApp/randomData")
	name = append(name, enc.NewTimestampComponent(utils.MakeTimestamp(timer.Now())))

	intCfg := &ndn.InterestConfig{
		MustBeFresh: true,
		Lifetime:    utils.IdPtr(6 * time.Second),
		Nonce:       utils.ConvertNonce(timer.Nonce()),
	}
	wire, _, finalName, err := app.Spec().MakeInterest(name, intCfg, nil, nil)
	if err != nil {
		logger.Errorf("Unable to make Interest: %+v", err)
		return
	}

	fmt.Printf("Sending Interest %s\n", finalName.String())
	ch := make(chan struct{})
	err = app.Express(finalName, intCfg, wire,
		func(result ndn.InterestResult, data ndn.Data, rawData, sigCovered enc.Wire, nackReason uint64) {
			switch result {
			case ndn.InterestResultNack:
				fmt.Printf("Nacked with reason=%d\n", nackReason)
			case ndn.InterestResultTimeout:
				fmt.Printf("Timeout\n")
			case ndn.InterestCancelled:
				fmt.Printf("Canceled\n")
			case ndn.InterestResultData:
				fmt.Printf("Received Data Name: %s\n", data.Name().String())
				fmt.Printf("%+v\n", data.Content().Join())
			}
			ch <- struct{}{}
		})
	if err != nil {
		logger.Errorf("Unable to send Interest: %+v", err)
		return
	}

	fmt.Printf("Wait for result ...\n")
	<-ch
}
