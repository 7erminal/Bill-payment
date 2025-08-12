package thirdparty

import (
	"billpayment_service/api"
	"billpayment_service/structs/requests"
	"billpayment_service/structs/responses"
	"encoding/json"
	"io"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

func ProcessDataBundlePurchase(c *beego.Controller, req requests.BillPaymentThirdPartyRequest) (responses.ThirdPartyBillPaymentResponse, error) {
	host, _ := beego.AppConfig.String("thirdPartyBaseUrl")
	prepaidId, _ := beego.AppConfig.String("hubtelPrepaidDepositID")
	authorizationKey, _ := beego.AppConfig.String("hubtelAuthorizationKey")

	logs.Info("Sending phone number ", req.PhoneNumber)

	serviceId, _ := GetServiceId(req.Network)

	request := api.NewRequest(
		host,
		"/"+prepaidId+"/"+serviceId,
		api.POST)
	request.HeaderField["Authorization"] = "Basic " + authorizationKey
	request.InterfaceParams["Destination"] = req.Destination
	request.InterfaceParams["Amount"] = req.Amount
	request.InterfaceParams["CallbackUrl"] = req.CallbackUrl
	request.InterfaceParams["ClientReference"] = req.ClientReference
	request.InterfaceParams["ExtraData"] = req.ExtraData
	// request.InterfaceParams["BundleId"] = req.BundleId

	// request.Params = {"UserId": strconv.Itoa(int(userid))}
	client := api.Client{
		Request: request,
		Type_:   "body",
	}
	res, err := client.SendRequest()
	if err != nil {
		logs.Error("client.Error: %v", err)
		c.Data["json"] = err.Error()
	}
	defer res.Body.Close()
	read, err := io.ReadAll(res.Body)
	if err != nil {
		c.Data["json"] = err.Error()
	}

	logs.Info("Raw response received is ", res)
	// data := map[string]interface{}{}
	// var dataOri responses.UserOriResponseDTO
	var data responses.ThirdPartyBillPaymentResponse
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}
