package responses

type CallbackResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        string
}

type TransactionStatusResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *TransactionStatusDataResponse
}

type TransactionStatusDataResponse struct {
	Date                  string  `json:"date"`
	TransactionID         string  `json:"transactionId"`
	Amount                float64 `json:"amount"`
	Status                string  `json:"status"`
	Charges               float64 `json:"charges"`
	AmountAfterCharges    float64 `json:"amountAfterCharges"`
	PaymentMethod         string  `json:"paymentMethod"`
	IsFulfilled           bool    `json:"isFulfilled"`
	ExternalTransactionId string  `json:"externalTransactionId"`
	ClientReference       string  `json:"clientReference"`
	CurrencyCode          string  `json:"currencyCode"`
}

type TransactionStatusThirdPartyResponse struct {
	Message      string                         `json:"message"`
	ResponseCode string                         `json:"responseCode"`
	Data         *TransactionStatusDataResponse `json:"data"`
}
