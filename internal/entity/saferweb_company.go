package entity

import "time"

type SaferwebCompany struct {
	ID        int64
	DOTNumber string
	LegalName string
	DBAName   string
	RawJSON   []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
