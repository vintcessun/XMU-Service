package api

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/vintcessun/XMU-Service/utils"
)

var qrRegex = regexp.MustCompile(`<input[^>]*?name="execution"[^>]*?value="([^"]*)"[^>]*?>`)
var lntURL = "https://lnt.xmu.edu.cn/"
var retryLoginTimes = 3
var lntURLParsed *url.URL

func init() {
	url, err := url.Parse(lntURL)
	if err != nil {
		panic("login初始化登录地址失败")
	}
	lntURLParsed = url
}

type LntServiceQr struct {
	client    *resty.Client
	Session   string
	QrcodeId  string
	execution string
	service   string
}

func (l *LntServiceQr) GetInfo() error {
	ua := utils.GetFakeUAComputer()
	l.client = resty.New()
	l.client.SetHeader("User-Agent", ua)
	l.client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(100))

	var retErr error

	for range retryLoginTimes {
		time.Sleep(1 * time.Second)
		resp, err := l.client.R().Get(lntURL)
		if err != nil {
			retErr = err
			continue
		}
		endUrl := resp.RawResponse.Request.URL.String()
		serviceIndex := strings.Index(endUrl, "service=")
		if serviceIndex == -1 {
			continue
		}
		l.service = endUrl[(serviceIndex + 8):]
		html := resp.String()

		index := strings.Index(html, "qrLoginForm")
		if index == -1 {
			retErr = err
			continue
		}

		qrHtml := html[index:]

		executions := qrRegex.FindStringSubmatch(qrHtml)

		if len(executions) <= 1 {
			continue
		}

		l.execution = executions[1]

		resp, err = l.client.R().Get(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/qrCode/getToken?ts=%d", time.Now().UnixMilli()))
		if err != nil {
			retErr = err
			continue
		}
		l.QrcodeId = resp.String()

	}
	return retErr
}

func (l *LntServiceQr) GetQrState() (string, error) {
	resp, err := l.client.R().Get(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/qrCode/getStatus.htl?ts=%d&uuid=%s", time.Now().UnixMilli(), l.QrcodeId))
	if err != nil {
		return "", err
	}

	state := resp.String()

	return state, nil
}

func (l *LntServiceQr) Finish() error {
	data := map[string]string{
		"lt":        "",
		"uuid":      l.QrcodeId,
		"cllt":      "qrLogin",
		"dllt":      "generalLogin",
		"execution": l.execution,
		"_eventId":  "submit",
		"rmShown":   "1",
	}

	var retErr error

	for range retryLoginTimes {
		_, retErr := l.client.R().SetFormData(data).Post(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/login?display=qrLogin&service=%s", l.service))
		if retErr != nil {
			continue
		}

		sessionCookie, ok := utils.GetSessionCookie(l.client.GetClient().Jar.Cookies(lntURLParsed), "session")
		if !ok {
			continue
		}
		l.Session = sessionCookie.Value

		if !ok {
			continue
		}
		break
	}

	return retErr
}
