package requests

type CallbackMeta struct {
	Commission string
}

type CallbackData struct {
	AmountDebited         float64
	TransactionId         *string
	ClientReference       *string
	Description           *string
	ExternalTransactionId *string
	Amount                float64
	Charges               float64
	Meta                  *CallbackMeta
	RecipientName         *string
}

type CallbackRequest struct {
	ResponseCode string
	Data         CallbackData
}
