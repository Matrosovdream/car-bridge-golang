package model

type CarrierRequest struct {
	DOTNumber string `validate:"required,numeric"`
}

type CarrierResponse struct {
	DOTNumber string `json:"dot_number"`
	LegalName string `json:"legal_name"`
	DBAName   string `json:"dba_name,omitempty"`
}
