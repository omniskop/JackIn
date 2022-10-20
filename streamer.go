package main

type audioContainer struct {
	data [][2]float64
}

func newAudioContainer(len int) *audioContainer {
	return &audioContainer{
		data: make([][2]float64, len),
	}
}

func (c *audioContainer) Stream(samples [][2]float64) (n int, ok bool) {
	if len(c.data) == 0 {
		return 0, false
	}
	if len(c.data) < len(samples) {
		copy(samples, c.data)
		return len(c.data), true
	}
	copy(samples, c.data[:len(samples)])
	c.data = c.data[len(samples):]
	return len(samples), true
}

func (c *audioContainer) Err() error {
	return nil
}

func fill(samples [][2]float64, value float64) {
	for i := range samples {
		samples[i] = [2]float64{float64(value), float64(value)}
	}
}
