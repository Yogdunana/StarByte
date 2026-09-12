package dto

// PortalResponse is the grayscale membership portal payload.
// It does not replace application submit / profile APIs.
type PortalResponse struct {
	Applications []*ApplicationResponse `json:"applications"`
	Hint         string                 `json:"hint"`
}
