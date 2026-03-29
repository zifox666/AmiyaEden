package model

const (
	SystemCorporationID int64 = 98185110
	SystemDisplayName         = "Fuxi Legion"
)

type SystemIdentity struct {
	CorpID    int64  `json:"corp_id"`
	SiteTitle string `json:"site_title"`
}

func DefaultSystemIdentity() SystemIdentity {
	return SystemIdentity{
		CorpID:    SystemCorporationID,
		SiteTitle: SystemDisplayName,
	}
}
