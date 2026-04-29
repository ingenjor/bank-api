package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
)

var CBRServiceURL = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"

type CBRClient struct {
	client *http.Client
}

func NewCBRClient() *CBRClient {
	return &CBRClient{client: &http.Client{Timeout: 10 * time.Second}}
}

func (c *CBRClient) GetKeyRate() (float64, error) {
	from := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	soap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <KeyRate xmlns="http://web.cbr.ru/">
      <fromDate>%s</fromDate>
      <ToDate>%s</ToDate>
    </KeyRate>
  </soap12:Body>
</soap12:Envelope>`, from, to)
	req, err := http.NewRequest("POST", CBRServiceURL, bytes.NewBufferString(soap))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cbr request error: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(body); err != nil {
		return 0, fmt.Errorf("xml parse error: %w", err)
	}
	elements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(elements) == 0 {
		return 0, errors.New("key rate data not found")
	}
	latest := elements[0]
	rateEl := latest.FindElement("./Rate")
	if rateEl == nil {
		return 0, errors.New("Rate element missing")
	}
	var rate float64
	fmt.Sscanf(rateEl.Text(), "%f", &rate)
	return rate + 5.0, nil
}
