package responses

type DSTVQueryData struct {
	Display string
	Value   string
	Amount  float64
}

type DstvQueryResponseData struct {
	ResponseCode string
	Message      string
	Label        string
	Display      string
	Value        string
	Amount       float64
	Data         []DSTVQueryData
}

type DSTVQueryResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *[]DSTVQueryData
}

type BillPaymentMeta struct {
	Commission string
}

type DSTVPaymentResponse struct {
	ResponseCode    string
	Message         string
	ClientReference string
	TransactionId   string
	Amount          float64
	Meta            BillPaymentMeta
	Commission      string
}

type DSTVBillPaymentDataResponse struct {
	Description   string
	Amount        float64
	TransactionId string
}

type DSTVBillPaymentResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *DSTVBillPaymentDataResponse
}

type WaterBillQueryData struct {
	Display string
	Value   string
	Amount  float64
}

type WaterBillQueryResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *[]WaterBillQueryData
}

type StartimesQueryData struct {
	Display string
	Value   string
	Amount  float64
}

type StartimesQueryResponse struct {
	StatusCode    bool
	StatusMessage string
	Result        *[]StartimesQueryData
}
