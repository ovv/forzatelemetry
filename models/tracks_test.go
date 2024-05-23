package models_test

import (
	"testing"

	"forzatelemetry/models"
)

func TestTrack(t *testing.T) {
	track := models.Track{Name: "Foo", Layout: "Circuit"}
	wants := "Foo - Circuit"

	if track.FullName() != wants {
		t.Errorf("got %s instead of %s", track.FullName(), wants)
	}
}

func TestTracksGeneration(t *testing.T) {
	if len(models.Tracks) < 10 {
		t.Errorf("found less than 10 generated tracks")
	}
}
