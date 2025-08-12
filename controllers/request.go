package controllers

import (
	thirdparty "billpayment_service/controllers/thirdParty"
	"billpayment_service/models"
	"billpayment_service/structs/requests"
	"billpayment_service/structs/responses"
	"encoding/json"
	"strconv"
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
	sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := false
	responseMessage := "Request not processed"

	statusCode := "PENDING" // Assuming 5002 is the status code for "Request Pending"

	reqText, err := json.Marshal(req)
	if err != nil {
		c.Data["json"] = "Invalid request format"
		c.ServeJSON()
		return
	}

	status, err := models.GetStatus_codesByCode(statusCode)
	if err == nil {
		// Get customer by ID
		if cust, err := models.GetCustomerByPhoneNumber(phoneNumber); err == nil {
			// Restructure the request to match the model
			serviceCode := "BILL_PAYMENT"
			if service, err := models.GetServicesByCode(serviceCode); err == nil {
				v := models.Request{
					RequestId:       req.RequestId,
					CustId:          cust,
					Request:         string(reqText),
					RequestType:     service.ServiceName,
					RequestStatus:   status.StatusDescription,
					RequestAmount:   req.Amount,
					RequestResponse: "",
					RequestDate:     time.Now(),
					DateCreated:     time.Now(),
					DateModified:    time.Now(),
				}
				if _, err := models.AddRequest(&v); err == nil {
					// Create a transaction record
					transaction := models.Bil_transactions{
						TransactionRefNumber: "TRX-" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.FormatInt(v.RequestId, 10),
						Service:              service, // Assuming service ID is 1 for airtime
						Request:              &v,
						TransactionBy:        cust,
						Amount:               req.Amount,
						TransactingCurrency:  "GHC", // Assuming USD for simplicity
						SourceChannel:        sourceSystem,
						Source:               phoneNumber,
						Destination:          req.DestinationAccount,
						Charge:               0.0,    // Assuming no charge for simplicity
						Status:               status, // Assuming 1 means successful
						DateCreated:          time.Now(),
						DateModified:         time.Now(),
						CreatedBy:            1,
						ModifiedBy:           1,
						Active:               1, // Assuming active status
					}
					if _, err := models.AddBil_transactions(&transaction); err == nil {
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
							tReq := requests.ThirdPartyDSTVPaymentRequest{
								Amount:          req.Amount,
								Destination:     req.DestinationAccount,
								ClientReference: transaction.TransactionRefNumber, // Use the request ID as the transaction ID
								CallbackUrl:     callbackurl,                      // Optional field for callback URL
								ExtraData:       selectedPackage,                  // Assuming this is the bundle key request
								ServiceId:       biller.BillerReferenceId,
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
								BilTransactionId:       &transaction,
								Amount:                 req.Amount,
								Biller:                 biller,
								SenderAccountNumber:    phoneNumber,
								RecipientAccountNumber: req.DestinationAccount,
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
								responseCode = false
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
									responseCode = true
									responseMessage = "Request is being processed"
									if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "PENDING:: Failed to update transaction status"
										} else {
											responseCode = true
											responseMessage = "Request is being processed"
										}
									} else {
										logs.Error("Failed to get status for pending transaction: %v", err)
										responseCode = false
										responseMessage = "PENDING: Failed to get status for pending transaction"
									}
								} else if thirdPartyResponse.ResponseCode == "0000" {
									// Transaction is successful
									// Update the transaction status to successful
									responseCode = true
									responseMessage = "Request is successful"
									if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "SUCCESS:: Failed to update transaction status"
										} else {
											// Prepare the response
											logs.Info("Transaction successful: ", transaction)
											responseCode = true
											responseMessage = "Transaction successful"
										}
									} else {
										logs.Error("Failed to get status for successful transaction: %v", err)
										responseCode = false
										responseMessage = "SUCCESS:: Failed to get status for successful transaction"
									}
								} else {
									// Transaction failed
									// Update the transaction status to failed
									responseCode = false
									responseMessage = "Transaction failed"
									if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "FAILED:: Failed to update transaction status"
										}
									} else {
										logs.Error("Failed to get status for failed transaction: %v", err)
										responseCode = false
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
								v.RequestResponse = string(resText)
								v.DateModified = time.Now()
								if err := models.UpdateRequestById(&v); err != nil {
									logs.Error("Failed to update request response: %v", err)
									responseCode = true
									responseMessage = "Success response:: Failed to update request response"
								} else {
									logs.Info("Request response updated successfully")
								}

								c.Ctx.Output.SetStatus(200)
								// Prepare the response

								// Create the response object
								respData := responses.DSTVBillPaymentDataResponse{
									Description:   "Payment for DSTV bill",
									Amount:        req.Amount,
									TransactionId: transaction.TransactionRefNumber,
								}
								response := responses.DSTVBillPaymentResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        &respData,
								}
								c.Data["json"] = response
							} else {
								logs.Error("Failed to process dstv request: %v", err)
								responseCode = false
								responseMessage = "Failed to process dstv request"
								resp := responses.DSTVBillPaymentResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        nil,
								}
								c.Data["json"] = resp
							}
						} else {
							logs.Error("Failed to get biller by code: %v", err)
							responseCode = false
							responseMessage = "Failed to get biller by code"
							resp := responses.DSTVBillPaymentResponse{
								StatusCode:    responseCode,
								StatusMessage: responseMessage,
								Result:        nil,
							}
							c.Data["json"] = resp
						}

					} else {
						logs.Error("Failed to create transaction record: %v", err)
						responseCode = false
						responseMessage = "Failed to create transaction record"
						resp := responses.DSTVBillPaymentResponse{
							StatusCode:    responseCode,
							StatusMessage: responseMessage,
							Result:        nil,
						}
						c.Data["json"] = resp
					}
				} else {
					logs.Error("Failed to create request record: %v", err)
					responseCode = false
					responseMessage = "Failed to create request log"
					resp := responses.DSTVBillPaymentResponse{
						StatusCode:    responseCode,
						StatusMessage: responseMessage,
						Result:        nil,
					}
					c.Data["json"] = resp
				}
			} else {
				logs.Error("Service not found: %v", err)
				responseCode = false
				responseMessage = "Failed to create transaction record. Service not found."
				resp := responses.DSTVBillPaymentResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
			}
		} else {
			logs.Error("Customer not found: %v", err)
			responseCode = false
			responseMessage = "Failed to create transaction record"
			resp := responses.DSTVBillPaymentResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		}
	} else {
		logs.Error("Status not found: %v", err)
		responseCode = false
		responseMessage = "Failed to create transaction record"
		resp := responses.DSTVBillPaymentResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	}
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
// @router /ghana-water-account-query/:accountNumber [get]
func (c *RequestController) GhanaWaterAccountQuery() {
	accountNumberStr := c.Ctx.Input.Param(":accountNumber")

	logs.Info("Received request to query Ghana water account with number: ", accountNumberStr)

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
	sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := false
	responseMessage := "Request not processed"

	statusCode := "PENDING" // Assuming 5002 is the status code for "Request Pending"

	reqText, err := json.Marshal(req)
	if err != nil {
		c.Data["json"] = "Invalid request format"
		c.ServeJSON()
		return
	}
	logs.Info("Received request to pay ECG bill: ", string(reqText))

	status, err := models.GetStatus_codesByCode(statusCode)
	if err == nil {
		// Get customer by ID
		if cust, err := models.GetCustomerByPhoneNumber(phoneNumber); err == nil {
			// Restructure the request to match the model
			serviceCode := "BILL_PAYMENT"
			if service, err := models.GetServicesByCode(serviceCode); err == nil {
				v := models.Request{
					RequestId:       req.RequestId,
					CustId:          cust,
					Request:         string(reqText),
					RequestType:     service.ServiceName,
					RequestStatus:   status.StatusDescription,
					RequestAmount:   req.Amount,
					RequestResponse: "",
					RequestDate:     time.Now(),
					DateCreated:     time.Now(),
					DateModified:    time.Now(),
				}
				if _, err := models.AddRequest(&v); err == nil {
					// Create a transaction record
					transaction := models.Bil_transactions{
						TransactionRefNumber: "TRX-" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.FormatInt(v.RequestId, 10),
						Service:              service, // Assuming service ID is 1 for airtime
						Request:              &v,
						TransactionBy:        cust,
						Amount:               req.Amount,
						TransactingCurrency:  "GHC", // Assuming USD for simplicity
						SourceChannel:        sourceSystem,
						Source:               phoneNumber,
						Destination:          req.DestinationAccount,
						Charge:               0.0,    // Assuming no charge for simplicity
						Status:               status, // Assuming 1 means successful
						DateCreated:          time.Now(),
						DateModified:         time.Now(),
						CreatedBy:            1,
						ModifiedBy:           1,
						Active:               1, // Assuming active status
					}
					if _, err := models.AddBil_transactions(&transaction); err == nil {
						// Go to fulfillment
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
								BilTransactionId:       &transaction,
								Amount:                 req.Amount,
								Biller:                 biller,
								SenderAccountNumber:    phoneNumber,
								RecipientAccountNumber: req.DestinationAccount,
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
								responseCode = false
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
									responseCode = true
									responseMessage = "Request is being processed"
									if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "PENDING:: Failed to update transaction status"
										} else {
											responseCode = true
											responseMessage = "Request is being processed"
										}
									} else {
										logs.Error("Failed to get status for pending transaction: %v", err)
										responseCode = false
										responseMessage = "PENDING: Failed to get status for pending transaction"
									}
								} else if thirdPartyResponse.ResponseCode == "0000" {
									// Transaction is successful
									// Update the transaction status to successful
									responseCode = true
									responseMessage = "Request is successful"
									if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "SUCCESS:: Failed to update transaction status"
										} else {
											// Prepare the response
											logs.Info("Transaction successful: ", transaction)
											responseCode = true
											responseMessage = "Transaction successful"
										}
									} else {
										logs.Error("Failed to get status for successful transaction: %v", err)
										responseCode = false
										responseMessage = "SUCCESS:: Failed to get status for successful transaction"
									}
								} else {
									// Transaction failed
									// Update the transaction status to failed
									responseCode = false
									responseMessage = "Transaction failed"
									if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "FAILED:: Failed to update transaction status"
										}
									} else {
										logs.Error("Failed to get status for failed transaction: %v", err)
										responseCode = false
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
								v.RequestResponse = string(resText)
								v.DateModified = time.Now()
								if err := models.UpdateRequestById(&v); err != nil {
									logs.Error("Failed to update request response: %v", err)
									responseCode = true
									responseMessage = "Success response:: Failed to update request response"
								} else {
									logs.Info("Request response updated successfully")
								}

								c.Ctx.Output.SetStatus(200)
								// Prepare the response

								// Create the response object
								respData := responses.ThirdPartyBillPaymentDataApiResponse{
									Description:   "Payment for " + billerCode + " bill",
									Amount:        req.Amount,
									TransactionId: transaction.TransactionRefNumber,
								}
								response := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        &respData,
								}
								c.Data["json"] = response
							} else {
								logs.Error("Failed to process "+billerCode+" request: %v", err)
								responseCode = false
								responseMessage = "Failed to process " + billerCode + " request"
								resp := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        nil,
								}
								c.Data["json"] = resp
							}
						} else {
							logs.Error("Failed to get biller by code: %v", err)
							responseCode = false
							responseMessage = "Failed to get biller by code"
							resp := responses.ThirdPartyBillPaymentApiResponse{
								StatusCode:    responseCode,
								StatusMessage: responseMessage,
								Result:        nil,
							}
							c.Data["json"] = resp
						}

					} else {
						logs.Error("Failed to create transaction record: %v", err)
						responseCode = false
						responseMessage = "Failed to create transaction record"
						resp := responses.ThirdPartyBillPaymentApiResponse{
							StatusCode:    responseCode,
							StatusMessage: responseMessage,
							Result:        nil,
						}
						c.Data["json"] = resp
					}
				} else {
					logs.Error("Failed to create request record: %v", err)
					responseCode = false
					responseMessage = "Failed to create request log"
					resp := responses.ThirdPartyBillPaymentApiResponse{
						StatusCode:    responseCode,
						StatusMessage: responseMessage,
						Result:        nil,
					}
					c.Data["json"] = resp
				}
			} else {
				logs.Error("Service not found: %v", err)
				responseCode = false
				responseMessage = "Failed to create transaction record. Service not found."
				resp := responses.ThirdPartyBillPaymentApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
			}
		} else {
			logs.Error("Customer not found: %v", err)
			responseCode = false
			responseMessage = "Failed to create transaction record"
			resp := responses.ThirdPartyBillPaymentApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		}
	} else {
		logs.Error("Status not found: %v", err)
		responseCode = false
		responseMessage = "Failed to create transaction record"
		resp := responses.ThirdPartyBillPaymentApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	}
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
	sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := false
	responseMessage := "Request not processed"

	statusCode := "PENDING" // Assuming 5002 is the status code for "Request Pending"

	reqText, err := json.Marshal(req)
	if err != nil {
		c.Data["json"] = "Invalid request format"
		c.ServeJSON()
		return
	}

	status, err := models.GetStatus_codesByCode(statusCode)
	if err == nil {
		// Get customer by ID
		if cust, err := models.GetCustomerByPhoneNumber(phoneNumber); err == nil {
			// Restructure the request to match the model
			serviceCode := "BILL_PAYMENT"
			if service, err := models.GetServicesByCode(serviceCode); err == nil {
				v := models.Request{
					RequestId:       req.RequestId,
					CustId:          cust,
					Request:         string(reqText),
					RequestType:     service.ServiceName,
					RequestStatus:   status.StatusDescription,
					RequestAmount:   req.Amount,
					RequestResponse: "",
					RequestDate:     time.Now(),
					DateCreated:     time.Now(),
					DateModified:    time.Now(),
				}
				if _, err := models.AddRequest(&v); err == nil {
					// Create a transaction record
					transaction := models.Bil_transactions{
						TransactionRefNumber: "TRX-" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.FormatInt(v.RequestId, 10),
						Service:              service, // Assuming service ID is 1 for airtime
						Request:              &v,
						TransactionBy:        cust,
						Amount:               req.Amount,
						TransactingCurrency:  "GHC", // Assuming USD for simplicity
						SourceChannel:        sourceSystem,
						Source:               phoneNumber,
						Destination:          req.DestinationAccount,
						Charge:               0.0,    // Assuming no charge for simplicity
						Status:               status, // Assuming 1 means successful
						DateCreated:          time.Now(),
						DateModified:         time.Now(),
						CreatedBy:            1,
						ModifiedBy:           1,
						Active:               1, // Assuming active status
					}
					if _, err := models.AddBil_transactions(&transaction); err == nil {
						// Go to fulfillment
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
								ClientReference: transaction.TransactionRefNumber, // Use the request ID as the transaction ID
								CallbackUrl:     callbackurl,                      // Optional field for callback URL
								ExtraData:       selectedPackage,                  // Assuming this is the bundle key request
								ServiceId:       biller.BillerReferenceId,
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
								BilTransactionId:       &transaction,
								Amount:                 req.Amount,
								Biller:                 biller,
								SenderAccountNumber:    phoneNumber,
								RecipientAccountNumber: req.DestinationAccount,
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
								responseCode = false
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
									responseCode = true
									responseMessage = "Request is being processed"
									if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "PENDING:: Failed to update transaction status"
										} else {
											responseCode = true
											responseMessage = "Request is being processed"
										}
									} else {
										logs.Error("Failed to get status for pending transaction: %v", err)
										responseCode = false
										responseMessage = "PENDING: Failed to get status for pending transaction"
									}
								} else if thirdPartyResponse.ResponseCode == "0000" {
									// Transaction is successful
									// Update the transaction status to successful
									responseCode = true
									responseMessage = "Request is successful"
									if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "SUCCESS:: Failed to update transaction status"
										} else {
											// Prepare the response
											logs.Info("Transaction successful: ", transaction)
											responseCode = true
											responseMessage = "Transaction successful"
										}
									} else {
										logs.Error("Failed to get status for successful transaction: %v", err)
										responseCode = false
										responseMessage = "SUCCESS:: Failed to get status for successful transaction"
									}
								} else {
									// Transaction failed
									// Update the transaction status to failed
									responseCode = false
									responseMessage = "Transaction failed"
									if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "FAILED:: Failed to update transaction status"
										}
									} else {
										logs.Error("Failed to get status for failed transaction: %v", err)
										responseCode = false
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
								v.RequestResponse = string(resText)
								v.DateModified = time.Now()
								if err := models.UpdateRequestById(&v); err != nil {
									logs.Error("Failed to update request response: %v", err)
									responseCode = true
									responseMessage = "Success response:: Failed to update request response"
								} else {
									logs.Info("Request response updated successfully")
								}

								c.Ctx.Output.SetStatus(200)
								// Prepare the response

								// Create the response object
								respData := responses.ThirdPartyBillPaymentDataApiResponse{
									Description:   "Payment for " + billerCode + " bill",
									Amount:        req.Amount,
									TransactionId: transaction.TransactionRefNumber,
								}
								response := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        &respData,
								}
								c.Data["json"] = response
							} else {
								logs.Error("Failed to process "+billerCode+" request: %v", err)
								responseCode = false
								responseMessage = "Failed to process " + billerCode + " request"
								resp := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        nil,
								}
								c.Data["json"] = resp
							}
						} else {
							logs.Error("Failed to get biller by code: %v", err)
							responseCode = false
							responseMessage = "Failed to get biller by code"
							resp := responses.ThirdPartyBillPaymentApiResponse{
								StatusCode:    responseCode,
								StatusMessage: responseMessage,
								Result:        nil,
							}
							c.Data["json"] = resp
						}

					} else {
						logs.Error("Failed to create transaction record: %v", err)
						responseCode = false
						responseMessage = "Failed to create transaction record"
						resp := responses.ThirdPartyBillPaymentApiResponse{
							StatusCode:    responseCode,
							StatusMessage: responseMessage,
							Result:        nil,
						}
						c.Data["json"] = resp
					}
				} else {
					logs.Error("Failed to create request record: %v", err)
					responseCode = false
					responseMessage = "Failed to create request log"
					resp := responses.ThirdPartyBillPaymentApiResponse{
						StatusCode:    responseCode,
						StatusMessage: responseMessage,
						Result:        nil,
					}
					c.Data["json"] = resp
				}
			} else {
				logs.Error("Service not found: %v", err)
				responseCode = false
				responseMessage = "Failed to create transaction record. Service not found."
				resp := responses.ThirdPartyBillPaymentApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
			}
		} else {
			logs.Error("Customer not found: %v", err)
			responseCode = false
			responseMessage = "Failed to create transaction record"
			resp := responses.ThirdPartyBillPaymentApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		}
	} else {
		logs.Error("Status not found: %v", err)
		responseCode = false
		responseMessage = "Failed to create transaction record"
		resp := responses.ThirdPartyBillPaymentApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	}
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
	sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := false
	responseMessage := "Request not processed"

	statusCode := "PENDING" // Assuming 5002 is the status code for "Request Pending"

	reqText, err := json.Marshal(req)
	if err != nil {
		c.Data["json"] = "Invalid request format"
		c.ServeJSON()
		return
	}

	status, err := models.GetStatus_codesByCode(statusCode)
	if err == nil {
		// Get customer by ID
		if cust, err := models.GetCustomerByPhoneNumber(phoneNumber); err == nil {
			// Restructure the request to match the model
			serviceCode := "BILL_PAYMENT"
			if service, err := models.GetServicesByCode(serviceCode); err == nil {
				v := models.Request{
					RequestId:       req.RequestId,
					CustId:          cust,
					Request:         string(reqText),
					RequestType:     service.ServiceName,
					RequestStatus:   status.StatusDescription,
					RequestAmount:   req.Amount,
					RequestResponse: "",
					RequestDate:     time.Now(),
					DateCreated:     time.Now(),
					DateModified:    time.Now(),
				}
				if _, err := models.AddRequest(&v); err == nil {
					// Create a transaction record
					transaction := models.Bil_transactions{
						TransactionRefNumber: "TRX-" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.FormatInt(v.RequestId, 10),
						Service:              service, // Assuming service ID is 1 for airtime
						Request:              &v,
						TransactionBy:        cust,
						Amount:               req.Amount,
						TransactingCurrency:  "GHC", // Assuming USD for simplicity
						SourceChannel:        sourceSystem,
						Source:               phoneNumber,
						Destination:          req.DestinationAccount,
						Charge:               0.0,    // Assuming no charge for simplicity
						Status:               status, // Assuming 1 means successful
						DateCreated:          time.Now(),
						DateModified:         time.Now(),
						CreatedBy:            1,
						ModifiedBy:           1,
						Active:               1, // Assuming active status
					}
					if _, err := models.AddBil_transactions(&transaction); err == nil {
						// Go to fulfillment
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
								BilTransactionId:       &transaction,
								Amount:                 req.Amount,
								Biller:                 biller,
								SenderAccountNumber:    phoneNumber,
								RecipientAccountNumber: req.DestinationAccount,
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
								responseCode = false
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
									responseCode = true
									responseMessage = "Request is being processed"
									if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "PENDING:: Failed to update transaction status"
										} else {
											responseCode = true
											responseMessage = "Request is being processed"
										}
									} else {
										logs.Error("Failed to get status for pending transaction: %v", err)
										responseCode = false
										responseMessage = "PENDING: Failed to get status for pending transaction"
									}
								} else if thirdPartyResponse.ResponseCode == "0000" {
									// Transaction is successful
									// Update the transaction status to successful
									responseCode = true
									responseMessage = "Request is successful"
									if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "SUCCESS:: Failed to update transaction status"
										} else {
											// Prepare the response
											logs.Info("Transaction successful: ", transaction)
											responseCode = true
											responseMessage = "Transaction successful"
										}
									} else {
										logs.Error("Failed to get status for successful transaction: %v", err)
										responseCode = false
										responseMessage = "SUCCESS:: Failed to get status for successful transaction"
									}
								} else {
									// Transaction failed
									// Update the transaction status to failed
									responseCode = false
									responseMessage = "Transaction failed"
									if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "FAILED:: Failed to update transaction status"
										}
									} else {
										logs.Error("Failed to get status for failed transaction: %v", err)
										responseCode = false
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
								v.RequestResponse = string(resText)
								v.DateModified = time.Now()
								if err := models.UpdateRequestById(&v); err != nil {
									logs.Error("Failed to update request response: %v", err)
									responseCode = true
									responseMessage = "Success response:: Failed to update request response"
								} else {
									logs.Info("Request response updated successfully")
								}

								c.Ctx.Output.SetStatus(200)
								// Prepare the response

								// Create the response object
								respData := responses.ThirdPartyBillPaymentDataApiResponse{
									Description:   "Payment for " + billerCode + " bill",
									Amount:        req.Amount,
									TransactionId: transaction.TransactionRefNumber,
								}
								response := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        &respData,
								}
								c.Data["json"] = response
							} else {
								logs.Error("Failed to process "+billerCode+" request: %v", err)
								responseCode = false
								responseMessage = "Failed to process " + billerCode + " request"
								resp := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        nil,
								}
								c.Data["json"] = resp
							}
						} else {
							logs.Error("Failed to get biller by code: %v", err)
							responseCode = false
							responseMessage = "Failed to get biller by code"
							resp := responses.ThirdPartyBillPaymentApiResponse{
								StatusCode:    responseCode,
								StatusMessage: responseMessage,
								Result:        nil,
							}
							c.Data["json"] = resp
						}

					} else {
						logs.Error("Failed to create transaction record: %v", err)
						responseCode = false
						responseMessage = "Failed to create transaction record"
						resp := responses.ThirdPartyBillPaymentApiResponse{
							StatusCode:    responseCode,
							StatusMessage: responseMessage,
							Result:        nil,
						}
						c.Data["json"] = resp
					}
				} else {
					logs.Error("Failed to create request record: %v", err)
					responseCode = false
					responseMessage = "Failed to create request log"
					resp := responses.ThirdPartyBillPaymentApiResponse{
						StatusCode:    responseCode,
						StatusMessage: responseMessage,
						Result:        nil,
					}
					c.Data["json"] = resp
				}
			} else {
				logs.Error("Service not found: %v", err)
				responseCode = false
				responseMessage = "Failed to create transaction record. Service not found."
				resp := responses.ThirdPartyBillPaymentApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
			}
		} else {
			logs.Error("Customer not found: %v", err)
			responseCode = false
			responseMessage = "Failed to create transaction record"
			resp := responses.ThirdPartyBillPaymentApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		}
	} else {
		logs.Error("Status not found: %v", err)
		responseCode = false
		responseMessage = "Failed to create transaction record"
		resp := responses.ThirdPartyBillPaymentApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	}
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
	sourceSystem := c.Ctx.Input.Header("SourceSystem")

	responseCode := false
	responseMessage := "Request not processed"

	statusCode := "PENDING" // Assuming 5002 is the status code for "Request Pending"

	reqText, err := json.Marshal(req)
	if err != nil {
		c.Data["json"] = "Invalid request format"
		c.ServeJSON()
		return
	}

	status, err := models.GetStatus_codesByCode(statusCode)
	if err == nil {
		// Get customer by ID
		if cust, err := models.GetCustomerByPhoneNumber(phoneNumber); err == nil {
			// Restructure the request to match the model
			serviceCode := "BILL_PAYMENT"
			if service, err := models.GetServicesByCode(serviceCode); err == nil {
				v := models.Request{
					RequestId:       req.RequestId,
					CustId:          cust,
					Request:         string(reqText),
					RequestType:     service.ServiceName,
					RequestStatus:   status.StatusDescription,
					RequestAmount:   req.Amount,
					RequestResponse: "",
					RequestDate:     time.Now(),
					DateCreated:     time.Now(),
					DateModified:    time.Now(),
				}
				if _, err := models.AddRequest(&v); err == nil {
					// Create a transaction record
					transaction := models.Bil_transactions{
						TransactionRefNumber: "TRX-" + strconv.FormatInt(time.Now().Unix(), 10) + strconv.FormatInt(v.RequestId, 10),
						Service:              service, // Assuming service ID is 1 for airtime
						Request:              &v,
						TransactionBy:        cust,
						Amount:               req.Amount,
						TransactingCurrency:  "GHC", // Assuming USD for simplicity
						SourceChannel:        sourceSystem,
						Source:               phoneNumber,
						Destination:          req.DestinationAccount,
						Charge:               0.0,    // Assuming no charge for simplicity
						Status:               status, // Assuming 1 means successful
						DateCreated:          time.Now(),
						DateModified:         time.Now(),
						CreatedBy:            1,
						ModifiedBy:           1,
						Active:               1, // Assuming active status
					}
					if _, err := models.AddBil_transactions(&transaction); err == nil {
						// Go to fulfillment
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
								ClientReference: transaction.TransactionRefNumber, // Use the request ID as the transaction ID
								CallbackUrl:     callbackurl,                      // Optional field for callback URL
								ExtraData:       selectedPackage,                  // Assuming this is the bundle key request
								ServiceId:       biller.BillerReferenceId,
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
								BilTransactionId:       &transaction,
								Amount:                 req.Amount,
								Biller:                 biller,
								SenderAccountNumber:    phoneNumber,
								RecipientAccountNumber: req.DestinationAccount,
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
								responseCode = false
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
							if thirdPartyResponse, err := thirdparty.ProcessGhanaWaterBillPayment(&c.Controller, tReq); err == nil {

								if thirdPartyResponse.ResponseCode == "0001" {
									// Transaction is pending
									// Update the transaction status to pending
									responseCode = true
									responseMessage = "Request is being processed"
									if status, err := models.GetStatus_codesByCode("PENDING"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "PENDING:: Failed to update transaction status"
										} else {
											responseCode = true
											responseMessage = "Request is being processed"
										}
									} else {
										logs.Error("Failed to get status for pending transaction: %v", err)
										responseCode = false
										responseMessage = "PENDING: Failed to get status for pending transaction"
									}
								} else if thirdPartyResponse.ResponseCode == "0000" {
									// Transaction is successful
									// Update the transaction status to successful
									responseCode = true
									responseMessage = "Request is successful"
									if status, err := models.GetStatus_codesByCode("SUCCESS"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "SUCCESS:: Failed to update transaction status"
										} else {
											// Prepare the response
											logs.Info("Transaction successful: ", transaction)
											responseCode = true
											responseMessage = "Transaction successful"
										}
									} else {
										logs.Error("Failed to get status for successful transaction: %v", err)
										responseCode = false
										responseMessage = "SUCCESS:: Failed to get status for successful transaction"
									}
								} else {
									// Transaction failed
									// Update the transaction status to failed
									responseCode = false
									responseMessage = "Transaction failed"
									if status, err := models.GetStatus_codesByCode("FAILED"); err == nil {
										transaction.Status = status
										if err := models.UpdateBil_transactionsById(&transaction); err != nil {
											logs.Error("Failed to update transaction status: %v", err)
											responseCode = false
											responseMessage = "FAILED:: Failed to update transaction status"
										}
									} else {
										logs.Error("Failed to get status for failed transaction: %v", err)
										responseCode = false
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
								v.RequestResponse = string(resText)
								v.DateModified = time.Now()
								if err := models.UpdateRequestById(&v); err != nil {
									logs.Error("Failed to update request response: %v", err)
									responseCode = true
									responseMessage = "Success response:: Failed to update request response"
								} else {
									logs.Info("Request response updated successfully")
								}

								c.Ctx.Output.SetStatus(200)
								// Prepare the response

								// Create the response object
								respData := responses.ThirdPartyBillPaymentDataApiResponse{
									Description:   "Payment for " + billerCode + " bill",
									Amount:        req.Amount,
									TransactionId: transaction.TransactionRefNumber,
								}
								response := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        &respData,
								}
								c.Data["json"] = response
							} else {
								logs.Error("Failed to process "+billerCode+" request: %v", err)
								responseCode = false
								responseMessage = "Failed to process " + billerCode + " request"
								resp := responses.ThirdPartyBillPaymentApiResponse{
									StatusCode:    responseCode,
									StatusMessage: responseMessage,
									Result:        nil,
								}
								c.Data["json"] = resp
							}
						} else {
							logs.Error("Failed to get biller by code: %v", err)
							responseCode = false
							responseMessage = "Failed to get biller by code"
							resp := responses.ThirdPartyBillPaymentApiResponse{
								StatusCode:    responseCode,
								StatusMessage: responseMessage,
								Result:        nil,
							}
							c.Data["json"] = resp
						}

					} else {
						logs.Error("Failed to create transaction record: %v", err)
						responseCode = false
						responseMessage = "Failed to create transaction record"
						resp := responses.ThirdPartyBillPaymentApiResponse{
							StatusCode:    responseCode,
							StatusMessage: responseMessage,
							Result:        nil,
						}
						c.Data["json"] = resp
					}
				} else {
					logs.Error("Failed to create request record: %v", err)
					responseCode = false
					responseMessage = "Failed to create request log"
					resp := responses.ThirdPartyBillPaymentApiResponse{
						StatusCode:    responseCode,
						StatusMessage: responseMessage,
						Result:        nil,
					}
					c.Data["json"] = resp
				}
			} else {
				logs.Error("Service not found: %v", err)
				responseCode = false
				responseMessage = "Failed to create transaction record. Service not found."
				resp := responses.ThirdPartyBillPaymentApiResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
			}
		} else {
			logs.Error("Customer not found: %v", err)
			responseCode = false
			responseMessage = "Failed to create transaction record"
			resp := responses.ThirdPartyBillPaymentApiResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
		}
	} else {
		logs.Error("Status not found: %v", err)
		responseCode = false
		responseMessage = "Failed to create transaction record"
		resp := responses.ThirdPartyBillPaymentApiResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
	}
	c.ServeJSON()
}
