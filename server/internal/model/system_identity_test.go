package model

import "testing"

func TestDefaultSystemIdentity(t *testing.T) {
	identity := DefaultSystemIdentity()

	if identity.CorpID != SystemCorporationID {
		t.Fatalf("expected corp ID %d, got %d", SystemCorporationID, identity.CorpID)
	}
	if identity.SiteTitle != SystemDisplayName {
		t.Fatalf("expected site title %q, got %q", SystemDisplayName, identity.SiteTitle)
	}
}
