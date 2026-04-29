package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"bank-api/internal/integration"
)

func TestCBRClient_GetKeyRate(t *testing.T) {
	xmlResponse := `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	  <soap:Body>
	    <KeyRateResponse xmlns="http://web.cbr.ru/">
	      <KeyRateResult>
	        <xsd:schema>schema</xsd:schema>xml<diffgr:diffgram xmlns:diffgr="urn:schemas-microsoft-com:xml-diffgram-v1">
	          <KeyRate xmlns="">
	            <KR>
	              <Rate>7.50</Rate>
	            </KR>
	          </KeyRate>
	        </diffgr:diffgram>
	      </KeyRateResult>
	    </KeyRateResponse>
	  </soap:Body>
	</soap:Envelope>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.Write([]byte(xmlResponse))
	}))
	defer server.Close()

	integration.CBRServiceURL = server.URL
	defer func() { integration.CBRServiceURL = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx" }()

	client := integration.NewCBRClient()
	rate, err := client.GetKeyRate()
	assert.NoError(t, err)
	assert.Equal(t, 12.5, rate) // 7.5 + маржа 5
}
