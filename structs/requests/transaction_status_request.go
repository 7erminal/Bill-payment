package requests

type TransactionStatusRequest struct {
	TransactionID           string
	ThirdParthTransactionID string
	NetworkTransactionID    string
}

type TransactionStatusThirdPartyRequest struct {
	TransactionID           string
	ThirdParthTransactionID string
	NetworkTransactionID    string
}
