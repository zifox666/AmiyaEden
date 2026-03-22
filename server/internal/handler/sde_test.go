package handler

import (
	"amiya-eden/internal/repository"
	"testing"
)

func TestMergeGetNamesNamespacesPreservesNamespacesAndFlatFirstWins(t *testing.T) {
	result := newGetNamesResponse()

	mergeGetNamesNamespaces(&result, repository.SdeNameMap{
		"type": {
			1: "Rifter",
			2: "Punisher",
		},
		"solar_system": {
			1: "Jita",
			3: "Amarr",
		},
	})

	if got := result.Names["type"][1]; got != "Rifter" {
		t.Fatalf("expected type namespace to keep Rifter, got %q", got)
	}
	if got := result.Names["solar_system"][1]; got != "Jita" {
		t.Fatalf("expected solar_system namespace to keep Jita, got %q", got)
	}
	if got := result.Flat[1]; got != "Jita" {
		t.Fatalf("expected sorted namespace merge to make flat[1] use solar_system first, got %q", got)
	}
	if got := result.Flat[2]; got != "Punisher" {
		t.Fatalf("expected flat[2] = Punisher, got %q", got)
	}
	if got := result.Flat[3]; got != "Amarr" {
		t.Fatalf("expected flat[3] = Amarr, got %q", got)
	}
}

func TestMergeGetNamesESIPreservesFlatCompatibility(t *testing.T) {
	result := newGetNamesResponse()
	result.Flat[42] = "Existing Flat Name"

	mergeGetNamesESI(&result, []getNamesESIEntry{
		{ID: 42, Name: "Pilot Forty Two"},
		{ID: 99, Name: "Capsuleer"},
	})

	if got := result.Names["esi"][42]; got != "Pilot Forty Two" {
		t.Fatalf("expected names.esi[42] to be updated, got %q", got)
	}
	if got := result.Flat[42]; got != "Existing Flat Name" {
		t.Fatalf("expected flat compatibility map to keep first value, got %q", got)
	}
	if got := result.Flat[99]; got != "Capsuleer" {
		t.Fatalf("expected flat[99] = Capsuleer, got %q", got)
	}
}
