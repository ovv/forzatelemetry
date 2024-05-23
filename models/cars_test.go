package models_test

import (
	"testing"

	"forzatelemetry/models"
)

func TestCarGeneration(t *testing.T) {
	if len(models.Cars) < 10 {
		t.Errorf("found less than 10 generated cars")
	}
}
