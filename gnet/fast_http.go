package gnet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/seveye/goms/rpc"
	"github.com/seveye/goms/util"
	"github.com/valyala/fasthttp"
)

func getRemote2(ctx *routing.Context) string {
	ip := string(ctx.Request.Header.Peek("x-forwarded-for"))
	if ip != "" && ip != "unknown" {
		ip = strings.Split(ip, ",")[0]
	}
	if ip == "" {
		ip = strings.Split(ctx.RemoteIP().String(), ":")[0]
	}
	return ip
}

func RunApiHttp2(port int, dev bool) error {
	router := routing.New()

	router.Options("/*", func(ctx *routing.Context) error {
		return nil
	})
	router.Post("/*", ProtectedHandler2)
	if dev {
		router.Get("/*", ProtectedHandler2)
	}
	util.Info("启动api2服务器", "port", port)

	return fasthttp.ListenAndServe(fmt.Sprintf(":%v", port), router.HandleRequest)
}

// ProtectedHandler2 入口
func ProtectedHandler2(ctx *routing.Context) error {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")

	urlStr := string(ctx.Path())
	remote := getRemote2(ctx)
	arr := strings.Split(urlStr, "/")
	if len(arr) != 4 || arr[1] != "api" {
		JsonResponse2(ctx, map[string]interface{}{
			"code":    1,
			"codeMsg": "请求地址格式错误",
		})
		// util.Info("登陆请求地址格式错误", "url", urlStr)
		return nil
	}

	reqBuff := ctx.PostBody()
	if len(reqBuff) == 0 {
		reqBuff = []byte("{}")
	}

	context := &rpc.Context{Remote: remote}

	// begin := time.Now()
	serviceMethon := util.Upper(arr[2], 1) + "." + util.Upper(arr[3], 1)
	var (
		rspBuff []byte
		err     error
	)
	util.ProtectCall(func() {
		_, rspBuff, err = rpc.JsonCall(context, serviceMethon, reqBuff)
	}, func() {
		err = fmt.Errorf("server internal error")
	})
	// cost := time.Since(begin).String()
	// util.Debug("请求信息", "api", serviceMethon, "remote", remote, "url", urlStr, "reqBuff", string(reqBuff), "rspBuff", string(rspBuff), "err", err, "cost", cost)
	if err != nil {
		JsonResponse2(ctx, map[string]interface{}{
			"code":    1,
			"codeMsg": err.Error(),
		})
		util.Info("JsonCall失败", "api", serviceMethon, "remote", remote, "url", urlStr, "err", err)
		return nil
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Write(rspBuff)
	return nil
}

func JsonResponse2(ctx *routing.Context, response interface{}) {
	json, err := json.Marshal(response)
	if err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Write(json)
}
