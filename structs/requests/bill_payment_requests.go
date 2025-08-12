package requests

type GetBillPaymentRequest struct {
	PhoneNumber string `json:"phone_number" valid:"required~Phone number is required"`
	Network     string `json:"network" valid:"required~Network is required"`
}

type ThirdPartyGetBillPaymentsRequest struct {
	Destination string `json:"destination" valid:"required~Phone number is required"`
}

type BillPaymentRequest struct {
	Amount      float64 `json:"amount" valid:"required~Amount is required"`
	Biller_code string  `json:"network" valid:"required~Network is required"`
	Destination string  `json:"destination" valid:"required~Destination is required"`
	ServiceId   int64   `json:"service" valid:"required~Service is required"`
}

type BillPaymentKeyRequest struct {
	Bundle string `json:"bundle" valid:"required~Bundle key is required"`
}

type BillPaymentThirdPartyRequest struct {
	PhoneNumber     string
	Amount          float64
	Network         string
	Destination     string
	CallbackUrl     string
	ClientReference string
	ExtraData       BillPaymentKeyRequest
	BundleId        string
	ServiceId       string
}

type GhanaWaterBillPaymentKeyRequest struct {
	Bundle    string `json:"bundle" valid:"required~Bundle key is required"`
	Email     string
	SessionId string
}

type GhanaWaterBillPaymentThirdPartyRequest struct {
	PhoneNumber     string
	Amount          float64
	Network         string
	Destination     string
	CallbackUrl     string
	ClientReference string
	ExtraData       GhanaWaterBillPaymentKeyRequest
	Bundle          string
	ServiceId       string
}

type ECGQueryRequest struct {
	BillerID           string
	DestinationAccount string
}

type ThirdPartyQueryRequest struct {
	BillerID           string
	DestinationAccount string
}

type ECGPaymentRequest struct {
	RequestId          int64
	DestinationAccount string
	Amount             float64
	PackageType        string
}

type StartimesPaymentRequest struct {
	RequestId          int64
	DestinationAccount string
	Amount             float64
	PackageType        string
}

type GhanaWaterPaymentRequest struct {
	RequestId          int64
	DestinationAccount string
	Amount             float64
	Bundle             string
	SessionId          string
	Email              string
}
