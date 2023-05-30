package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
	"zkCache/consistenthash"
	"zkCache/pkg/response"
	"zkCache/pkg/valid"
	"zkCache/zklog"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	ServiceHost = "localhost"
	ServicePort = "9999"
	ServiceURL  = "http://" + ServiceHost + ":" + ServicePort + "/services"
)

type registry struct {
	registration map[ServiceName][]string // sericeName:[]string || 服务名:URLS
	mutex        *sync.RWMutex
	virtualNode  map[ServiceName]*consistenthash.Map
}

var selfReg = registry{
	registration: make(map[ServiceName][]string, 0),
	mutex:        new(sync.RWMutex),
	virtualNode:  make(map[ServiceName]*consistenthash.Map),
}

func RegisterHandlers(router *gin.Engine) {
	zklog.Logger.Info("Request received")
	// 获取服务
	router.GET("/services", getService)
	router.POST("/services/get", getService)
	// 注册服务
	router.POST("/services", addService)
	// 注销服务
	router.DELETE("/services", removeService)
}

// 服务注册
func addService(ctx *gin.Context) {
	var r RegistrationVO
	ctx.ShouldBind(&r)
	err := valid.Verification.Verify(r)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		response.ResponseMsg.FailResponse(ctx, err, nil)
		return
	}
	zklog.Logger.WithFields(logrus.Fields{
		"ServiceName": r.ServiceName,
		"ServiceURL":  r.ServiceURL,
	}).Info("Adding service:")

	err = selfReg.add(r)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		response.ResponseMsg.FailResponse(ctx, err, nil)
		return
	}
	response.ResponseMsg.SuccessResponse(ctx, nil)
}

// /服务注销
func removeService(ctx *gin.Context) {
	var r RegistrationVO
	ctx.ShouldBind(&r)
	err := valid.Verification.Verify(r)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		response.ResponseMsg.FailResponse(ctx, err, nil)
		return
	}
	url := r.ServiceURL
	zklog.Logger.Info("Remove service at URL:", url)
	err = selfReg.remove(r)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		response.ResponseMsg.FailResponse(ctx, err, nil)
		return
	}
	response.ResponseMsg.SuccessResponse(ctx, nil)
}

func urlsExistUrl(urls []string, serviceUrl string) bool {
	urlMap := make(map[string]struct{}, 0)
	for i := 0; i < len(urls); i++ {
		urlMap[urls[i]] = struct{}{}
	}
	_, exist := urlMap[serviceUrl]
	return exist
}
func (r *registry) add(reg RegistrationVO) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	serviceName := reg.ServiceName
	serviceUrl := reg.ServiceURL
	if _, ok := r.registration[serviceName]; !ok {
		r.registration[serviceName] = make([]string, 0)
	}

	if exist := urlsExistUrl(r.registration[serviceName], serviceUrl); !exist {
		r.registration[serviceName] = append(r.registration[serviceName], serviceUrl)
	}

	// 注册虚拟节点
	if _, ok := r.virtualNode[serviceName]; !ok {
		r.virtualNode[serviceName] = consistenthash.New(5, nil)
	}
	r.virtualNode[serviceName].Set(serviceUrl)
	go updateNodesMsg(serviceName)
	return nil
}
func (r *registry) remove(reg RegistrationVO) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	serviceName := reg.ServiceName
	serviceUrl := reg.ServiceURL
	if _, exist := r.registration[serviceName]; exist {
		for i := range r.registration[serviceName] {
			if r.registration[serviceName][i] == serviceUrl {
				r.registration[serviceName] = append(r.registration[serviceName][:i], r.registration[serviceName][i+1:]...)
				selfReg.virtualNode[serviceName].RemoveNodeByUrl(serviceUrl)
				go updateNodesMsg(serviceName)
				return nil
			}
		}
		return response.NewErrWithMsg(response.PARAMETER_ERROR,
			fmt.Sprintf("Found serviceName: %s ,not found URL: %s", serviceName, serviceUrl))
	}
	return response.NewErrWithMsg(response.PARAMETER_ERROR,
		fmt.Sprintf("Not found serviceName: %s ,not found URL: %s", serviceName, serviceUrl))
}

// 心跳检测
func Heartbeat(interval time.Duration) {
	for {
		checkReg := selfReg
		tempUrlsMap := make(map[ServiceName]map[string]int)
		for i := 0; i < 3; i++ {
			for serviceName, serviceURLs := range checkReg.registration {
				for _, url := range serviceURLs {
					resp, err := http.Get(url + "/healthy")
					if err != nil || resp.StatusCode != http.StatusOK {
						zklog.Logger.WithFields(logrus.Fields{
							"sericeName": serviceName,
							"serviceURL": url,
						}).Error("[心跳检测] 检测错误...")
						urlsMap, ok := tempUrlsMap[serviceName]
						if !ok {
							tempUrlsMap[serviceName] = make(map[string]int)
							urlsMap = tempUrlsMap[serviceName]
						}
						counts := urlsMap[url]
						urlsMap[url] = counts + 1
					}
					// else {
					// 	zklog.Logger.WithFields(logrus.Fields{
					// 		"sericeName": serviceName,
					// 		"serviceURL": url,
					// 	}).Info("[心跳检测] 检测通过...")
					// }
				}
			}
		}
		removeUrlsMap := make(map[ServiceName][]string)
		for serviceName, urlsMap := range tempUrlsMap {
			for url, counts := range urlsMap {
				if counts == 3 {
					_, exist := removeUrlsMap[serviceName]
					if !exist {
						removeUrlsMap[serviceName] = make([]string, 0)
					}
					removeUrlsMap[serviceName] = append(removeUrlsMap[serviceName], url)
				}
			}
		}
		//移除心跳检测失败的
		go removeUrls(removeUrlsMap)
		time.Sleep(interval)
	}
}
func removeUrls(removeUrlsMap map[ServiceName][]string) {
	for serviceName, serviceUrls := range removeUrlsMap {
		for _, url := range serviceUrls {
			selfReg.remove(RegistrationVO{
				ServiceName: serviceName,
				ServiceURL:  url,
			})
			selfReg.virtualNode[serviceName].RemoveNodeByUrl(url)
		}
		go updateNodesMsg(serviceName)
	}

}

func updateNodesMsg(serviceName ServiceName) {
	urls := selfReg.virtualNode[serviceName].GetUrlsSortByKey()
	zklog.Logger.WithField("urls", urls).Debug()
	for _, url := range urls {
		data := make(map[string][]string)
		data["urls"] = urls
		jsonData, err := json.Marshal(data)
		if err != nil {
			zklog.Logger.WithField("err", err).Error()
		}

		method := "GET"
		payload := strings.NewReader(string(jsonData))
		client := &http.Client{}
		req, err := http.NewRequest(method, url+"/updateNodePool", payload)

		if err != nil {
			zklog.Logger.WithField("err", err).Error()
			return
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			zklog.Logger.WithField("err", err).Error()
			return
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			zklog.Logger.WithField("err", err).Error()
			return
		}
		zklog.Logger.WithField("send update msg response data", string(body)).Debug()
	}

}

// 注册中心拉取服务 | 环形hash
func getService(ctx *gin.Context) {
	var r GetServiceVO
	ctx.ShouldBind(&r)
	err := valid.Verification.Verify(r)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		response.ResponseMsg.FailResponse(ctx, err, nil)
		return
	}
	selfReg.mutex.RLock()
	defer selfReg.mutex.RUnlock()
	if len(selfReg.registration[r.ServiceName]) == 0 {
		response.ResponseMsg.FailResponse(ctx, response.NewErr(response.ERROR), nil)
		return
	}
	// rand.Seed(time.Now().UnixNano())
	// index := rand.Intn(len(selfReg.registration[r.ServiceName]))
	// url := selfReg.registration[r.ServiceName][index]

	// 根据key获取url
	url := selfReg.virtualNode[r.ServiceName].Get(r.Key)

	zklog.Logger.WithFields(logrus.Fields{
		"Selected Instance:": url,
		// "index":              index,
		"counts": len(selfReg.registration[r.ServiceName]),
	}).Info("Selected Instance:", url)
	if err != nil {
		zklog.Logger.WithField("err", err).Error()
		response.ResponseMsg.FailResponse(ctx, err, nil)
		return
	}
	response.ResponseMsg.SuccessResponse(ctx, GetServiceDTO{
		Url: url,
	})
}
