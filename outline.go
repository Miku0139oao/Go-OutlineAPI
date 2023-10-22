package outline_Controller

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Keys struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Port      int    `json:"port"`
	Method    string `json:"method"`
	AccessUrl string `json:"accessUrl"`
}

type Payload struct {
	Limit Limit `json:"limit"`
}
type Limit struct {
	Bytes int `json:"bytes"`
}

type Usage struct {
	ID    string
	Value int
}

type Rename struct {
	Name string `json:"name"`
}

type AccessKeys struct {
	Keys []Keys `json:"accessKeys"`
}

var (
	API_URL = ""
)

func SetApi_Url(url string) {
	API_URL = url
}

func RawRequest(Method, EndPoint string, BodyStruct ...io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(Method, API_URL+EndPoint, nil)

	if BodyStruct != nil {
		req, err = http.NewRequest(Method, API_URL+EndPoint, BodyStruct[0])
	}

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return resp, nil
}

func ListAccessKeys() ([]Keys, error) {
	resp, err := RawRequest(http.MethodGet, "/access-keys")
	if err != nil {
		return nil, err
	}
	var Key AccessKeys
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &Key); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return Key.Keys, nil
}

func CreateAccessKey() (Keys, error) {
	resp, err := RawRequest(http.MethodPost, "/access-keys")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var Key Keys
	if err = json.Unmarshal(body, &Key); err != nil {
		fmt.Println(err)
		return Key, err
	}
	return Key, nil
}

func RenameAccessKey(ID, Name string) bool {
	Payload, err := json.Marshal(Rename{Name: Name})
	if err != nil {
		fmt.Println(err)
		//return false
	}
	resp, err := RawRequest(http.MethodPut, "/access-keys/"+ID+"/name", bytes.NewBuffer(Payload))
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return true
	}
	return false
}

func RemoveAccessKey(ID string) bool {
	resp, err := RawRequest(http.MethodDelete, "/access-keys/"+ID)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return true
	}
	return false
}

func SetVPNLimit(ID string, limits int) bool {
	data, err := json.Marshal(Payload{
		Limit: Limit{Bytes: limits * 1000 * 1000},
	})
	resp, err := RawRequest(http.MethodPut, "/access-keys/"+ID+"/data-limit", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(bytes.NewBuffer(data).String())
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return true
	}
	return false
}

func RemoveVPNLimit(ID string) bool {
	resp, err := RawRequest(http.MethodDelete, "/access-keys/"+ID+"/data-limit")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 204 {
		return true
	}
	return false
}

/*
 Check Data Transfer
*/

func Metrics() (res []Usage, err error) {
	resp, err := RawRequest(http.MethodGet, "/metrics/transfer")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println(err)
		return nil, err
	}
	v := result["bytesTransferredByUserId"].(map[string]any)
	for Key, value := range v {
		res = append(res, Usage{
			ID:    Key,
			Value: int(value.(float64)),
		})
	}
	return res, nil
}
