package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/forgoer/openssl"
	"github.com/go-resty/resty/v2"
	"github.com/vintcessun/XMU-Service/utils"
)

var executionRegex = regexp.MustCompile(`<input[^>]*?name="execution"[^>]*?value="([^"]*)"[^>]*?>`)
var saltRegex = regexp.MustCompile(`<input[^>]*?id="pwdEncryptSalt"[^>]*?value="([^"]*)"[^>]*?>`)
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
	client    *utils.ServiceClient
	Session   string
	QrcodeId  string
	execution string
}

func getClient(loginType string) (*utils.ServiceClient, error) {
	ua := utils.GetFakeUAComputer()
	client := resty.New()
	client.SetHeader("User-Agent", ua)
	client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(100))

	client.R().Get(lntURL)

	var retErr error

	for range retryLoginTimes {
		time.Sleep(1 * time.Second)
		resp, err := client.R().Get(lntURL)
		if err != nil {
			retErr = err
			continue
		}
		endUrl := resp.RawResponse.Request.URL.String()
		serviceIndex := strings.Index(endUrl, "service=")
		if serviceIndex == -1 {
			continue
		}
		service := endUrl[(serviceIndex + 8):]

		return utils.NewClient(client, fmt.Sprintf("https://ids.xmu.edu.cn/authserver/login?type=%s&service=%s", loginType, service))
	}

	return nil, retErr
}

func (l *LntServiceQr) GetInfo() error {
	client, err := getClient("qrLogin")
	if err != nil {
		return err
	}
	l.client = client

	var retErr error
	for range retryLoginTimes {
		index := strings.Index(l.client.Html, "qrLoginForm")
		if index == -1 {
			retErr = err
			continue
		}

		qrHtml := l.client.Html[index:]

		executions := executionRegex.FindStringSubmatch(qrHtml)

		if len(executions) <= 1 {
			continue
		}

		l.execution = executions[1]

		resp, err := l.client.Client.R().Get(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/qrCode/getToken?ts=%d", time.Now().UnixMilli()))
		if err != nil {
			retErr = err
			continue
		}
		l.QrcodeId = resp.String()
	}
	return retErr
}

func (l *LntServiceQr) GetQrState() (string, error) {
	resp, err := l.client.Client.R().Get(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/qrCode/getStatus.htl?ts=%d&uuid=%s", time.Now().UnixMilli(), l.QrcodeId))
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

	for range retryLoginTimes {
		_, retErr := l.client.Client.R().SetFormData(data).Post(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/login?display=qrLogin&service=%s", l.client.Service))
		if retErr != nil {
			continue
		}

		sessionCookie, ok := utils.GetSessionCookie(l.client.Client.GetClient().Jar.Cookies(lntURLParsed), "session")
		if !ok {
			continue
		}
		l.Session = sessionCookie.Value

		if !ok {
			continue
		}
		return nil
	}

	return errors.New("LoginFailed")
}

type LntServicePassword struct {
	client            *utils.ServiceClient
	Session           string
	Username          string
	Password          string
	execution         string
	salt              string
	randomPassword    string
	iv                string
	encryptedPassword string
}

type ApiLoginChechIfNeedCaptcha struct {
	IsNeed bool `json:"isNeed"`
}

const aesChars = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"

func randomString(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	result := make([]byte, length)
	for i := range result {
		idx := r.Intn(len(aesChars))
		result[i] = aesChars[idx]
	}

	return string(result)
}

func (l *LntServicePassword) Login() error {
	client, err := getClient("userNameLogin")
	if err != nil {
		return err
	}
	l.client = client

	for range retryLoginTimes {
		time.Sleep(1 * time.Second)

		resp, err := l.client.Client.R().Get(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/checkNeedCaptcha.htl?username=%s&_=%d", l.Username, time.Now().UnixMilli()))
		if err != nil {
			continue
		}
		data, err := utils.UnmarshalJSON[ApiLoginChechIfNeedCaptcha](resp.Body())
		if data.IsNeed {
			return fmt.Errorf("LoginNeedCaptcha")
		}

		index := strings.Index(l.client.Html, "pwdLoginDiv")
		if index == -1 {
			continue
		}

		pwdHtml := l.client.Html[index:]

		executions := executionRegex.FindStringSubmatch(pwdHtml)

		if len(executions) <= 1 {
			continue
		}

		l.execution = executions[1]

		salts := saltRegex.FindStringSubmatch(pwdHtml)
		if len(salts) <= 1 {
			continue
		}

		l.salt = salts[1]

		l.randomPassword = randomString(64) + l.Password
		l.iv = randomString(16)

		encrypted, err := openssl.AesCBCEncrypt([]byte(l.randomPassword), []byte(l.salt), []byte(l.iv), "PKCS7")

		if err != nil {
			continue
		}
		l.encryptedPassword = base64.StdEncoding.EncodeToString(encrypted)

		postData := map[string]string{
			"username":  l.Username,
			"password":  l.encryptedPassword,
			"captcha":   "",
			"_eventId":  "submit",
			"cllt":      "userNameLogin",
			"dllt":      "generalLogin",
			"lt":        "",
			"execution": l.execution,
		}

		_, err = l.client.Client.R().SetFormData(postData).Post(fmt.Sprintf("https://ids.xmu.edu.cn/authserver/login?service=%s", l.client.Service))
		if err != nil {
			continue
		}

		sessionCookie, ok := utils.GetSessionCookie(l.client.Client.GetClient().Jar.Cookies(lntURLParsed), "session")
		if !ok {
			continue
		}
		l.Session = sessionCookie.Value

		if !ok {
			continue
		}
		return nil
	}
	return errors.New("LoginFailed")
}
