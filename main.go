package main

import (
	"flag"
	"net/http"

	"github.com/k8stech/alertmanager-wechatrobot-webhook/model"
	"github.com/k8stech/alertmanager-wechatrobot-webhook/notifier"

	"github.com/gin-gonic/gin"
)

var (
	h           bool
	RobotKey    string
	addr        string
	grafanaUrl  string
	alertDomain string
)

func init() {
	flag.BoolVar(&h, "h", false, "help")
	flag.StringVar(&RobotKey, "RobotKey", "", "global wechatrobot webhook, you can overwrite by alert rule with annotations wechatRobot")
	flag.StringVar(&addr, "addr", ":8999", "listen addr")
	flag.StringVar(&grafanaUrl, "grafanaUrl", "grafana.vnnox.com/d/PwMJtdvnr/k8s-chu-neng-cnanduat", "grafanaUrl url")
	flag.StringVar(&alertDomain, "alertDomain", "emscn-prometheus.ampaura.tech", "alertDomain url")
}

func main() {

	flag.Parse()

	if h {
		flag.Usage()
		return
	}

	router := gin.Default()
	router.POST("/webhook", func(c *gin.Context) {
		var notification model.Notification
		err := c.BindJSON(&notification)
		//bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		//if err != nil {
		//	log.Printf("Error reading request body: %v", err)
		//	c.AbortWithStatus(http.StatusBadRequest)
		//	return
		//}
		//// 重新设置请求体以确保后续处理可以正常进行
		//c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		//
		//// 打印请求体内容
		//log.Printf("Request Body: %s", bodyBytes)
		////gin.LogFormatter(c.Request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		RobotKey := c.DefaultQuery("key", RobotKey)

		err = notifier.Send(notification, RobotKey, grafanaUrl, alertDomain)
		//fmt.Println("notification:", notification, "RobotKey:", RobotKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		}

		c.JSON(http.StatusOK, gin.H{"message": "send to wechatbot successful!"})

	})
	router.Run(addr)
}
