package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/faiface/beep"
	"github.com/gen2brain/malgo"
)

func main() {
	cardLabel := flag.String("card", "", "Label of the sound card to use")
	jackName := flag.String("name", "", "Name that should be used in jack")
	sourceRate := flag.Int("sourcerate", 48000, "Sample rate of the source card (for resampling)")
	targetRate := flag.Int("targetrate", 48000, "Sample rate of the jack server (for resampling)")
	listDevices := flag.Bool("list", false, "List available sound cards")
	flag.Parse()

	// get audio source
	context, deviceInfos, close, err := setupAudio()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer close()

	if *listDevices {
		printDeviceList(context, deviceInfos)
		return
	}

	cardID, cardName := findCard(deviceInfos, *cardLabel)

	fmt.Printf("using card %q\r\n", cardName)

	buffer, err := getDeviceSource(context, cardID, *sourceRate)
	if err != nil {
		fmt.Println(err)
		return
	}

	// maybe resample the audio
	var resampled beep.Streamer = buffer
	if *sourceRate != *targetRate {
		resampled = beep.Resample(3, beep.SampleRate(*sourceRate), beep.SampleRate(*targetRate), buffer)
	}

	// play in jack
	if *jackName == "" {
		*jackName = fmt.Sprintf("%s (jackin)", cardName)
	}

	err = playWithJack(resampled, *jackName)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// setupAudio initializes malgo and returns a list of available devices
func setupAudio() (*malgo.AllocatedContext, []malgo.DeviceInfo, func(), error) {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	deviceInfos, err := context.Devices(malgo.Capture)
	if err != nil {
		context.Uninit()
		context.Free()
		return nil, nil, nil, err
	}

	return context, deviceInfos, func() {
		context.Uninit()
		context.Free()
	}, err
}

// printDeviceList prints a list of available devices
func printDeviceList(context *malgo.AllocatedContext, deviceInfos []malgo.DeviceInfo) {
	var table Table
	table.Add("", "name", "channels", "sample rate")
	for i, info := range deviceInfos {
		full, err := context.DeviceInfo(malgo.Capture, info.ID, malgo.Shared)
		if err != nil {
			continue
		}
		table.Add(
			fmt.Sprintf("%2d", i), deviceName(info.Name()),
			fmt.Sprintf("%d-%d", full.MinChannels, full.MinChannels),
			fmt.Sprintf("%d-%d", full.MinSampleRate, full.MaxSampleRate),
		)
		// fmt.Printf("%2d: %q\t | channels: %d-%d, samplerate: %d-%d\r\n",
		// 	i, deviceName(info.Name()), full.MinChannels, full.MaxChannels, full.MinSampleRate, full.MaxSampleRate)
	}
	fmt.Println(table.String())
}

// findCard returns the ID and name of the card with the given label
func findCard(devices []malgo.DeviceInfo, name string) (*malgo.DeviceID, string) {
	var selectedCard malgo.DeviceInfo
	var cardID *malgo.DeviceID
	for _, info := range devices {
		if deviceName(info.Name()) == name {
			selectedCard = info
			cardID = &info.ID
			break
		}
	}
	if cardID == nil {
		return nil, "default"
	}
	return cardID, deviceName(selectedCard.Name())
}

// deviceName removes null-terminators from the device name
func deviceName(str string) string {
	return strings.Trim(str, "\x00")
}
