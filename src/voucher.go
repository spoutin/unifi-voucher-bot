package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var authenticationPath = "/api/auth/login"
var voucherPath = "/proxy/network/api/s/default/stat/voucher"

type Unifi struct {
	Client   *http.Client
	BaseURL  *url.URL
	Username string
	Password string
}

func InitVoucherClient(baseUrl string, username string, password string) (*Unifi, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: UnifiSSLVerify},
	}

	client := &http.Client{
		Jar:       jar,
		Transport: tr,
	}
	return &Unifi{
		Client:   client,
		BaseURL:  u,
		Username: username,
		Password: password,
	}, nil
}

func (unifi *Unifi) Authenticate() error {
	u, err := url.JoinPath(unifi.BaseURL.String(), authenticationPath)
	if err != nil {
		return err
	}
	auth := map[string]string{
		"username": unifi.Username,
		"password": unifi.Password,
	}
	marshal, err := json.Marshal(auth)
	if err != nil {
		return err
	}
	response, err := unifi.Client.Post(u, "application/json", bytes.NewBuffer(marshal))
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("unifi authentication failed with status code %d", response.StatusCode)
	}
	logger.Info("unifi authentication successful")
	return nil
}

func (unifi *Unifi) GetVouchers() ([]Voucher, error) {
	// curl -k -X GET -b cookie.txt  https://10.0.0.11/proxy/network/api/s/default/stat/voucher | jq .
	u, err := url.JoinPath(unifi.BaseURL.String(), voucherPath)
	if err != nil {
		return nil, err
	}
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	authenicated := false
	for _, c := range unifi.Client.Jar.Cookies(parsedUrl) {
		if c.Name == "TOKEN" && c.Valid() == nil {
			authenicated = true
			break
		}
	}
	if !authenicated {
		logger.Warn("Authentication is required")
		err = unifi.Authenticate()
		if err != nil {
			return nil, err
		}
	}
	response, err := unifi.Client.Get(u)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unifi voucher failed with status code %d", response.StatusCode)
	}
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	vouchers := new(Vouchers)
	err = json.Unmarshal(bodyBytes, &vouchers)
	if err != nil {
		return nil, err
	}
	return vouchers.Voucher, nil
}

type Vouchers struct {
	Meta struct {
		Rc string `json:"rc"`
	} `json:"meta"`
	Voucher []Voucher `json:"data"`
}

type Voucher struct {
	Duration      int    `json:"duration"`
	QosOverwrite  bool   `json:"qos_overwrite"`
	Note          string `json:"note"`
	Code          string `json:"code"`
	ForHotspot    bool   `json:"for_hotspot"`
	CreateTime    int    `json:"create_time"`
	Quota         int    `json:"quota"`
	SiteId        string `json:"site_id"`
	Id            string `json:"_id"`
	AdminName     string `json:"admin_name"`
	Used          int    `json:"used"`
	Status        string `json:"status"`
	StatusExpires int    `json:"status_expires"`
}
