package controllers

import (
	thirdparty "billpayment_service/controllers/thirdParty"
	"billpayment_service/models"
	"billpayment_service/structs/requests"
	"billpayment_service/structs/responses"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

// RequestController operations for Request
type RequestController struct {
	beego.Controller
}

// URLMapping ...
func (c *RequestController) URLMapping() {
	c.Mapping("PayDSTVBill", c.PayDSTVBill)
	c.Mapping("DSTVAccountQuery", c.DSTVAccountQuery)
	c.Mapping("ECGAccountQuery", c.ECGAccountQuery)
	c.Mapping("AccountQuery", c.AccountQuery)
	c.Mapping("GhanaWaterAccountQuery", c.GhanaWaterAccountQuery)
	c.Mapping("StartimesAccountQuery", c.StartimesAccountQuery)
	c.Mapping("GoTVAccountQuery", c.GoTVAccountQuery)
	c.Mapping("PayECGBill", c.PayECGBill)
	c.Mapping("PayWaterBill", c.PayWaterBill)
	c.Mapping("PayGoTVBill", c.PayGoTVBill)
	c.Mapping("PayStartimesBill", c.PayStartimesBill)
	c.Mapping("BilTransactions", c.BilTransactions)
}

// PayDSTVBill ...
// @Title Pay DSTV Bill
// @Description create Request
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	body		body 	requests.DSTVPaymentRequest	true		"body for Request content"
// @Success 201 {int} models.Request
// @Failure 403 body is empty
// @router /pay-dstv-bill [post]
func (c *RequestController) PayDSTVBill() {
	var req requests.DSTVPaymentRequest
	json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	// Validate the request

	// authorization := ctx.Input.Header("Authorization")
	phoneNumber := c.Ctx.Input.Header("PhoneNumber")
	// sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := "500"
	responseMessage := "Request not processed"
	bilRespData := responses.BilTransactionsData{}

	// transactionInt, err := strconv.ParseInt(req.TransactionId, 10, 64)
	// if err != nil {
	// 	logs.Error("Failed to parse TransactionId: %v", err)
	// }
	if transaction, err := models.GetBil_transactionsByTransactionRefNum(req.TransactionId); err == nil {
		// Go to fulfillment
		// Formulate the request to send to the third-party service
		selectedPackage := requests.ThirdPartyDSTVReqExtraData{
			Bundle: req.PackageType,
		}

		callbackurl := ""
		if cbr, err := models.GetApplication_propertyByCode("BILL_PAYMENT_CALLBACK_URL"); err == nil {
			callbackurl = cbr.PropertyValue
		} else {
			logs.Error("Failed to get callback URL: %v", err)
		}

		billerCode := "DSTV"
		biller, err := models.GetBillerByCode(billerCode)

		if err == nil {
			clientId, _ := strconv.ParseInt(req.ClientId, 10, 64)
			if corpInfo, err := models.GetCorporateinfoById(clientId); err == nil {
				tReq := requests.ThirdPartyDSTVPaymentRequest{
					Amount:          req.Amount,
					Destination:     req.DestinationAccount,
					ClientReference: transaction.TransactionRefNumber, // Use the request ID as the transaction ID
					CallbackUrl:     callbackurl,                      // Optional field for callback URL
					ExtraData:       selectedPackage,                  // Assuming this is the bundle key request
					ServiceId:       biller.BillerReferenceId,
					PrepaidId:       corpInfo.PrepaidId,
				}

				// Insert in INS Transactions table
				reqText, err := json.Marshal(tReq)
				if err != nil {
					logs.Error("Failed to marshal request text: %v", err)
					// c.Data["json"] = "Invalid request format"
					// c.ServeJSON()
					// return
				}

				insTransaction := models.Bil_ins_transactions{
					BilTransactionId:       transaction,
					Amount:                 req.Amount,
					Biller:                 biller,
					SenderAccountNumber:    phoneNumber,
					RecipientAccountNumber: req.DestinationAccount,
					Network:                billerCode,
					Status:                 transaction.Status,
					Request:                string(reqText),
					DateCreated:            time.Now(),
					DateModified:           time.Now(),
					CreatedBy:              1,
					ModifiedBy:             1,
					Active:                 1,
				}

				if _, err := models.AddBil_ins_transactions(&insTransaction); err != nil {
					logs.Error("Failed to create INS transaction record: %v", err)
					responseCode = "500"
					responseMessage = "Failed to create INS transaction record"
					// resp := responses.ThirdPartyBillPaymentApiResponse{
					// 	StatusCode:    responseCode,
					// 	StatusMessage: responseMessage,
					// 	Result:        nil,
					// }
					// c.Data["json"] = resp
					// c.ServeJSON()
					// return
				}

				// Call the third-party service to process the request
				logs.Info("Processing dstv bill payment with third-party service: ", tReq)
				if thirdPartyResponse, err := thirdparty.ProcessDSTVBillPayment(&c.Controller, tReq); err == nil {

					if thirdPartyResponse.ResponseCode == "0001" {
						// Transaction is pending
						// Update the transaction status to pending
						responseCode = "200"
						responseMessage = "Request is being processed"
						if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "507"
								responseMessage = "PENDING:: Failed to update transaction status"
							} else {
								responseCode = "200"
								responseMessage = "Request is being processed"
							}
						} else {
							logs.Error("Failed to get status for pending transaction: %v", err)
							responseCode = "508"
							responseMessage = "PENDING: Failed to get status for pending transaction"
						}
					} else if thirdPartyResponse.ResponseCode == "0000" {
						// Transaction is successful
						// Update the transaction status to successful
						responseCode = "200"
						responseMessage = "Request is successful"
						if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "501"
								responseMessage = "SUCCESS:: Failed to update transaction status"
							} else {
								// Prepare the response
								logs.Info("Transaction successful: ", transaction)
								responseCode = "200"
								responseMessage = "Transaction successful"
							}
						} else {
							logs.Error("Failed to get status for successful transaction: %v", err)
							responseCode = "509"
							responseMessage = "SUCCESS:: Failed to get status for successful transaction"
						}
					} else {
						// Transaction failed
						// Update the transaction status to failed
						responseCode = "502"
						responseMessage = "Transaction failed"
						if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "503"
								responseMessage = "FAILED:: Failed to update transaction status"
							} else {
								responseCode = "504"
								responseMessage = "Transaction failed"
							}
						} else {
							logs.Error("Failed to get status for failed transaction: %v", err)
							responseCode = "509"
							responseMessage = "FAILED:: Failed to get status for failed transaction"
						}
					}

					resText, err := json.Marshal(thirdPartyResponse)
					if err != nil {
						logs.Error("Failed to marshal response text: %v", err)
						// c.Data["json"] = "Invalid request format"
						// c.ServeJSON()
						// return
					}
					if v, err := models.GetRequestById(transaction.Request.RequestId); err == nil {
						v.RequestResponse = string(resText)
						v.DateModified = time.Now()
						if err := models.UpdateRequestById(v); err != nil {
							logs.Error("Failed to update request response: %v", err)
							responseCode = "200"
							responseMessage = "Success response:: Failed to update request response"
						} else {
							logs.Info("Request response updated successfully")
							responseCode = "200"
							responseMessage = "Request processed successfully"
						}
					} else {
						logs.Error("Failed to retrieve request by ID: %v", err)
					}

					c.Ctx.Output.SetStatus(200)
					// Prepare the response

					// if status, err := models.GetStatus_codesByCode("IN_PROGRESS"); err == nil {
					// 	respData = responses.DSTVBillPaymentDataResponse{
					// 		Description:   "Payment for DSTV bill",
					// 		Amount:        req.Amount,
					// 		TransactionId: transaction.TransactionRefNumber,
					// 	}
					// }

					// Create the response object
				} else {
					logs.Error("Failed to process dstv request: %v", err)
					responseCode = "503"
					responseMessage = "Failed to process dstv request"
				}
			} else {
				logs.Error("Failed to get corporate info: %v", err)
				responseCode = "506"
				responseMessage = "Failed to get corporate info"
			}
		} else {
			logs.Error("Failed to get biller by code: %v", err)
			responseCode = "502"
			responseMessage = "Failed to get biller by code"
		}

	}

	response := responses.BilTransactionsResponse{
		StatusCode:    responseCode,
		StatusMessage: responseMessage,
		Result:        &bilRespData,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// DSTVAccountQuery ...
// @Title DSTV Account Query
// @Description get Request by id
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	accountNumber		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Request
// @Failure 403 :accountNumber is empty
// @router /dstv-account-query/:accountNumber [get]
func (c *RequestController) DSTVAccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")

	logs.Info("Received request to query DSTV account with number: ", accountNumberStr)

	responseCode := false
	responseMessage := "Request not processed"

	// statusCode := "PENDING"
	billerCode := "DSTV"
	biller, err := models.GetBillerByCode(billerCode)
	if err != nil {
		// c.Data["json"] = err.Error()
		responseMessage = "An error occurred while processing your request " + err.Error()
		resp := responses.DSTVQueryResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	} else {
		req := requests.DSTVQueryRequest{
			DestinationAccount: accountNumberStr,
			BillerID:           biller.BillerReferenceId,
		}
		logs.Info("Querying DSTV account with request: ", req)
		getAccountDetails, err := thirdparty.DSTVAccountQuery(&c.Controller, req)
		if err != nil {
			logs.Error("Failed to get account details: %v", err)
			responseMessage = "An error occurred while processing your request " + err.Error()
			resp := responses.DSTVQueryResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		} else {
			logs.Info("Account details retrieved successfully: ", getAccountDetails)
			if getAccountDetails.ResponseCode != "0000" {
				logs.Error("Failed to retrieve account details: ", getAccountDetails.Message)
				responseMessage = "An error occurred while processing your request " + getAccountDetails.Message
				resp := responses.DSTVQueryResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			logs.Info("Account details retrieved successfully. Sending response: ", getAccountDetails)
			responseCode = true
			responseMessage = "Request processed successfully"
			resp := responses.DSTVQueryResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        &getAccountDetails.Data,
			}
			c.Data["json"] = resp
		}
	}
	c.ServeJSON()
}

// GoTVAccountQuery ...
// @Title GoTV Account Query
// @Description get Request by id
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	accountNumber		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Request
// @Failure 403 :accountNumber is empty
// @router /gotv-account-query/:accountNumber [get]
func (c *RequestController) GoTVAccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")

	logs.Info("Received request to query DSTV account with number: ", accountNumberStr)

	responseCode := false
	responseMessage := "Request not processed"

	// statusCode := "PENDING"
	billerCode := "DSTV"
	biller, err := models.GetBillerByCode(billerCode)
	if err != nil {
		// c.Data["json"] = err.Error()
		responseMessage = "An error occurred while processing your request " + err.Error()
		resp := responses.DSTVQueryResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	} else {
		req := requests.DSTVQueryRequest{
			DestinationAccount: accountNumberStr,
			BillerID:           biller.BillerReferenceId,
		}
		logs.Info("Querying DSTV account with request: ", req)
		getAccountDetails, err := thirdparty.DSTVAccountQuery(&c.Controller, req)
		if err != nil {
			logs.Error("Failed to get account details: %v", err)
			responseMessage = "An error occurred while processing your request " + err.Error()
			resp := responses.DSTVQueryResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		} else {
			logs.Info("Account details retrieved successfully: ", getAccountDetails)
			if getAccountDetails.ResponseCode != "0000" {
				logs.Error("Failed to retrieve account details: ", getAccountDetails.Message)
				responseMessage = "An error occurred while processing your request " + getAccountDetails.Message
				resp := responses.DSTVQueryResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			logs.Info("Account details retrieved successfully. Sending response: ", getAccountDetails)
			responseCode = true
			responseMessage = "Request processed successfully"
			resp := responses.DSTVQueryResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        &getAccountDetails.Data,
			}
			c.Data["json"] = resp
		}
	}
	c.ServeJSON()
}

// ECGAccountQuery ...
// @Title ECG Account Query
// @Description get Request by id
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	accountNumber		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Request
// @Failure 403 :accountNumber is empty
// @router /ecg-account-query/:accountNumber [get]
func (c *RequestController) ECGAccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")

	logs.Info("Received request to query Ghana water account with number: ", accountNumberStr)

	responseCode := false
	responseMessage := "Request not processed"

	// statusCode := "PENDING"
	billerCode := "ECG"
	biller, err := models.GetBillerByCode(billerCode)
	if err != nil {
		// c.Data["json"] = err.Error()
		responseMessage = "An error occurred while processing your request " + err.Error()
		resp := responses.ThirdPartyAccountQueryApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	} else {
		req := requests.ECGQueryRequest{
			DestinationAccount: accountNumberStr,
			BillerID:           biller.BillerReferenceId,
		}
		logs.Info("Querying DSTV account with request: ", req)
		getAccountDetails, err := thirdparty.ECGAccountQuery(&c.Controller, req)
		if err != nil {
			logs.Error("Failed to get account details: %v", err)
			responseMessage = "An error occurred while processing your request " + err.Error()
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		} else {
			logs.Info("Account details retrieved successfully: ", getAccountDetails)
			if getAccountDetails.ResponseCode != "0000" {
				logs.Error("Failed to retrieve account details: ", getAccountDetails.Message)
				responseMessage = "An error occurred while processing your request " + getAccountDetails.Message
				resp := responses.ThirdPartyAccountQueryApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			logs.Info("Account details retrieved successfully. Sending response: ", getAccountDetails)
			responseCode = true
			responseMessage = "Request processed successfully"
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        &getAccountDetails.Data,
			}
			c.Data["json"] = resp
		}
	}
	c.ServeJSON()
}

// GhanaWaterAccountQuery ...
// @Title Water Account Query
// @Description get Request by id
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	accountNumber		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Request
// @Failure 403 :accountNumber is empty
// @router /ghana-water-account-query/:accountNumber/:mobileNumber [get]
func (c *RequestController) GhanaWaterAccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")
	mobilerNumberStr := c.Ctx.Input.Param(":mobileNumber")

	logs.Info("Received request to query Ghana water account with number: ", accountNumberStr, " and mobile number: ", mobilerNumberStr)

	responseCode := false
	responseMessage := "Request not processed"

	// statusCode := "PENDING"
	billerCode := "GH_WATER"
	biller, err := models.GetBillerByCode(billerCode)
	if err != nil {
		// c.Data["json"] = err.Error()
		responseMessage = "An error occurred while processing your request " + err.Error()
		resp := responses.ThirdPartyAccountQueryApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	} else {
		req := requests.ThirdPartyQueryRequest2{
			DestinationAccount: accountNumberStr,
			BillerID:           biller.BillerReferenceId,
			MobileNumber:       mobilerNumberStr,
		}
		logs.Info("Querying Ghana Water account with request: ", req)
		getAccountDetails, err := thirdparty.GhanaWaterAccountQuery(&c.Controller, req)
		if err != nil {
			logs.Error("Failed to get account details: %v", err)
			responseMessage = "An error occurred while processing your request " + err.Error()
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		} else {
			logs.Info("Account details retrieved successfully: ", getAccountDetails)
			if getAccountDetails.ResponseCode != "0000" {
				logs.Error("Failed to retrieve account details: ", getAccountDetails.Message)
				responseMessage = "An error occurred while processing your request " + getAccountDetails.Message
				resp := responses.ThirdPartyAccountQueryApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			logs.Info("Account details retrieved successfully. Sending response: ", getAccountDetails)
			responseCode = true
			responseMessage = "Request processed successfully"
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        &getAccountDetails.Data,
			}
			c.Data["json"] = resp
		}
	}
	c.ServeJSON()
}

// StarTimesAccountQuery ...
// @Title Star times Account Query
// @Description get Request by id
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	accountNumber		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Request
// @Failure 403 :accountNumber is empty
// @router /startimes-account-query/:accountNumber [get]
func (c *RequestController) StartimesAccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")

	logs.Info("Received request to query Ghana water account with number: ", accountNumberStr)

	responseCode := false
	responseMessage := "Request not processed"

	// statusCode := "PENDING"
	billerCode := "STARTIMES"
	biller, err := models.GetBillerByCode(billerCode)
	if err != nil {
		// c.Data["json"] = err.Error()
		responseMessage = "An error occurred while processing your request " + err.Error()
		resp := responses.ThirdPartyAccountQueryApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	} else {
		req := requests.ThirdPartyQueryRequest{
			DestinationAccount: accountNumberStr,
			BillerID:           biller.BillerReferenceId,
		}
		logs.Info("Querying DSTV account with request: ", req)
		getAccountDetails, err := thirdparty.AccountQuery(&c.Controller, req)
		if err != nil {
			logs.Error("Failed to get account details: %v", err)
			responseMessage = "An error occurred while processing your request " + err.Error()
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		} else {
			logs.Info("Account details retrieved successfully: ", getAccountDetails)
			if getAccountDetails.ResponseCode != "0000" {
				logs.Error("Failed to retrieve account details: ", getAccountDetails.Message)
				responseMessage = "An error occurred while processing your request " + getAccountDetails.Message
				resp := responses.ThirdPartyAccountQueryApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			logs.Info("Account details retrieved successfully. Sending response: ", getAccountDetails)
			responseCode = true
			responseMessage = "Request processed successfully"
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        &getAccountDetails.Data,
			}
			c.Data["json"] = resp
		}
	}
	c.ServeJSON()
}

// AccountQuery ...
// @Title Account Query
// @Description Account query
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	billercode		path 	string	true		"The key for staticblock"
// @Param	accountNumber		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Request
// @Failure 403 :accountNumber is empty
// @router /account-query/:billercode/:accountNumber [get]
func (c *RequestController) AccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")

	logs.Info("Received request to query Ghana water account with number: ", accountNumberStr)

	responseCode := false
	responseMessage := "Request not processed"

	// statusCode := "PENDING"
	billerCode := c.Ctx.Input.Param(":billercode")
	biller, err := models.GetBillerByCode(billerCode)
	if err != nil {
		// c.Data["json"] = err.Error()
		responseMessage = "An error occurred while processing your request " + err.Error()
		resp := responses.ThirdPartyAccountQueryApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	} else {
		req := requests.ECGQueryRequest{
			DestinationAccount: accountNumberStr,
			BillerID:           biller.BillerReferenceId,
		}
		logs.Info("Querying DSTV account with request: ", req)
		getAccountDetails, err := thirdparty.ECGAccountQuery(&c.Controller, req)
		if err != nil {
			logs.Error("Failed to get account details: %v", err)
			responseMessage = "An error occurred while processing your request " + err.Error()
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		} else {
			logs.Info("Account details retrieved successfully: ", getAccountDetails)
			if getAccountDetails.ResponseCode != "0000" {
				logs.Error("Failed to retrieve account details: ", getAccountDetails.Message)
				responseMessage = "An error occurred while processing your request " + getAccountDetails.Message
				resp := responses.ThirdPartyAccountQueryApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			logs.Info("Account details retrieved successfully. Sending response: ", getAccountDetails)
			responseCode = true
			responseMessage = "Request processed successfully"
			resp := responses.ThirdPartyAccountQueryApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        &getAccountDetails.Data,
			}
			c.Data["json"] = resp
		}
	}
	c.ServeJSON()
}

// PayECGBill ...
// @Title Pay ECG Bill
// @Description create Request
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	body		body 	requests.DSTVPaymentRequest	true		"body for Request content"
// @Success 201 {int} models.Request
// @Failure 403 body is empty
// @router /pay-ecg-bill [post]
func (c *RequestController) PayECGBill() {
	var req requests.ECGPaymentRequest
	json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	// Validate the request

	// authorization := ctx.Input.Header("Authorization")
	phoneNumber := c.Ctx.Input.Header("PhoneNumber")
	// sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := "501"
	responseMessage := "Request not processed"
	bilRespData := responses.BilTransactionsData{}

	// transactionInt, err := strconv.ParseInt(req.TransactionId, 10, 64)
	// if err != nil {
	// 	logs.Error("Failed to parse TransactionId: %v", err)
	// }
	if transaction, err := models.GetBil_transactionsByTransactionRefNum(req.TransactionId); err == nil {
		// Go to fulfillment
		clientId, _ := strconv.ParseInt(req.ClientId, 10, 64)
		if corpInfo, err := models.GetCorporateinfoById(clientId); err == nil {
			// Formulate the request to send to the third-party service
			selectedPackage := requests.BillPaymentKeyRequest{
				Bundle: req.PackageType,
			}

			callbackurl := ""
			if cbr, err := models.GetApplication_propertyByCode("BILL_PAYMENT_CALLBACK_URL"); err == nil {
				callbackurl = cbr.PropertyValue
			} else {
				logs.Error("Failed to get callback URL: %v", err)
			}

			billerCode := "ECG"
			biller, err := models.GetBillerByCode(billerCode)

			if err == nil {
				tReq := requests.BillPaymentThirdPartyRequest{
					Amount:          req.Amount,
					Destination:     req.DestinationAccount,
					ClientReference: transaction.TransactionRefNumber, // Use the request ID as the transaction ID
					CallbackUrl:     callbackurl,                      // Optional field for callback URL
					ExtraData:       selectedPackage,                  // Assuming this is the bundle key request
					ServiceId:       biller.BillerReferenceId,
					PrepaidId:       corpInfo.PrepaidId,
				}

				// Call the third-party service to process the request
				logs.Info("Processing dstv bill payment with third-party service: ", tReq)

				// Insert in INS Transactions table
				reqText, err := json.Marshal(tReq)
				if err != nil {
					logs.Error("Failed to marshal request text: %v", err)
					// c.Data["json"] = "Invalid request format"
					// c.ServeJSON()
					// return
				}

				insTransaction := models.Bil_ins_transactions{
					BilTransactionId:       transaction,
					Amount:                 req.Amount,
					Biller:                 biller,
					SenderAccountNumber:    phoneNumber,
					RecipientAccountNumber: req.DestinationAccount,
					Status:                 transaction.Status,
					Network:                billerCode,
					Request:                string(reqText),
					DateCreated:            time.Now(),
					DateModified:           time.Now(),
					CreatedBy:              1,
					ModifiedBy:             1,
					Active:                 1,
				}

				if _, err := models.AddBil_ins_transactions(&insTransaction); err != nil {
					logs.Error("Failed to create INS transaction record: %v", err)
					responseCode = "500"
					responseMessage = "Failed to create INS transaction record"
					// resp := responses.ThirdPartyBillPaymentApiResponse{
					// 	StatusCode:    responseCode,
					// 	StatusMessage: responseMessage,
					// 	Result:        nil,
					// }
					// c.Data["json"] = resp
					// c.ServeJSON()
					// return
				}
				logs.Info("Processing bill payment with third-party service: ", tReq)
				if thirdPartyResponse, err := thirdparty.ProcessBillPayment(&c.Controller, tReq); err == nil {

					if thirdPartyResponse.ResponseCode == "0001" {
						// Transaction is pending
						// Update the transaction status to pending
						responseCode = "200"
						responseMessage = "Request is being processed"
						if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "PENDING:: Failed to update transaction status"
							} else {
								responseCode = "200"
								responseMessage = "Request is being processed"
							}
						} else {
							logs.Error("Failed to get status for pending transaction: %v", err)
							responseCode = "506"
							responseMessage = "PENDING: Failed to get status for pending transaction"
						}
					} else if thirdPartyResponse.ResponseCode == "0000" {
						// Transaction is successful
						// Update the transaction status to successful
						responseCode = "200"
						responseMessage = "Request is successful"
						if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "SUCCESS:: Failed to update transaction status"
							} else {
								// Prepare the response
								logs.Info("Transaction successful: ", transaction)
								responseCode = "200"
								responseMessage = "Transaction successful"
							}
						} else {
							logs.Error("Failed to get status for successful transaction: %v", err)
							responseCode = "506"
							responseMessage = "SUCCESS:: Failed to get status for successful transaction"
						}
					} else {
						// Transaction failed
						// Update the transaction status to failed
						responseCode = "500"
						responseMessage = "Transaction failed"
						if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "500"
								responseMessage = "FAILED:: Failed to update transaction status"
							} else {
								responseCode = "502"
								responseMessage = "Transaction failed"
							}
						} else {
							logs.Error("Failed to get status for failed transaction: %v", err)
							responseCode = "506"
							responseMessage = "FAILED:: Failed to get status for failed transaction"
						}
					}

					resText, err := json.Marshal(thirdPartyResponse)
					if err != nil {
						logs.Error("Failed to marshal response text: %v", err)
						// c.Data["json"] = "Invalid request format"
						// c.ServeJSON()
						// return
					}

					if v, err := models.GetRequestById(transaction.Request.RequestId); err == nil {
						v.RequestResponse = string(resText)
						v.DateModified = time.Now()
						if err := models.UpdateRequestById(v); err != nil {
							logs.Error("Failed to update request response: %v", err)
							responseCode = "200"
							responseMessage = "Success response:: Failed to update request response"
						} else {
							responseCode = "200"
							responseMessage = "Success response:: Request response updated successfully"
							logs.Info("Request response updated successfully")
						}
					} else {
						logs.Error("Failed to get request by ID: %v", err)
					}

					c.Ctx.Output.SetStatus(200)
					// Prepare the response
					statusDesc := "PENDING"
					if transaction.Status, err = models.GetStatus_codesByCode("IN_PROGRESS"); err == nil {
						statusDesc = transaction.Status.StatusDescription
					}

					amountStr := fmt.Sprintf("%.2f", transaction.Amount)
					commisionStr := fmt.Sprintf("%.2f", thirdPartyResponse.Data.Meta.Commission)

					bilRespData = responses.BilTransactionsData{
						TransactionId:           transaction.TransactionId,
						TransactionRefNumber:    transaction.TransactionRefNumber,
						Service:                 transaction.Service.ServiceName,
						Amount:                  amountStr,
						Request:                 transaction.Request.RequestId,
						TransactionBy:           transaction.TransactionBy.FullName,
						TransactingCurrency:     transaction.TransactingCurrency,
						SourceChannel:           transaction.SourceChannel,
						Destination:             req.DestinationAccount,
						Charge:                  thirdPartyResponse.Data.Meta.Commission,
						BillerName:              biller.BillerName,
						NetworkName:             req.PackageType,
						Commission:              commisionStr,
						ExternalReferenceNumber: thirdPartyResponse.Data.TransactionId,
						Status:                  statusDesc,
						DateCreated:             transaction.DateCreated.Format("2006-01-02 15:04:05"),
						DateModified:            transaction.DateModified.Format("2006-01-02 15:04:05"),
						CreatedBy:               1,
						ModifiedBy:              1,
						Active:                  1,
					}

				} else {
					logs.Error("Failed to process "+billerCode+" request: %v", err)
					responseCode = "500"
					responseMessage = "Failed to process " + billerCode + " request"
				}
			} else {
				logs.Error("Failed to get biller by code: %v", err)
				responseCode = "504"
				responseMessage = "Failed to get biller by code"
			}
		} else {
			logs.Error("Failed to get corporate info: %v", err)
			responseCode = "505"
			responseMessage = "Failed to get corporate info"
		}
	} else {
		logs.Error("Failed to find transaction: %v", err)
		responseCode = "500"
		responseMessage = "Failed to find transaction"
	}

	response := responses.BilTransactionsResponse{
		StatusCode:    responseCode,
		StatusMessage: responseMessage,
		Result:        &bilRespData,
	}
	c.Data["json"] = response

	c.ServeJSON()
}

// PayStartimesBill ...
// @Title Pay Startimes Bill
// @Description create Request
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	body		body 	requests.StartimesPaymentRequest	true		"body for Request content"
// @Success 201 {int} models.Request
// @Failure 403 body is empty
// @router /pay-startimes-bill [post]
func (c *RequestController) PayStartimesBill() {
	var req requests.StartimesPaymentRequest
	json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	// Validate the request

	// authorization := ctx.Input.Header("Authorization")
	phoneNumber := c.Ctx.Input.Header("PhoneNumber")
	// sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := "501"
	responseMessage := "Request not processed"
	result := responses.BilTransactionsData{}

	// statusCode := "PENDING" // Assuming 5002 is the status code for "Request Pending"

	// transactionInt, err := strconv.ParseInt(req.TransactionId, 10, 64)
	// if err != nil {
	// 	logs.Error("Failed to parse TransactionId: %v", err)
	// }
	if transaction, err := models.GetBil_transactionsByTransactionRefNum(req.TransactionId); err == nil {
		// Go to fulfillment
		clientId, _ := strconv.ParseInt(req.ClientId, 10, 64)
		if corpInfo, err := models.GetCorporateinfoById(clientId); err == nil {
			// Formulate the request to send to the third-party service
			selectedPackage := requests.BillPaymentKeyRequest{
				Bundle: req.PackageType,
			}

			callbackurl := ""
			if cbr, err := models.GetApplication_propertyByCode("BILL_PAYMENT_CALLBACK_URL"); err == nil {
				callbackurl = cbr.PropertyValue
			} else {
				logs.Error("Failed to get callback URL: %v", err)
			}

			billerCode := "STARTIMES"
			biller, err := models.GetBillerByCode(billerCode)

			if err == nil {
				tReq := requests.BillPaymentThirdPartyRequest{
					Amount:          req.Amount,
					Destination:     req.DestinationAccount,
					ClientReference: transaction.TransactionRefNumber,
					CallbackUrl:     callbackurl,
					ExtraData:       selectedPackage,
					ServiceId:       biller.BillerReferenceId,
					PrepaidId:       corpInfo.PrepaidId,
				}

				logs.Info("Processing bill payment with third-party service: ", tReq)

				reqText, err := json.Marshal(tReq)
				if err != nil {
					logs.Error("Failed to marshal request text: %v", err)
				}

				insTransaction := models.Bil_ins_transactions{
					BilTransactionId:       transaction,
					Amount:                 req.Amount,
					Biller:                 biller,
					SenderAccountNumber:    phoneNumber,
					RecipientAccountNumber: req.DestinationAccount,
					Status:                 transaction.Status,
					Network:                billerCode,
					Request:                string(reqText),
					DateCreated:            time.Now(),
					DateModified:           time.Now(),
					CreatedBy:              1,
					ModifiedBy:             1,
					Active:                 1,
				}

				if _, err := models.AddBil_ins_transactions(&insTransaction); err != nil {
					logs.Error("Failed to create INS transaction record: %v", err)
					responseCode = "500"
					responseMessage = "Failed to create INS transaction record"
				}
				logs.Info("Processing bill payment with third-party service: ", tReq)
				if thirdPartyResponse, err := thirdparty.ProcessBillPayment(&c.Controller, tReq); err == nil {

					if thirdPartyResponse.ResponseCode == "0001" {
						responseCode = "200"
						responseMessage = "Request is being processed"
						if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "PENDING:: Failed to update transaction status"
							} else {
								responseCode = "200"
								responseMessage = "Request is being processed"
							}
						} else {
							logs.Error("Failed to get status for pending transaction: %v", err)
							responseCode = "506"
							responseMessage = "PENDING: Failed to get status for pending transaction"
						}
					} else if thirdPartyResponse.ResponseCode == "0000" {
						responseCode = "200"
						responseMessage = "Request is successful"
						if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "SUCCESS:: Failed to update transaction status"
							} else {
								logs.Info("Transaction successful: ", transaction)
								responseCode = "200"
								responseMessage = "Transaction successful"
							}
						} else {
							logs.Error("Failed to get status for successful transaction: %v", err)
							responseCode = "506"
							responseMessage = "SUCCESS:: Failed to get status for successful transaction"
						}
					} else {
						responseCode = "500"
						responseMessage = "Transaction failed"
						if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "500"
								responseMessage = "FAILED:: Failed to update transaction status"
							} else {
								logs.Info("Transaction failed: ", transaction)
								responseCode = "502"
								responseMessage = "Transaction failed"
							}
						} else {
							logs.Error("Failed to get status for failed transaction: %v", err)
							responseCode = "506"
							responseMessage = "FAILED:: Failed to get status for failed transaction"
						}
					}

					resText, err := json.Marshal(thirdPartyResponse)
					if err != nil {
						logs.Error("Failed to marshal response text: %v", err)
					}

					if v, err := models.GetRequestById(transaction.Request.RequestId); err == nil {
						v.RequestResponse = string(resText)
						v.DateModified = time.Now()
						if err := models.UpdateRequestById(v); err != nil {
							logs.Error("Failed to update request response: %v", err)
							responseCode = "200"
							responseMessage = "Success response:: Failed to update request response"
						} else {
							logs.Info("Request response updated successfully")
						}
					} else {
						logs.Error("Failed to get request by ID: %v", err)
					}

					c.Ctx.Output.SetStatus(200)

					statusDesc := "PENDING"
					if transaction.Status, err = models.GetStatus_codesByCode("IN_PROGRESS"); err == nil {
						statusDesc = transaction.Status.StatusDescription
					}

					amountStr := fmt.Sprintf("%.2f", transaction.Amount)
					commisionStr := fmt.Sprintf("%.2f", thirdPartyResponse.Data.Meta.Commission)

					result = responses.BilTransactionsData{
						TransactionId:           transaction.TransactionId,
						TransactionRefNumber:    transaction.TransactionRefNumber,
						Service:                 transaction.Service.ServiceName,
						Amount:                  amountStr,
						Request:                 transaction.Request.RequestId,
						TransactionBy:           transaction.TransactionBy.FullName,
						TransactingCurrency:     transaction.TransactingCurrency,
						SourceChannel:           transaction.SourceChannel,
						Destination:             req.DestinationAccount,
						Charge:                  thirdPartyResponse.Data.Meta.Commission,
						BillerName:              biller.BillerName,
						NetworkName:             req.PackageType,
						Commission:              commisionStr,
						ExternalReferenceNumber: thirdPartyResponse.Data.TransactionId,
						Status:                  statusDesc,
						DateCreated:             transaction.DateCreated.Format("2006-01-02 15:04:05"),
						DateModified:            transaction.DateModified.Format("2006-01-02 15:04:05"),
						CreatedBy:               1,
						ModifiedBy:              1,
						Active:                  1,
					}
				} else {
					logs.Error("Failed to process "+billerCode+" request: %v", err)
					responseCode = "500"
					responseMessage = "Failed to process " + billerCode + " request"
				}
			} else {
				logs.Error("Failed to get biller by code: %v", err)
				responseCode = "504"
				responseMessage = "Failed to get biller by code"
			}
		} else {
			logs.Error("Failed to get corporate info: %v", err)
			responseCode = "505"
			responseMessage = "Failed to get corporate info"
		}

	} else {
		logs.Error("Failed to get transaction by ID: %v", err)
		responseCode = "500"
		responseMessage = "Failed to get transaction by ID"
	}

	resp := responses.BilTransactionsResponse{
		StatusCode:    responseCode,
		StatusMessage: responseMessage,
		Result:        &result,
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// PayGoTVBill ...
// @Title Pay GoTV Bill
// @Description create Request
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	body		body 	requests.GOTVPaymentRequest	true		"body for Request content"
// @Success 201 {int} models.Request
// @Failure 403 body is empty
// @router /pay-gotv-bill [post]
func (c *RequestController) PayGoTVBill() {
	var req requests.GOTVPaymentRequest
	json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	// Validate the request

	// authorization := ctx.Input.Header("Authorization")
	phoneNumber := c.Ctx.Input.Header("PhoneNumber")
	// sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := "501"
	responseMessage := "Request not processed"
	result := responses.BilTransactionsData{}

	// transactionInt, err := strconv.ParseInt(req.TransactionId, 10, 64)
	// if err != nil {
	// 	logs.Error("Failed to parse TransactionId: %v", err)
	// }
	if transaction, err := models.GetBil_transactionsByTransactionRefNum(req.TransactionId); err == nil {
		// Go to fulfillment
		clientId, _ := strconv.ParseInt(req.ClientId, 10, 64)
		if corpInfo, err := models.GetCorporateinfoById(clientId); err == nil {
			// Formulate the request to send to the third-party service
			selectedPackage := requests.BillPaymentKeyRequest{
				Bundle: req.PackageType,
			}

			callbackurl := ""
			if cbr, err := models.GetApplication_propertyByCode("BILL_PAYMENT_CALLBACK_URL"); err == nil {
				callbackurl = cbr.PropertyValue
			} else {
				logs.Error("Failed to get callback URL: %v", err)
			}

			billerCode := "GOTV"
			biller, err := models.GetBillerByCode(billerCode)

			if err == nil {
				tReq := requests.BillPaymentThirdPartyRequest{
					Amount:          req.Amount,
					Destination:     req.DestinationAccount,
					ClientReference: transaction.TransactionRefNumber, // Use the request ID as the transaction ID
					CallbackUrl:     callbackurl,                      // Optional field for callback URL
					ExtraData:       selectedPackage,                  // Assuming this is the bundle key request
					ServiceId:       biller.BillerReferenceId,
					PrepaidId:       corpInfo.PrepaidId,
				}

				// Call the third-party service to process the request
				logs.Info("Processing gotv bill payment with third-party service: ", tReq)

				// Insert in INS Transactions table
				reqText, err := json.Marshal(tReq)
				if err != nil {
					logs.Error("Failed to marshal request text: %v", err)
				}

				insTransaction := models.Bil_ins_transactions{
					BilTransactionId:       transaction,
					Amount:                 req.Amount,
					Biller:                 biller,
					SenderAccountNumber:    phoneNumber,
					RecipientAccountNumber: req.DestinationAccount,
					Status:                 transaction.Status,
					Network:                billerCode,
					Request:                string(reqText),
					DateCreated:            time.Now(),
					DateModified:           time.Now(),
					CreatedBy:              1,
					ModifiedBy:             1,
					Active:                 1,
				}

				if _, err := models.AddBil_ins_transactions(&insTransaction); err != nil {
					logs.Error("Failed to create INS transaction record: %v", err)
					responseCode = "500"
					responseMessage = "Failed to create INS transaction record"
				}
				logs.Info("Processing bill payment with third-party service: ", tReq)
				if thirdPartyResponse, err := thirdparty.ProcessBillPayment(&c.Controller, tReq); err == nil {

					if thirdPartyResponse.ResponseCode == "0001" {
						// Transaction is pending
						responseCode = "200"
						responseMessage = "Request is being processed"
						if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "PENDING:: Failed to update transaction status"
							} else {
								responseCode = "200"
								responseMessage = "Request is being processed"
							}
						} else {
							logs.Error("Failed to get status for pending transaction: %v", err)
							responseCode = "506"
							responseMessage = "PENDING: Failed to get status for pending transaction"
						}
					} else if thirdPartyResponse.ResponseCode == "0000" {
						// Transaction is successful
						responseCode = "200"
						responseMessage = "Request is successful"
						if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "SUCCESS:: Failed to update transaction status"
							} else {
								logs.Info("Transaction successful: ", transaction)
								responseCode = "200"
								responseMessage = "Transaction successful"
							}
						} else {
							logs.Error("Failed to get status for successful transaction: %v", err)
							responseCode = "506"
							responseMessage = "SUCCESS:: Failed to get status for successful transaction"
						}
					} else {
						// Transaction failed
						responseCode = "500"
						responseMessage = "Transaction failed"
						if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "500"
								responseMessage = "FAILED:: Failed to update transaction status"
							} else {
								logs.Info("Transaction failed: ", transaction)
								responseCode = "502"
								responseMessage = "Transaction failed"
							}
						} else {
							logs.Error("Failed to get status for failed transaction: %v", err)
							responseCode = "506"
							responseMessage = "FAILED:: Failed to get status for failed transaction"
						}
					}

					resText, err := json.Marshal(thirdPartyResponse)
					if err != nil {
						logs.Error("Failed to marshal response text: %v", err)
					}

					if v, err := models.GetRequestById(transaction.Request.RequestId); err == nil {
						v.RequestResponse = string(resText)
						v.DateModified = time.Now()
						if err := models.UpdateRequestById(v); err != nil {
							logs.Error("Failed to update request response: %v", err)
							responseCode = "200"
							responseMessage = "Success response:: Failed to update request response"
						} else {
							logs.Info("Request response updated successfully")
						}
					} else {
						logs.Error("Failed to get request by ID: %v", err)
					}

					c.Ctx.Output.SetStatus(200)

					statusDesc := "PENDING"
					if transaction.Status, err = models.GetStatus_codesByCode("IN_PROGRESS"); err == nil {
						statusDesc = transaction.Status.StatusDescription
					}

					amountStr := fmt.Sprintf("%.2f", transaction.Amount)
					commisionStr := fmt.Sprintf("%.2f", thirdPartyResponse.Data.Meta.Commission)

					result = responses.BilTransactionsData{
						TransactionId:           transaction.TransactionId,
						TransactionRefNumber:    transaction.TransactionRefNumber,
						Service:                 transaction.Service.ServiceName,
						Amount:                  amountStr,
						Request:                 transaction.Request.RequestId,
						TransactionBy:           transaction.TransactionBy.FullName,
						TransactingCurrency:     transaction.TransactingCurrency,
						SourceChannel:           transaction.SourceChannel,
						Destination:             req.DestinationAccount,
						Charge:                  thirdPartyResponse.Data.Meta.Commission,
						BillerName:              biller.BillerName,
						NetworkName:             req.PackageType,
						Commission:              commisionStr,
						ExternalReferenceNumber: thirdPartyResponse.Data.TransactionId,
						Status:                  statusDesc,
						DateCreated:             transaction.DateCreated.Format("2006-01-02 15:04:05"),
						DateModified:            transaction.DateModified.Format("2006-01-02 15:04:05"),
						CreatedBy:               1,
						ModifiedBy:              1,
						Active:                  1,
					}
				} else {
					logs.Error("Failed to process "+billerCode+" request: %v", err)
					responseCode = "500"
					responseMessage = "Failed to process " + billerCode + " request"
				}
			} else {
				logs.Error("Failed to get biller by code: %v", err)
				responseCode = "504"
				responseMessage = "Failed to get biller by code"
			}
		} else {
			logs.Error("Failed to get corporate info: %v", err)
			responseCode = "505"
			responseMessage = "Failed to get corporate info"
		}
	} else {
		logs.Error("Failed to get transaction by ID: %v", err)
		responseCode = "500"
		responseMessage = "Failed to get transaction by ID"
	}

	response := responses.BilTransactionsResponse{
		StatusCode:    responseCode,
		StatusMessage: responseMessage,
		Result:        &result,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// PayWaterBill ...
// @Title Pay Water Bill
// @Description create Request
// @Param	PhoneNumber		header 	string true		"header for Customer's phone number"
// @Param	SourceSystem		header 	string true		"header for Source system"
// @Param	body		body 	requests.DSTVPaymentRequest	true		"body for Request content"
// @Success 201 {int} models.Request
// @Failure 403 body is empty
// @router /pay-water-bill [post]
func (c *RequestController) PayWaterBill() {
	var req requests.GhanaWaterPaymentRequest
	json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	// Validate the request

	// authorization := ctx.Input.Header("Authorization")
	phoneNumber := c.Ctx.Input.Header("PhoneNumber")
	// sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := "501"
	responseMessage := "Request not processed"
	result := responses.BilTransactionsData{}
	// transactionInt, err := strconv.ParseInt(req.TransactionId, 10, 64)
	// if err != nil {
	// 	logs.Error("Failed to parse TransactionId: %v", err)
	// }
	if transaction, err := models.GetBil_transactionsByTransactionRefNum(req.TransactionId); err == nil {
		// Go to fulfillment
		clientId, _ := strconv.ParseInt(req.ClientId, 10, 64)
		if corpInfo, err := models.GetCorporateinfoById(clientId); err == nil {
			// Formulate the request to send to the third-party service
			selectedPackage := requests.GhanaWaterBillPaymentKeyRequest{
				Bundle:    req.Bundle,
				SessionId: req.SessionId, // Assuming this is the session ID for the request
				Email:     req.Email,     // Assuming this is the email for the request
			}

			callbackurl := ""
			if cbr, err := models.GetApplication_propertyByCode("BILL_PAYMENT_CALLBACK_URL"); err == nil {
				callbackurl = cbr.PropertyValue
			} else {
				logs.Error("Failed to get callback URL: %v", err)
			}

			billerCode := "GH_WATER"
			biller, err := models.GetBillerByCode(billerCode)

			if err == nil {
				tReq := requests.GhanaWaterBillPaymentThirdPartyRequest{
					Amount:          req.Amount,
					Destination:     req.DestinationAccount,
					ClientReference: transaction.TransactionRefNumber,
					CallbackUrl:     callbackurl,
					ExtraData:       selectedPackage,
					ServiceId:       biller.BillerReferenceId,
					PrepaidId:       corpInfo.PrepaidId,
				}

				logs.Info("Processing bill payment with third-party service: ", tReq)

				// Insert in INS Transactions table
				reqText, err := json.Marshal(tReq)
				if err != nil {
					logs.Error("Failed to marshal request text: %v", err)
				}

				insTransaction := models.Bil_ins_transactions{
					BilTransactionId:       transaction,
					Amount:                 req.Amount,
					Biller:                 biller,
					SenderAccountNumber:    phoneNumber,
					RecipientAccountNumber: req.DestinationAccount,
					Status:                 transaction.Status,
					Network:                billerCode,
					Request:                string(reqText),
					DateCreated:            time.Now(),
					DateModified:           time.Now(),
					CreatedBy:              1,
					ModifiedBy:             1,
					Active:                 1,
				}

				if _, err := models.AddBil_ins_transactions(&insTransaction); err != nil {
					logs.Error("Failed to create INS transaction record: %v", err)
					responseCode = "500"
					responseMessage = "Failed to create INS transaction record"
				}
				logs.Info("Processing bill payment with third-party service: ", tReq)
				if thirdPartyResponse, err := thirdparty.ProcessGhanaWaterBillPayment(&c.Controller, tReq); err == nil {

					if thirdPartyResponse.ResponseCode == "0001" {
						responseCode = "200"
						responseMessage = "Request is being processed"
						if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "PENDING:: Failed to update transaction status"
							} else {
								responseCode = "200"
								responseMessage = "Request is being processed"
							}
						} else {
							logs.Error("Failed to get status for pending transaction: %v", err)
							responseCode = "506"
							responseMessage = "PENDING: Failed to get status for pending transaction"
						}
					} else if thirdPartyResponse.ResponseCode == "0000" {
						responseCode = "200"
						responseMessage = "Request is successful"
						if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "200"
								responseMessage = "SUCCESS:: Failed to update transaction status"
							} else {
								logs.Info("Transaction successful: ", transaction)
								responseCode = "200"
								responseMessage = "Transaction successful"
							}
						} else {
							logs.Error("Failed to get status for successful transaction: %v", err)
							responseCode = "506"
							responseMessage = "SUCCESS:: Failed to get status for successful transaction"
						}
					} else {
						responseCode = "500"
						responseMessage = "Transaction failed"
						if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
							transaction.Status = status
							if err := models.UpdateBil_transactionsById(transaction); err != nil {
								logs.Error("Failed to update transaction status: %v", err)
								responseCode = "500"
								responseMessage = "FAILED:: Failed to update transaction status"
							} else {
								logs.Info("Transaction failed: ", transaction)
								responseCode = "502"
								responseMessage = "Transaction failed"
							}
						} else {
							logs.Error("Failed to get status for failed transaction: %v", err)
							responseCode = "506"
							responseMessage = "FAILED:: Failed to get status for failed transaction"
						}
					}

					resText, err := json.Marshal(thirdPartyResponse)
					if err != nil {
						logs.Error("Failed to marshal response text: %v", err)
					}

					if v, err := models.GetRequestById(transaction.Request.RequestId); err == nil {
						v.RequestResponse = string(resText)
						v.DateModified = time.Now()
						if err := models.UpdateRequestById(v); err != nil {
							logs.Error("Failed to update request response: %v", err)
							responseCode = "200"
							responseMessage = "Success response:: Failed to update request response"
						} else {
							logs.Info("Request response updated successfully")
						}
					} else {
						logs.Error("Failed to get request by ID: %v", err)
					}

					c.Ctx.Output.SetStatus(200)

					statusDesc := "PENDING"
					if transaction.Status, err = models.GetStatus_codesByCode("IN_PROGRESS"); err == nil {
						statusDesc = transaction.Status.StatusDescription
					}

					amountStr := fmt.Sprintf("%.2f", transaction.Amount)
					commisionStr := fmt.Sprintf("%.2f", thirdPartyResponse.Data.Meta.Commission)

					result = responses.BilTransactionsData{
						TransactionId:           transaction.TransactionId,
						TransactionRefNumber:    transaction.TransactionRefNumber,
						Service:                 transaction.Service.ServiceName,
						Amount:                  amountStr,
						Request:                 transaction.Request.RequestId,
						TransactionBy:           transaction.TransactionBy.FullName,
						TransactingCurrency:     transaction.TransactingCurrency,
						SourceChannel:           transaction.SourceChannel,
						Destination:             req.DestinationAccount,
						Charge:                  thirdPartyResponse.Data.Meta.Commission,
						BillerName:              biller.BillerName,
						NetworkName:             req.Bundle,
						Commission:              commisionStr,
						ExternalReferenceNumber: thirdPartyResponse.Data.TransactionId,
						Status:                  statusDesc,
						DateCreated:             transaction.DateCreated.Format("2006-01-02 15:04:05"),
						DateModified:            transaction.DateModified.Format("2006-01-02 15:04:05"),
						CreatedBy:               1,
						ModifiedBy:              1,
						Active:                  1,
					}
				} else {
					logs.Error("Failed to process "+billerCode+" request: %v", err)
					responseCode = "500"
					responseMessage = "Failed to process " + billerCode + " request"
				}
			} else {
				logs.Error("Failed to get biller by code: %v", err)
				responseCode = "504"
				responseMessage = "Failed to get biller by code"
			}
		} else {
			logs.Error("Failed to get corporate info: %v", err)
			responseCode = "505"
			responseMessage = "Failed to get corporate info"
		}
	} else {
		logs.Error("Failed to get transaction by ID: %v", err)
		responseCode = "500"
		responseMessage = "Failed to get transaction by ID"
	}

	response := responses.BilTransactionsResponse{
		StatusCode:    responseCode,
		StatusMessage: responseMessage,
		Result:        &result,
	}
	c.Data["json"] = response

	c.ServeJSON()
}

// BilTransactions ...
// @Title Airtime and Data Bundle Transactions
// @Description get Request
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.Request
// @Failure 403
// @router /bil-transactions/ [get]
func (c *RequestController) BilTransactions() {
	var fields []string
	var sortby []string
	var order []string
	var query = make(map[string]string)
	var limit int64 = 12
	var offset int64

	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	// limit: 10 (default is 10)
	if v, err := c.GetInt64("limit"); err == nil {
		limit = v
	}
	// offset: 0 (default is 0)
	if v, err := c.GetInt64("offset"); err == nil {
		offset = v
	}
	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		sortby = strings.Split(v, ",")
	}
	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		order = strings.Split(v, ",")
	}
	// query: k:v,k:v
	if v := c.GetString("query"); v != "" {
		for _, cond := range strings.Split(v, ",") {
			kv := strings.SplitN(cond, ":", 2)
			if len(kv) != 2 {
				c.Data["json"] = errors.New("error: invalid query key/value pair")
				c.ServeJSON()
				return
			}
			k, v := kv[0], kv[1]
			query[k] = v
		}
	}

	query["BilTransactionId__Service__ServiceCode__in"] = "BILL_PAYMENT,AIRTIME,DATA_BUNDLE"
	// query["BilTransactionId__Service__ServiceCode__in"] = "BILL_PAYMENT","AIRTIME","DATA_BUNDLE"

	statusCode := "500"
	statusMessage := "No records found"

	bilTransactions := []*responses.BilTransactionsData{}

	l, err := models.GetAllBil_ins_transactions(query, fields, sortby, order, offset, limit)
	if err != nil {
		logs.Error("Error fetching records: ", err)

		statusCode = "500"
		statusMessage = "An error occurred" + err.Error()
	} else {
		logs.Info("Records fetched: ", l)
		if len(l) > 0 {
			for _, record := range l {
				logs.Info("Record: ", record)
				bilTxn, ok := record.(models.Bil_ins_transactions)
				if ok {
					bilTxnJson, _ := json.MarshalIndent(bilTxn, "", "  ")
					logs.Info("Bil_ins_transaction: %s", string(bilTxnJson))
					// logs.Info("BilTransactionId: %+v", bilTxn.BilTransactionId)

					amountString := fmt.Sprintf("%.2f", bilTxn.BilTransactionId.Amount)
					chargeString := fmt.Sprintf("%.2f", bilTxn.BilTransactionId.Charge)
					bilTransaction := responses.BilTransactionsData{
						TransactionId:           bilTxn.BilInsTransactionId,
						TransactionRefNumber:    bilTxn.BilTransactionId.TransactionRefNumber,
						Service:                 bilTxn.BilTransactionId.Service.ServiceName,
						TransactionBy:           bilTxn.BilTransactionId.TransactionBy.FullName,
						Amount:                  amountString,
						TransactingCurrency:     bilTxn.BilTransactionId.TransactingCurrency,
						SourceChannel:           bilTxn.BilTransactionId.SourceChannel,
						Source:                  bilTxn.SenderAccountNumber,
						Destination:             bilTxn.RecipientAccountNumber,
						Charge:                  chargeString,
						Status:                  bilTxn.BilTransactionId.Status.StatusCode,
						DateCreated:             bilTxn.DateCreated.Format(time.RFC3339),
						DateModified:            bilTxn.DateModified.Format(time.RFC3339),
						CreatedBy:               bilTxn.CreatedBy,
						ModifiedBy:              bilTxn.ModifiedBy,
						Active:                  bilTxn.Active,
						BillerName:              bilTxn.Biller.BillerName,
						NetworkName:             bilTxn.Network,
						ExternalReferenceNumber: bilTxn.BilTransactionId.ExternalReferenceNumber,
					}

					bilTransactions = append(bilTransactions, &bilTransaction)

				}
			}

			statusCode = "200"
			statusMessage = "Records found"
		} else {
			statusCode = "204"
			statusMessage = "No records found"
		}
	}

	response := responses.BilTransactionsListResponse{
		StatusCode:    statusCode,
		StatusMessage: statusMessage,
		Result:        bilTransactions,
	}
	c.Data["json"] = response
	c.ServeJSON()
}
