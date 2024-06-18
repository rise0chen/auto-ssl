package qiniu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
)

var (
	Host = "http://api.qiniu.com"
)

func request(mac *auth.Credentials, method string, path string, body interface{}) (resData []byte,
	err error) {
	urlStr := fmt.Sprintf("%s%s", Host, path)
	reqData, _ := json.Marshal(body)
	req, reqErr := http.NewRequest(method, urlStr, bytes.NewReader(reqData))
	if reqErr != nil {
		err = reqErr
		return
	}

	accessToken, signErr := mac.SignRequest(req)
	if signErr != nil {
		err = signErr
		return
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = respErr
		return
	}
	defer resp.Body.Close()

	resData, ioErr := ioutil.ReadAll(resp.Body)
	if ioErr != nil {
		err = ioErr
		return
	}

	return
}
