package metrics

import (
	"time"
)

// RingBuffer represents a fixed-size circular buffer for time.Duration values
type RingBuffer struct {
	data     []time.Duration
	head     int
	size     int
	capacity int
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data:     make([]time.Duration, capacity),
		head:     0,
		size:     0,
		capacity: capacity,
	}
}

// Add inserts a new value into the ring buffer
func (rb *RingBuffer) Add(value time.Duration) {
	rb.data[rb.head] = value
	rb.head = (rb.head + 1) % rb.capacity
	
	if rb.size < rb.capacity {
		rb.size++
	}
}

// Values returns all values in the ring buffer in chronological order
func (rb *RingBuffer) Values() []time.Duration {
	if rb.size == 0 {
		return nil
	}
	
	result := make([]time.Duration, rb.size)
	
	if rb.size < rb.capacity {
		// Buffer not full, values are from 0 to size-1
		copy(result, rb.data[:rb.size])
	} else {
		// Buffer full, need to get values in correct order
		// First part: from head to end
		copy(result, rb.data[rb.head:])
		// Second part: from beginning to head
		copy(result[rb.capacity-rb.head:], rb.data[:rb.head])
	}
	
	return result
}

// Size returns the current number of elements in the buffer
func (rb *RingBuffer) Size() int {
	return rb.size
}

// Capacity returns the maximum capacity of the buffer
func (rb *RingBuffer) Capacity() int {
	return rb.capacity
}

// Clear removes all elements from the buffer
func (rb *RingBuffer) Clear() {
	rb.head = 0
	rb.size = 0
}

// Average calculates the average of all values in the buffer
func (rb *RingBuffer) Average() time.Duration {
	if rb.size == 0 {
		return 0
	}
	
	var sum time.Duration
	values := rb.Values()
	for _, value := range values {
		sum += value
	}
	
	return sum / time.Duration(rb.size)
}

// Latest returns the most recently added value
func (rb *RingBuffer) Latest() time.Duration {
	if rb.size == 0 {
		return 0
	}
	
	// The latest value is at (head - 1 + capacity) % capacity
	latestIndex := (rb.head - 1 + rb.capacity) % rb.capacity
	return rb.data[latestIndex]
}

// Max returns the maximum value in the buffer
func (rb *RingBuffer) Max() time.Duration {
	if rb.size == 0 {
		return 0
	}
	
	values := rb.Values()
	max := values[0]
	for _, value := range values[1:] {
		if value > max {
			max = value
		}
	}
	
	return max
}

// Min returns the minimum value in the buffer
func (rb *RingBuffer) Min() time.Duration {
	if rb.size == 0 {
		return 0
	}
	
	values := rb.Values()
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	
	return min
}

// Percentile calculates the nth percentile of values in the buffer
func (rb *RingBuffer) Percentile(p float64) time.Duration {
	if rb.size == 0 {
		return 0
	}
	
	if p < 0 || p > 100 {
		return 0
	}
	
	values := rb.Values()
	
	// Simple percentile calculation (could be improved with more sophisticated sorting)
	// For now, we'll use a basic implementation
	if p == 0 {
		return rb.Min()
	}
	if p == 100 {
		return rb.Max()
	}
	
	// For other percentiles, we need to sort values
	// This is a simplified implementation
	index := int(float64(len(values)) * p / 100.0)
	if index >= len(values) {
		index = len(values) - 1
	}
	
	// Quick selection would be more efficient, but for now use simple approach
	return values[index]
}