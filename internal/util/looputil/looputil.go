package looputil

import (
	"github.com/bitrise-io/stepman/models"
)

// Batcher provides easy iteration through a collection of steps
// in a batched manner
type Batcher struct {
	Collection []models.StepModel
	nextStart  int
}

// Next returns the next batch from the underlying collection
func (b *Batcher) Next(size int) []models.StepModel {
	if remaining := len(b.Collection) - b.nextStart; size > remaining {
		size = remaining
	}

	end := b.nextStart + size
	batch := b.Collection[b.nextStart:end]

	b.nextStart = end
	return batch
}

// HasNext returns true if a call to Next would yield
// a nonempty slice
func (b Batcher) HasNext() bool {
	return b.nextStart < len(b.Collection)
}
