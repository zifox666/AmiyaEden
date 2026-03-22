package esi

func isCorporationAllowed(corporationID int64, allowCorporations []int64) bool {
	if corporationID == 0 || len(allowCorporations) == 0 {
		return false
	}
	for _, allowedID := range allowCorporations {
		if allowedID == corporationID {
			return true
		}
	}
	return false
}
