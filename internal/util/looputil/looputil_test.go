package looputil_test

import (
	"testing"
	"github.com/bitrise-io/stepman/models"
	"github.com/lszucs/build-aggregator/internal/util/looputil"
)

func TestHasNext(t *testing.T) {
	c := []models.StepModel{
		models.StepModel{},
		models.StepModel{},
		models.StepModel{},
	}
	batcher := looputil.Batcher{Collection: c}

	batcher.Next(1)
	batcher.Next(1)
	batcher.Next(1)
	batcher.Next(1)

	got := batcher.HasNext()
	if got != false {
		t.Errorf("got true want false")
	}
}

func TestNext(t *testing.T) {
	fst, snd, trd := "first", "second", "third"
	c := []models.StepModel{
		models.StepModel{Title: &fst},
		models.StepModel{Title: &snd},
		models.StepModel{Title: &trd},
	}
	batcher := looputil.Batcher{Collection: c}

	first := batcher.Next(1)
	second := batcher.Next(1)
	third := batcher.Next(1)

	if *first[0].Title != "first" {
		t.Errorf("first batch next error")
	}
	if *second[0].Title != "second" {
		t.Errorf("second batch next error")
	}
	if *third[0].Title != "third" {
		t.Errorf("third batch next error")
	}
}