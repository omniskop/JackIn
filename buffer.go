package main

type RingBuffer[T any] struct {
	data   []T
	head   int // index of the newest element
	tail   int // index of the oldest element
	length int
}

func NewRingBuffer[T any](len int) *RingBuffer[T] {
	return &RingBuffer[T]{
		data: make([]T, len),
	}
}

func (r *RingBuffer[T]) Len() int {
	return r.length
}

func (r *RingBuffer[T]) Cap() int {
	return len(r.data)
}

func (r *RingBuffer[T]) Push(v T) {
	// first increment the head
	r.head = (r.head + 1) % len(r.data)
	r.data[r.head] = v
	// if the head has catched up with the tail we need to move that up too
	if r.head == r.tail {
		r.tail = (r.tail + 1) % len(r.data)
	} else {
		r.length++
	}
}

func (r *RingBuffer[T]) Pop() (T, bool) {
	if r.length == 0 {
		return *new(T), false
	}
	r.length--
	datum := r.data[r.tail]
	r.tail = (r.tail + 1) % len(r.data)
	return datum, true
}

func (r *RingBuffer[T]) CopyTo(dst []T) int {
	if r.Len() == 0 {
		return 0
	}
	if len(dst) <= r.length {
		for i := range dst {
			dst[i], _ = r.Pop()
		}
		return len(dst)
	}
	l := r.length
	for i := 0; i < r.Len(); i++ {
		dst[i], _ = r.Pop()
	}
	return l
}

type AudioBuffer struct {
	buffer *RingBuffer[[2]float64]
}

func NewAudioBuffer(size int) *AudioBuffer {
	return &AudioBuffer{
		buffer: NewRingBuffer[[2]float64](size),
	}
}

func (b *AudioBuffer) Stream(samples [][2]float64) (n int, ok bool) {
	if b.buffer.Len() == 0 {
		fill(samples, 0)
		return len(samples), true
	}
	n = b.buffer.CopyTo(samples)
	return n, true
}

func (b *AudioBuffer) Err() error {
	return nil
}

func (b *AudioBuffer) Push(v ...[2]float64) {
	for _, sample := range v {
		b.buffer.Push(sample)
	}
}
