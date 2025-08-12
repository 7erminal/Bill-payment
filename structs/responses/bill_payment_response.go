package responses

type BillPaymentResponseResult struct {
	TransactionID     string
	PhoneNumber       string
	Amount            float64
	Network           string
	Destination       string
	TransactionStatus string
	TransactionDate   string
}

type BillPaymentResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *BillPaymentResponseResult
}

type ThirdPartyBillPaymentMeta struct {
	Commission string
}

type ThirdPartyBillPaymentDataResponse struct {
	ClientReference string
	Amount          float64
	TransactionId   string
	Meta            ThirdPartyBillPaymentMeta
}

type ThirdPartyBillPaymentResponse struct {
	ResponseCode string
	Message      string
	Data         ThirdPartyBillPaymentDataResponse
}

type ThirdPartyAccountQueryData struct {
	Display string
	Value   string
	Amount  float64
}

type ThirdPartyQueryResponseData struct {
	ResponseCode string
	Message      string
	Label        string
	Display      string
	Value        string
	Amount       float64
	Data         []ThirdPartyAccountQueryData
}

type ThirdPartyAccountQueryApiResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *[]ThirdPartyAccountQueryData
}

type ThirdPartyBillPaymentDataApiResponse struct {
	Description   string
	Amount        float64
	TransactionId string
}

type ThirdPartyBillPaymentApiResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *ThirdPartyBillPaymentDataApiResponse
}
