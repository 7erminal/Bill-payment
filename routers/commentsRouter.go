package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["billpayment_service/controllers:CallbackController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:CallbackController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/process`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:CallbackController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:CallbackController"],
        beego.ControllerComments{
            Method: "TransactionStatusCheck",
            Router: `/transaction-status-check`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:objectId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:objectId`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:ObjectController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:objectId`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "AccountQuery",
            Router: `/account-query/:billercode/:accountNumber`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "DSTVAccountQuery",
            Router: `/dstv-account-query/:accountNumber`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "ECGAccountQuery",
            Router: `/ecg-account-query/:accountNumber`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "GhanaWaterAccountQuery",
            Router: `/ghana-water-account-query/:accountNumber`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "GoTVAccountQuery",
            Router: `/gotv-account-query/:accountNumber`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "PayDSTVBill",
            Router: `/pay-dstv-bill`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "PayECGBill",
            Router: `/pay-ecg-bill`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "PayGoTVBill",
            Router: `/pay-gotv-bill`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "PayStartimesBill",
            Router: `/pay-startimes-bill`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "PayWaterBill",
            Router: `/pay-water-bill`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:RequestController"],
        beego.ControllerComments{
            Method: "StartimesAccountQuery",
            Router: `/startimes-account-query/:accountNumber`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:uid`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:uid`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:uid`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "Login",
            Router: `/login`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["billpayment_service/controllers:UserController"] = append(beego.GlobalControllerRouter["billpayment_service/controllers:UserController"],
        beego.ControllerComments{
            Method: "Logout",
            Router: `/logout`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
