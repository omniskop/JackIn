package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gen2brain/malgo"
)

func getDeviceSource(context *malgo.AllocatedContext, deviceID *malgo.DeviceID, samplingRate int) (*AudioBuffer, error) {
	buffer := NewAudioBuffer(samplingRate) // audio buffer for one second

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = 2
	if deviceID != nil {
		deviceConfig.Capture.DeviceID = deviceID.Pointer()
	}
	deviceConfig.SampleRate = uint32(samplingRate)

	// sizeInBytes := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))

	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		buf := make([][2]float32, framecount)
		sampleBuf := bytes.NewBuffer(pSample)
		err := binary.Read(sampleBuf, binary.LittleEndian, buf)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
		}

		outBuf := make([][2]float64, framecount)
		for i, b := range buf {
			outBuf[i][0] = float64(b[0])
			outBuf[i][1] = float64(b[1])
		}

		buffer.Push(outBuf...)
	}

	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	device, err := malgo.InitDevice(context.Context, deviceConfig, captureCallbacks)
	if err != nil {
		return nil, err
	}

	err = device.Start()
	if err != nil {
		return nil, err
	}

	// device.Uninit()

	//

	return buffer, nil
}
