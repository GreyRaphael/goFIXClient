package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// request headers strut
type ReqHeaders struct {
	UA     string `json:"User-Agent"`
	Cookie string `json:"Cookie"`
}

// response text["result"]["data"] struct
type Stock struct {
	SECURITY_CODE string          `json:"SECURITY_CODE"`
	F2            float64         `json:"f2"`
	_             json.RawMessage `json:"-"`
	Volume        int32
}

// response text["result"] struct
type EastResult struct {
	Data []Stock         `json:"data"`
	_    json.RawMessage `json:"-"`
}

// response text struct
type Txt struct {
	Version string          `json:"version"`
	Result  EastResult      `json:"result"`
	_       json.RawMessage `json:"-"`
}

func main() {
	var headers ReqHeaders
	data, err := os.ReadFile("config/headers.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &headers); err != nil {
		panic(err)
	}

	url := "https://datacenter-web.eastmoney.com/api/data/v1/get?callback=jQuery11230023591033888216595_1704874224569&sortColumns=SECURITY_CODE&sortTypes=-1&pageSize=500&pageNumber=1&reportName=RPT_INDEX_TS_COMPONENT&columns=SECUCODE%2CSECURITY_CODE%2CTYPE%2CSECURITY_NAME_ABBR%2CCLOSE_PRICE%2CINDUSTRY%2CREGION%2CWEIGHT%2CEPS%2CBPS%2CROE%2CTOTAL_SHARES%2CFREE_SHARES%2CFREE_CAP&quoteColumns=f2%2Cf3&quoteType=0&source=WEB&client=WEB&filter=(TYPE%3D%223%22)"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", headers.UA)
	req.Header.Set("Cookie", headers.Cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result Txt
	if err := json.Unmarshal(body[44:len(body)-2], &result); err != nil {
		panic(err)
	}

	outfileName := "input/zz500.csv"
	file, _ := os.Create(outfileName)
	defer file.Close()

	// writer := csv.NewWriter(os.Stdout)
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"code", "vol", "price"})
	for _, v := range result.Result.Data {
		if v.SECURITY_CODE[0] == '6' {
			v.SECURITY_CODE = fmt.Sprintf("%s.SS", v.SECURITY_CODE)
		} else {
			v.SECURITY_CODE = fmt.Sprintf("%s.SZ", v.SECURITY_CODE)
		}
		if v.SECURITY_CODE[0:2] == "68" {
			v.Volume = 200
		} else {
			v.Volume = 100
		}
		writer.Write([]string{v.SECURITY_CODE, fmt.Sprintf("%d", v.Volume), fmt.Sprintf("%.2f", v.F2)})
	}
	fmt.Println("write to", outfileName)
}
