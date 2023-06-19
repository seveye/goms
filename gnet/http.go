package gnet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"gitee.com/jkkkls/goms/rpc"
	"gitee.com/jkkkls/goms/util"
	"github.com/codegangsta/negroni"
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Data string `json:"data"`
}

type Token struct {
	Token string `json:"token"`
}

var (
	SecretKey = "www.winuim.com2020"
)

func RunApiServer(port int) error {
	ser := http.NewServeMux()
	ser.Handle("/", negroni.New(
		negroni.HandlerFunc(ValidateTokenMiddleware),
		negroni.Wrap(http.HandlerFunc(ProtectedHandler)),
	))

	s := http.Server{
		Addr:        fmt.Sprintf(":%v", port),
		Handler:     ser,
		IdleTimeout: 10 * time.Second,
	}

	util.Info("启动api服务器", "port", port)
	// s.SetKeepAlivesEnabled(false)
	return s.ListenAndServe()
	// return http.ListenAndServe(fmt.Sprintf(":%v", port), ser)
}

func RunApiHttpsServer(port int, crtFile, keyFile string) error {
	ser := http.NewServeMux()
	ser.Handle("/", negroni.New(
		negroni.HandlerFunc(ValidateTokenMiddleware),
		negroni.Wrap(http.HandlerFunc(ProtectedHandler)),
	))

	s := http.Server{
		Addr:        fmt.Sprintf(":%v", port),
		Handler:     ser,
		IdleTimeout: 10 * time.Second,
	}

	util.Info("启动api https服务器", "port", port)
	// s.SetKeepAlivesEnabled(false)
	return s.ListenAndServeTLS(crtFile, keyFile)
	// return http.ListenAndServe(fmt.Sprintf(":%v", port), ser)
}

func getRemote(r *http.Request) string {
	ip := r.Header.Get("x-forwarded-for")
	if ip != "" && ip != "unknown" {
		ip = strings.Split(ip, ",")[0]
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

// ProtectedHandler 入口
func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")

	if r.Method != "POST" {
		JsonResponse(w, map[string]interface{}{
			"code":    1,
			"codeMsg": "只支持POST方法",
		})
		return
	}

	urlStr := r.URL.Path
	remote := getRemote(r)
	arr := strings.Split(urlStr, "/")
	if len(arr) != 4 || arr[1] != "api" {
		JsonResponse(w, map[string]interface{}{
			"code":    1,
			"codeMsg": "请求地址格式错误",
		})
		util.Info("登陆请求地址格式错误", "url", urlStr)
		return
	}

	defer r.Body.Close()
	reqBuff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JsonResponse(w, map[string]interface{}{
			"code":    1,
			"codeMsg": "读取数据失败",
		})
		return
	}
	if len(reqBuff) == 0 {
		reqBuff = []byte("{}")
	}

	context := &rpc.Context{Remote: remote, Kvs: make(map[string]string)}
	m := r.URL.Query()
	for k, v := range m {
		context.Kvs[k] = v[0]
	}

	begin := time.Now()
	serviceMethon := util.Upper(arr[2], 1) + "." + util.Upper(arr[3], 1)
	var (
		rspBuff []byte
	)
	util.ProtectCall(func() {
		_, rspBuff, err = rpc.JsonCall(context, serviceMethon, reqBuff)
	}, func() {
		err = fmt.Errorf("server internal error")
	})
	cost := time.Since(begin).String()
	util.Debug("请求信息", "api", serviceMethon, "remote", remote, "url", urlStr, "reqBuff", string(reqBuff), "rspBuff", string(rspBuff), "err", err, "cost", cost)
	if err != nil {
		JsonResponse(w, map[string]interface{}{
			"code":    1,
			"codeMsg": err.Error(),
		})
		util.Info("JsonCall失败", "err", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(rspBuff)
}

func ValidateTokenMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	next(w, r)
}

func JsonResponse(w http.ResponseWriter, response interface{}) {
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
