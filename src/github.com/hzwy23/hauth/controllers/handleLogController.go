package controllers

import (

	"fmt"

	"text/template"

	"github.com/astaxie/beego/context"

	"github.com/hzwy23/dbobj"
	"github.com/hzwy23/hauth/utils"
	"github.com/hzwy23/hauth/utils/hret"
	"github.com/hzwy23/hauth/utils/logs"
	"github.com/hzwy23/hauth/utils/token/hjwt"
	"github.com/tealeg/xlsx"
	"github.com/hzwy23/hauth/models"
)

type HandleLogsController struct {
}

type handleLogs struct {
	Uuid        string `json:"uuid"`
	User_id     string `json:"user_id"`
	Handle_time string `json:"handle_time"`
	Client_ip   string `json:"client_ip"`
	Status_code string `json:"status_code"`
	Method      string `json:"method"`
	Url         string `json:"url"`
	Data        string `json:"data"`
}

var HandleLogsCtl = new(HandleLogsController)

func (HandleLogsController) GetHandleLogPage(ctx *context.Context) {
	ctx.Request.ParseForm()
	if !models.BasicAuth(ctx){
		hret.WriteHttpErrMsgs(ctx.ResponseWriter,403,"权限不足")
		return
	}
	hz, _ := template.ParseFiles("./views/hauth/handle_logs_page.tpl")
	hz.Execute(ctx.ResponseWriter, nil)
}

func (HandleLogsController) Download(ctx *context.Context) {
	ctx.Request.ParseForm()

	if !models.BasicAuth(ctx){
		hret.WriteHttpErrMsgs(ctx.ResponseWriter,403,"权限不足")
		return
	}

	ctx.ResponseWriter.Header().Set("Content-Type", "application/vnd.ms-excel")

	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "No Auth")
		return
	}
	sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? order by handle_time desc`
	var rst []handleLogs
	rows, err := dbobj.Query(sql, jclaim.Domain_id)
	defer rows.Close()
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
		return
	}
	err = dbobj.Scan(rows, &rst)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
		return
	}

	var file *xlsx.File
	var sheet *xlsx.Sheet

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("机构信息")
	if err != nil {
		fmt.Printf(err.Error())
	}

	row := sheet.AddRow()
	cell1 := row.AddCell()
	cell1.Value= "用户"
	cell2 := row.AddCell()
	cell2.Value= "操作日期"
	cell3 := row.AddCell()
	cell3.Value= "客户端IP"
	cell4 := row.AddCell()
	cell4.Value= "请求方法"
	cell5 := row.AddCell()
	cell5.Value = "API地址"
	cell6 := row.AddCell()
	cell6.Value = "返回状态"
	cell7 := row.AddCell()
	cell7.Value = "请求数据"


	for _,v:=range rst{
		row := sheet.AddRow()
		cell1 := row.AddCell()
		cell1.Value=v.User_id
		cell2 := row.AddCell()
		cell2.Value=v.Handle_time
		cell3 := row.AddCell()
		cell3.Value=v.Client_ip
		cell4 := row.AddCell()
		cell4.Value=v.Method
		cell5 := row.AddCell()
		cell5.Value = v.Url
		cell6 := row.AddCell()
		cell6.Value = v.Status_code
		cell7 := row.AddCell()
		cell7.Value = v.Data
	}
	file.Write(ctx.ResponseWriter)
}

func (HandleLogsController) GetHandleLogs(ctx *context.Context) {
	ctx.Request.ParseForm()

	if !models.BasicAuth(ctx){
		hret.WriteHttpErrMsgs(ctx.ResponseWriter,403,"权限不足")
		return
	}

	offset := ctx.Request.FormValue("offset")
	limit := ctx.Request.FormValue("limit")
	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "No Auth")
		return
	}
	sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? order by handle_time desc limit ?,?`
	var rst []handleLogs
	rows, err := dbobj.Query(sql, jclaim.Domain_id, offset, limit)
	defer rows.Close()
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
		return
	}
	err = dbobj.Scan(rows, &rst)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
		return
	}
	cntsql := `select count(*) from sys_handle_logs t
			where t.domain_id = ?`
	hret.WriteBootstrapTableJson(ctx.ResponseWriter, dbobj.Count(cntsql, jclaim.Domain_id), rst)
}

func (HandleLogsController) SerachLogs(ctx *context.Context) {
	ctx.Request.ParseForm()

	if !models.BasicAuth(ctx){
		hret.WriteHttpErrMsgs(ctx.ResponseWriter,403,"权限不足")
		return
	}

	userid := ctx.Request.FormValue("UserId")
	start := ctx.Request.FormValue("StartDate")
	end := ctx.Request.FormValue("EndDate")

	cookie, _ := ctx.Request.Cookie("Authorization")
	jclaim, err := hjwt.ParseJwt(cookie.Value)
	if err != nil {
		logs.Error(err)
		hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "No Auth")
		return
	}
	var rst []handleLogs
	if userid != "" && utils.ValidDate(start) && utils.ValidDate(end) {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and user_id = ? and handle_time >= str_to_date(?,'%Y-%m-%d')
			and handle_time < str_to_date(?,'%Y-%m-%d')
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, userid, start, end)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else if userid != "" && utils.ValidDate(start) {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and user_id = ? and handle_time >= str_to_date(?,'%Y-%m-%d')
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, userid, start)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else if userid != "" && utils.ValidDate(end) {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and user_id = ? and handle_time >= str_to_date(?,'%Y-%m-%d')
			and handle_time < str_to_date(?,'%Y-%m-%d')
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, userid, start, end)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else if utils.ValidDate(start) && utils.ValidDate(end) {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and handle_time >= str_to_date(?,'%Y-%m-%d')
			and handle_time < str_to_date(?,'%Y-%m-%d')
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, start, end)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else if utils.ValidDate(start) {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and handle_time >= str_to_date(?,'%Y-%m-%d')
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, start, end)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else if utils.ValidDate(end) {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and handle_time < str_to_date(?,'%Y-%m-%d')
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, start, end)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else if userid != "" {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? and user_id = ?
			order by handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id, userid)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	} else {
		sql := `select uuid,user_id,handle_time,client_ip,status_code,method,url,data from sys_handle_logs t
			where t.domain_id = ? order by user_id,handle_time desc`

		rows, err := dbobj.Query(sql, jclaim.Domain_id)
		defer rows.Close()
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
		err = dbobj.Scan(rows, &rst)
		if err != nil {
			logs.Error(err)
			hret.WriteHttpErrMsgs(ctx.ResponseWriter, 310, "query failed.")
			return
		}
	}

	hret.WriteJson(ctx.ResponseWriter, rst)
}
