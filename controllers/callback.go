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

// CallbackController operations for Callback
type CallbackController struct {
	beego.Controller
}

// URLMapping ...
func (c *CallbackController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("TransactionStatusCheck", c.TransactionStatusCheck)
}

// Post ...
// @Title Create
// @Description create Callback
// @Param	body		body 	requests.CallbackRequest	true		"body for Callback content"
// @Success 201 {object} responses.CallbackResponse
// @Failure 403 body is empty
// @router /process [post]
func (c *CallbackController) Post() {
	var v requests.CallbackRequest
	logs.Info("Received callback request: ", string(c.Ctx.Input.RequestBody))
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {
		c.Data["json"] = map[string]string{"error": "Invalid request body"}
		c.Ctx.Output.SetStatus(400)
		c.ServeJSON()
		return
	}

	responseCode := false
	responseMessage := "Invalid request"

	// Handle successful callback
	transactionId := ""
	if v.Data.ClientReference != nil {
		logs.Info("Transaction ID found in request: ", *v.Data.ClientReference)
		transactionId = *v.Data.ClientReference
	}
	logs.Info("About to get transaction by ID: ", transactionId)
	if resp, err := models.GetBil_transactionsByTransactionRefNum(transactionId); err == nil {
		logs.Info("Request ID: ", resp.Request.RequestId)
		if resp != nil {
			// Update the transaction status
			statusCode := "SUCCESS"
			if v.ResponseCode == "0000" {
				statusCode = "SUCCESS"
			} else {
				// Handle error in callback
				statusCode = "FAILED"

			}

			status, err := models.GetStatus_codesByCode(statusCode)
			if err == nil {
				resp.Status = status
				resp.DateModified = time.Now()
				if v.Data.ExternalTransactionId != nil {
					resp.ExternalReferenceNumber = *v.Data.ExternalTransactionId
				}
				if v.Data.Meta != nil {
					commission, err := strconv.ParseFloat(v.Data.Meta.Commission, 64)
					if err != nil {
						c.Data["json"] = map[string]string{"error": "Invalid commission value"}
						c.Ctx.Output.SetStatus(400)
						c.ServeJSON()
						return
					}
					resp.Commission = commission
				}
			} else {
				c.Data["json"] = map[string]string{"error": "Status code not found"}
				c.Ctx.Output.SetStatus(404)
			}

			if err := models.UpdateBil_transactionsById(resp); err != nil {
				logs.Info("Failed to update transaction status: %v", err)
				responseCode = false
				responseMessage = "Failed to update transaction status"
				resp := responses.CallbackResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        "FAILED",
				}
				c.Data["json"] = resp
				c.Ctx.Output.SetStatus(200)
			} else {
				// c.Data["json"] = map[string]string{"message": "Transaction updated successfully"}

				// Update request with callback data
				resText, err := json.Marshal(v)
				if err != nil {
					logs.Error("Failed to marshal callback request: %v", err)
					// c.Data["json"] = "Invalid request format"
					// c.ServeJSON()
					// return
				}

				logs.Info("Callback response text: %s", string(resText))
				logs.Info("Updating request", resp.Request.RequestId, " with callback response")
				if request, err := models.GetRequestById(resp.Request.RequestId); err == nil {
					logs.Info("Found request: ", request.RequestId)
					request.CallbackResponse = string(resText)

					request.DateModified = time.Now()
					if err := models.UpdateRequestById(request); err != nil {
						logs.Error("Failed to update request: %v", err)
						// c.Data["json"] = "Failed to update request"
						// c.ServeJSON()
						// return
					} else {
						logs.Info("Request updated successfully with callback response")
					}
				} else {
					logs.Error("Failed to retrieve request by ID: %v", err)
				}

				responseCode = true
				responseMessage = "Transaction updated successfully"
				resp := responses.CallbackResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        "SUCCESS",
				}
				c.Data["json"] = resp
				c.Ctx.Output.SetStatus(200)
			}
		} else {
			logs.Info("Transaction not found for ID: %s", transactionId)
			responseCode = false
			responseMessage = "Transaction not found"
			resp := responses.CallbackResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        "FAILED",
			}
			c.Data["json"] = resp
			// c.Data["json"] = map[string]string{"error": "Transaction not found"}
			c.Ctx.Output.SetStatus(200)
		}
	} else {
		c.Data["json"] = map[string]string{"error": "Failed to retrieve transaction"}
		logs.Info("Failed to retrieve transaction: %s", err.Error())
		responseCode = false
		responseMessage = "Failed to retrieve transaction"
		resp := responses.CallbackResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        "FAILED",
		}
		c.Data["json"] = resp
		// c.Data["json"] = map[string]string{"error": "Transaction not found"}
		c.Ctx.Output.SetStatus(200)
	}

	c.ServeJSON()
}

// StatusCheck ...
// @Title Status Check
// @Description Check the status of a transaction
// @Param	body		body 	requests.TransactionStatusRequest	true		"body for Callback content"
// @Success 201 {object} responses.CallbackResponse
// @Failure 403 body is empty
// @router /transaction-status-check [post]
func (c *CallbackController) TransactionStatusCheck() {
	var v requests.TransactionStatusRequest

	json.Unmarshal(c.Ctx.Input.RequestBody, &v)

	responseCode := false
	responseMessage := "Status check failed"

	logs.Info("Received transaction status check request: ", string(c.Ctx.Input.RequestBody))
	if transaction, err := models.GetBil_transactionsByTransactionRefNum(v.TransactionID); err == nil {
		if transaction != nil {
			responseCode = true
			responseMessage = "Transaction found"

			req := requests.TransactionStatusThirdPartyRequest{
				TransactionID:           transaction.TransactionRefNumber,
				ThirdParthTransactionID: "",
				NetworkTransactionID:    "",
			}

			response, err := thirdparty.GetTransactionStatus(&c.Controller, req)
			if err != nil {
				logs.Error("Failed to get transaction status: %v", err)
				// c.Data["json"] = map[string]string{"error": "Failed to get transaction status"}
				// c.Ctx.Output.SetStatus(500)
				responseCode = false
				responseMessage = "Failed to get transaction status"
				resp := responses.TransactionStatusResponse{
					StatusCode:    responseCode,
					StatusMessage: responseMessage,
					Result:        nil,
				}
				c.Data["json"] = resp
				return
			}
			resp := responses.TransactionStatusResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        response.Data,
			}
			c.Data["json"] = resp
			c.Ctx.Output.SetStatus(200)
		} else {
			logs.Info("Transaction not found for ID: %s", v.TransactionID)
			responseCode := false
			responseMessage := "Transaction not found"
			resp := responses.TransactionStatusResponse{
				StatusCode:    responseCode,
				StatusMessage: responseMessage,
				Result:        nil,
			}
			c.Data["json"] = resp
			c.Ctx.Output.SetStatus(404)
		}
	} else {
		logs.Error("Failed to retrieve transaction by ID: %v", err)
		// c.Data["json"] = map[string]string{"error": "Failed to retrieve transaction"}
		responseCode = false
		responseMessage = "Failed to retrieve transaction"
		resp := responses.TransactionStatusResponse{
			StatusCode:    responseCode,
			StatusMessage: responseMessage,
			Result:        nil,
		}
		c.Data["json"] = resp
		c.Ctx.Output.SetStatus(500)
	}
}
