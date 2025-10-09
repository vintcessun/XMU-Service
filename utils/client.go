package utils

import (
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ServiceClient struct {
	Client  *resty.Client
	Service string
	Html    string
}

var retryTimes = 3

func NewClient(client *resty.Client, url string) (*ServiceClient, error) {
	l := ServiceClient{}

	l.Client = client
	l.Client.SetRedirectPolicy(resty.FlexibleRedirectPolicy(100))

	var retErr error

	for range retryTimes {
		time.Sleep(1 * time.Second)
		resp, err := l.Client.R().Get(url)
		if err != nil {
			retErr = err
			continue
		}
		endUrl := resp.RawResponse.Request.URL.String()
		serviceIndex := strings.Index(endUrl, "service=")
		if serviceIndex == -1 {
			continue
		}
		l.Service = endUrl[(serviceIndex + 8):]

		l.Html = resp.String()

		break
	}

	return &l, retErr
}
