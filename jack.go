package main

import (
	"fmt"

	"github.com/faiface/beep"
	"github.com/xthexder/go-jack"
)

// processor adapts a beep.Streamer to jack
type processor struct {
	portL  *jack.Port
	portR  *jack.Port
	source beep.Streamer
}

func (p *processor) Process(nframes uint32) int {
	samplesL := p.portL.GetBuffer(nframes)
	samplesR := p.portR.GetBuffer(nframes)

	sourceSamples := make([][2]float64, len(samplesL))
	n, _ := p.source.Stream(sourceSamples)
	if n == 0 {
		return 0
	}

	for i, sample := range sourceSamples {
		samplesL[i] = jack.AudioSample(sample[0])
		samplesR[i] = jack.AudioSample(sample[1])
	}

	return 0
}

type portConnection struct {
	a, b jack.PortId
}

// playWithJack puts a beep.Streamer into jack under the given name
func playWithJack(input beep.Streamer, name string) error {
	jack.SetErrorFunction(func(err string) {
		fmt.Println("\tjack err:", err)
	})

	jack.SetInfoFunction(func(s string) {
		fmt.Println("\tjack info:", s)
	})

	// Note: On macOS the library can sometimes loose connection to jack shortly after starting.
	// For this reason we will try to reconnect and recover all lost connections.

	// A map between the names of output ports of this client and port id's that they are connected to.
	// Used to reconnect them when we loose connection to Jack.
	savedConnections := make(map[string][]jack.PortId)

	// When two ports should be connected they can be put into this channel and the main function will do so.
	// Making this channel buffered prevents blocking the callback in wich it is filled.
	// It works without, but otherwise reconnecting can take a couple seconds.
	connectChan := make(chan portConnection, 2)

	// The audio source that Jack will read from.
	proc := processor{source: input}

	for {
		err := connectJack(name, proc, savedConnections, connectChan)
		if err != nil {
			return err
		}
	}
}

func connectJack(name string, proc processor, savedConnections map[string][]jack.PortId, connectChan chan portConnection) error {
	client, status := jack.ClientOpen(name, jack.NoStartServer)
	if status != 0 {
		return fmt.Errorf("open client: %s", jack.StrError(status))
	}
	defer client.Close()

	if code := client.SetProcessCallback(proc.Process); code != 0 {
		return fmt.Errorf("set processor: %s", jack.StrError(code))
	}

	shutdown := make(chan struct{})
	client.OnShutdown(func() {
		fmt.Println("Disconnected from JACK server")
		close(shutdown)
	})

	// Callback that get's called when two ports are connected. The ports doesn't necessarily involve our client.
	client.SetPortConnectCallback(func(a, b jack.PortId, connected bool) {
		// Note: If GetPortById get's called with an invalid id, it will not return nil but an empty port.
		port := client.GetPortById(a)
		if !client.IsPortMine(port) {
			return
		}
		portName := port.GetName()
		if connected {
			savedConnections[portName] = append(savedConnections[portName], b)
			fmt.Printf("\tPort %q was connected to %v\r\n", port.GetShortName(), b)
		} else {
			savedConnections[portName] = removeFromSlice(savedConnections[portName], b)
		}
	})

	// Callback that get's called when a new ports get's created. The port doesn't necessarily involve our client.
	client.SetPortRegistrationCallback(func(p jack.PortId, registered bool) {
		port := client.GetPortById(p)
		if !client.IsPortMine(port) {
			return
		}
		portName := port.GetName()
		if connections, ok := savedConnections[portName]; ok {
			// Reconnect this port to all saved connections.
			// We can't do that in this callback so instead we will send it through a channel to the main function.
			for _, c := range connections {
				connectChan <- struct{ a, b jack.PortId }{p, c}
			}
			delete(savedConnections, portName)
		}
	})

	proc.portL = client.PortRegister("Left", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)
	proc.portR = client.PortRegister("Right", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)

	if code := client.Activate(); code != 0 {
		return fmt.Errorf("activate client: %s", jack.StrError(code))
	}

	fmt.Println("Jack connected")

	for {
		select {
		case <-shutdown:
			return nil
		case msg := <-connectChan:
			fmt.Printf("\tReconnecting %d to %d\r\n", msg.a, msg.b)
			a := client.GetPortById(msg.a)
			b := client.GetPortById(msg.b)
			client.ConnectPorts(a, b)
		}
	}
}

func removeFromSlice(list []jack.PortId, remove jack.PortId) []jack.PortId {
	for i, c := range list {
		if c == remove {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}
