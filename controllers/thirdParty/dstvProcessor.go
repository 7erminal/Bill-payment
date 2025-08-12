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

func DSTVAccountQuery(c *beego.Controller, req requests.DSTVQueryRequest) (responses.DstvQueryResponseData, error) {
	host, _ := beego.AppConfig.String("thirdPartyBaseUrl")
	authorizationKey, _ := beego.AppConfig.String("authorizationKey")
	prepaidId, _ := beego.AppConfig.String("hubtelPrepaidDepositID")

	logs.Info("Sending account number ", req.DestinationAccount)

	request := api.NewRequest(
		host,
		"/"+prepaidId+"/"+req.BillerID+"?destination="+req.DestinationAccount,
		api.GET)
	request.HeaderField["Authorization"] = "Basic " + authorizationKey

	// request.Params = {"UserId": strconv.Itoa(int(userid))}
	client := api.Client{
		Request: request,
		Type_:   "params",
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
	var data responses.DstvQueryResponseData
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}

func ProcessDSTVBillPayment(c *beego.Controller, req requests.ThirdPartyDSTVPaymentRequest) (responses.ThirdPartyBillPaymentResponse, error) {
	host, _ := beego.AppConfig.String("thirdPartyBaseUrl")
	prepaidId, _ := beego.AppConfig.String("hubtelPrepaidDepositID")
	authorizationKey, _ := beego.AppConfig.String("authorizationKey")

	request := api.NewRequest(
		host,
		"/"+prepaidId+"/"+req.ServiceId,
		api.POST)
	request.HeaderField["Authorization"] = "Basic " + authorizationKey
	request.InterfaceParams["Destination"] = req.Destination
	request.InterfaceParams["Amount"] = req.Amount
	request.InterfaceParams["CallbackUrl"] = req.CallbackUrl
	request.InterfaceParams["ClientReference"] = req.ClientReference

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
