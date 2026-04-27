package trading

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dracoDevs/go-ebay/internal/utils"
)

type Command interface {
	Body() interface{}
	CallName() string
	ParseResponse([]byte) (Response, error)
}

type Conf struct {
	BaseURL string

	DevId, AppId, CertId string
	RuName, AuthToken    string
	SiteId               int
	Logger               func(...interface{})
}

func (e Conf) Sandbox() Conf {
	e.BaseURL = "https://api.sandbox.ebay.com"
	return e
}

func (e Conf) Production() Conf {
	e.BaseURL = "https://api.ebay.com"
	return e
}

func (e Conf) RunCommand(c Command) (Response, error) {
	ec := request{conf: e, command: c}

	body := new(bytes.Buffer)
	body.Write([]byte(xml.Header))

	if err := xml.NewEncoder(body).Encode(ec); err != nil {
		return GenericResponse{}, err
	}

	if c.CallName() == "EndItem" || c.CallName() == "SetNotificationPreferences" || c.CallName() == "GetItemTransactions" || c.CallName() == "CompleteSale" || c.CallName() == "GetOrders" || c.CallName() == "GetMyeBaySelling" || c.CallName() == "GetMyMessages" || c.CallName() == "GetSellerList" {
		bodyStr := utils.RemoveTagXML(body.String(), c.CallName())
		body = bytes.NewBufferString(bodyStr)
	}

	if e.Logger != nil {
		e.Logger(body.String())
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/ws/api.dll", e.BaseURL), body)
	req.Header.Add("X-EBAY-API-DEV-NAME", e.DevId)
	req.Header.Add("X-EBAY-API-APP-NAME", e.AppId)
	req.Header.Add("X-EBAY-API-CERT-NAME", e.CertId)
	req.Header.Add("X-EBAY-API-CALL-NAME", c.CallName())
	req.Header.Add("X-EBAY-API-SITEID", strconv.Itoa(e.SiteId))
	req.Header.Add("X-EBAY-API-COMPATIBILITY-LEVEL", strconv.Itoa(1173))
	req.Header.Add("Content-Type", "text/xml")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			return GenericResponse{}, urlErr
		}
		return GenericResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		httpErr := httpError{statusCode: resp.StatusCode}
		httpErr.body, _ = io.ReadAll(resp.Body)
		return GenericResponse{}, httpErr
	}

	bodyContents, _ := io.ReadAll(resp.Body)
	if e.Logger != nil {
		e.Logger(string(bodyContents))
	}

	response, err := c.ParseResponse(bodyContents)
	if response.Failure() {
		return response, response.ResponseErrors()
	}

	return response, err
}
