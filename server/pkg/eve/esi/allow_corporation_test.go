package esi

import "testing"

func TestIsCorporationAllowed(t *testing.T) {
	t.Run("empty allow list rejects all corporations", func(t *testing.T) {
		if isCorporationAllowed(98000001, nil) {
			t.Fatal("expected corporation to be rejected when allow list is empty")
		}
	})

	t.Run("zero corporation id is rejected when allow list is configured", func(t *testing.T) {
		if isCorporationAllowed(0, []int64{98000001}) {
			t.Fatal("expected zero corporation id to be rejected")
		}
	})

	t.Run("configured allow list only accepts matching corporation", func(t *testing.T) {
		if !isCorporationAllowed(98000002, []int64{98000001, 98000002}) {
			t.Fatal("expected matching corporation id to be allowed")
		}
		if isCorporationAllowed(98000003, []int64{98000001, 98000002}) {
			t.Fatal("expected corporation outside allow list to be rejected")
		}
	})
}
