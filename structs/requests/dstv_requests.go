package requests

type DSTVQueryRequest struct {
	BillerID           string
	DestinationAccount string
}

type DSTVPaymentRequest struct {
	DestinationAccount string
	RequestId          int64
	Amount             float64
	PackageType        string
}

type GOTVPaymentRequest struct {
	DestinationAccount string
	RequestId          int64
	Amount             float64
	PackageType        string
}

type ThirdPartyDSTVReqExtraData struct {
	Bundle string `json:"bundle"`
}

type ThirdPartyDSTVPaymentRequest struct {
	Destination     string
	Amount          float64
	CallbackUrl     string
	ClientReference string
	ExtraData       ThirdPartyDSTVReqExtraData
	ServiceId       string
}
