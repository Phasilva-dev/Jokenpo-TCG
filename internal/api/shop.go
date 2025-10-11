package api

type PurchaseRequest struct {
	Quantity uint64 `json:"quantity"`
}

type PurchaseResponse struct {
	Cards []string `json:"cards"`
	Error string   `json:"error,omitempty"`
}