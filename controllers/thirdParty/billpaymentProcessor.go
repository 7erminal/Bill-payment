package thirdparty

import (
	"billpayment_service/api"
	"billpayment_service/structs/requests"
	"billpayment_service/structs/responses"
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

func GetServiceId(serviceName string) (string, error) {
	serviceId := ""
	if serviceName != "" && containsIgnoreCase(serviceName, "mtn") {
		tempSId, _ := beego.AppConfig.String("MTNAirtime")
		serviceId = tempSId
	}
	if serviceName != "" && containsIgnoreCase(serviceName, "telecel") {
		tempSId, _ := beego.AppConfig.String("TelecelAirtime")
		serviceId = tempSId
	}
	if serviceName != "" && containsIgnoreCase(serviceName, "airtelTigo") {
		tempSId, _ := beego.AppConfig.String("AirtelTigoAirtime")
		serviceId = tempSId
	}

	return serviceId, nil
}

// containsIgnoreCase checks if substr is in s, case-insensitive.
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func ProcessBillPayment(c *beego.Controller, req requests.BillPaymentThirdPartyRequest) (responses.ThirdPartyBillPaymentResponse, error) {
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

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, read, "", "  "); err != nil {
		logs.Info("Raw response received is ", string(read))
	} else {
		logs.Info("Raw response received is \n", prettyJSON.String())
	}
	// data := map[string]interface{}{}
	// var dataOri responses.UserOriResponseDTO
	var data responses.ThirdPartyBillPaymentResponse
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}

func ProcessGhanaWaterBillPayment(c *beego.Controller, req requests.GhanaWaterBillPaymentThirdPartyRequest) (responses.ThirdPartyBillPaymentResponse, error) {
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

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, read, "", "  "); err != nil {
		logs.Info("Raw response received is ", string(read))
	} else {
		logs.Info("Raw response received is \n", prettyJSON.String())
	}
	// data := map[string]interface{}{}
	// var dataOri responses.UserOriResponseDTO
	var data responses.ThirdPartyBillPaymentResponse
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}

func ECGAccountQuery(c *beego.Controller, req requests.ECGQueryRequest) (responses.ThirdPartyQueryResponseData, error) {
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

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, read, "", "  "); err != nil {
		logs.Info("Raw response received is ", string(read))
	} else {
		logs.Info("Raw response received is \n", prettyJSON.String())
	}
	// data := map[string]interface{}{}
	// var dataOri responses.UserOriResponseDTO
	var data responses.ThirdPartyQueryResponseData
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}

func AccountQuery(c *beego.Controller, req requests.ThirdPartyQueryRequest) (responses.ThirdPartyQueryResponseData, error) {
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

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, read, "", "  "); err != nil {
		logs.Info("Raw response received is ", string(read))
	} else {
		logs.Info("Raw response received is \n", prettyJSON.String())
	}
	// data := map[string]interface{}{}
	// var dataOri responses.UserOriResponseDTO
	var data responses.ThirdPartyQueryResponseData
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}

func GetTransactionStatus(c *beego.Controller, req requests.TransactionStatusThirdPartyRequest) (responses.TransactionStatusThirdPartyResponse, error) {
	host, _ := beego.AppConfig.String("statusCheckBaseUrl")
	posSale, _ := beego.AppConfig.String("hubtelPOSSale")
	authorizationKey, _ := beego.AppConfig.String("authorizationKey")

	logs.Info("Sending transaction ID ", req.TransactionID)

	// serviceId, _ := GetServiceId(req.Network)
	logs.Info("Pos Sale ID is ", posSale)
	logs.Info("Url is ", host+"/transactions/"+posSale+"/status?clientReference="+req.TransactionID)

	request := api.NewRequest(
		host,
		"/transactions/"+posSale+"/status?clientReference="+req.TransactionID,
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

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, read, "", "  "); err != nil {
		logs.Info("Raw response received is ", string(read))
	} else {
		logs.Info("Raw response received is \n", prettyJSON.String())
	}
	// data := map[string]interface{}{}
	// var dataOri responses.UserOriResponseDTO
	var data responses.TransactionStatusThirdPartyResponse
	json.Unmarshal(read, &data)
	c.Data["json"] = data

	logs.Info("Resp is ", data)
	// logs.Info("Resp is ", data.User.Branch.Country.DefaultCurrency)

	return data, nil
}
