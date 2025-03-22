package transformer

import (
	"bytes"
	"fmt"
	"time"

	"github.com/k8stech/alertmanager-wechatrobot-webhook/model"
)

// 新增一个函数来获取告警颜色
func getAlertColor(severity string) string {
	switch severity {
	case "critical":
		return "warning"
	case "firing":
		return "warning"
	case "resolved":
		return "green"
	default:
		return "comment"
	}
}

// TransformToMarkdown transform alertmanager notification to wechat markdow message
func TransformToMarkdown(notification model.Notification, grafanaURL string, alertDomain string) (markdown *model.WeChatMarkdown, robotURL string, err error) {
	status := notification.Status
	annotations := notification.CommonAnnotations
	robotURL = annotations["wechatRobot"]
	var buffer bytes.Buffer

	for _, alert := range notification.Alerts {
		labels := alert.Labels
		// 加载 CST 时区
		cstZone, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			// 处理错误，例如无法加载时区
			fmt.Println("Error loading location:", err)
		}

		// 将 UTC 时间转换为 CST 时间
		cstTime := alert.StartsAt.In(cstZone)
		// instance := labels["instance"]
		hostName := labels["node_name"]
		ip := labels["ip"]
		// // 获取告警等级
		// severity := labels["severity"]
		// 获取对应的颜色
		alertColor := getAlertColor(status)
		if status == "resolved" {
			buffer.WriteString(fmt.Sprintf("### 【监控告警】（已解决）主机: <font color='%s'> %s </font>\n", alertColor, hostName))
		} else {
			buffer.WriteString(fmt.Sprintf("### 【监控告警】主机: <font color='%s'> %s </font>\n", alertColor, hostName))
		}
		buffer.WriteString(fmt.Sprintf(">主机名称：**<font color=\"comment\">%s</font>**\n", hostName))
		buffer.WriteString(fmt.Sprintf(">IP地址   ：<font color=\"%s\">%s</font>\n", alertColor, ip))
		buffer.WriteString(fmt.Sprintf(">告警时间：<font color=\"comment\">%s</font>\n", cstTime.Format("2006-01-02 15:04:05")))
		if status == "resolved" {
			d := time.Since(alert.StartsAt)
			hours := int(d.Hours())
			minutes := int(d.Minutes()) % 60
			seconds := int(d.Seconds()) % 60
			elapsed := ""
			if hours > 0 {
				elapsed = fmt.Sprintf("%dh %dm", hours, minutes)
			} else if minutes > 0 {
				elapsed = fmt.Sprintf("%dm %ds", minutes, seconds)
			} else {
				elapsed = fmt.Sprintf("%ds", seconds)
			}
			buffer.WriteString(fmt.Sprintf(">持续时长：<font color=\"comment\">%s</font>\n", elapsed))
		}
		buffer.WriteString(fmt.Sprintf(">告警名称：<font color=\"%s\">%s</font>\n", alertColor, alert.Annotations["summary"]))
		buffer.WriteString(fmt.Sprintf(">详细信息：<font color=\"comment\">%s</font>\n", alert.Annotations["description"]))
		buffer.WriteString(fmt.Sprintf(">当前状态：<font color=\"comment\">%s</font>\n\n", status))
	}

	markdown = &model.WeChatMarkdown{
		MsgType: "markdown",
		Markdown: &model.Markdown{
			Content: buffer.String(),
		},
	}

	return
}
