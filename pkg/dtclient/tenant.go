package dtclient

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TenantInfo struct {
	ID                    string
	Token                 string
	Endpoints             []string
	CommunicationEndpoint string
}

func (dc *dynatraceClient) GetTenantInfo() (*TenantInfo, error) {
	url := fmt.Sprintf("%s/v1/deployment/installer/agent/connectioninfo", dc.url)
	response, err := dc.makeRequest(
		url,
		dynatracePaaSToken,
	)

	if err != nil {
		return nil, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logger.Error(err, err.Error())
		}
	}()

	data, err := dc.getServerResponseData(response)
	if err != nil {
		err = dc.handleErrorResponseFromAPI(data, response.StatusCode)
		if err != nil {
			logger.Error(err, err.Error())
		}
		return nil, err
	}

	tenantInfo, err := dc.readResponseForTenantInfo(data)
	if err != nil {
		logger.Error(err, err.Error())
		return nil, err
	}
	if len(tenantInfo.Endpoints) <= 0 {
		logger.Info("tenant has no endpoints")
	}

	tenantInfo.CommunicationEndpoint = tenantInfo.findCommunicationEndpoint()
	return tenantInfo, nil
}

func (dc *dynatraceClient) readResponseForTenantInfo(response []byte) (*TenantInfo, error) {
	type jsonResponse struct {
		TenantUUID             string
		TenantToken            string
		CommunicationEndpoints []string
	}

	jr := &jsonResponse{}
	err := json.Unmarshal(response, jr)
	if err != nil {
		logger.Error(err, "error unmarshalling json response")
		return nil, err
	}

	return &TenantInfo{
		ID:        jr.TenantUUID,
		Token:     jr.TenantToken,
		Endpoints: jr.CommunicationEndpoints,
	}, nil
}

func (tenantInfo *TenantInfo) findCommunicationEndpoint() string {
	endpointIndex := tenantInfo.findCommunicationEndpointIndex()
	if endpointIndex < 0 {
		return ""
	}

	endpoint := tenantInfo.Endpoints[endpointIndex]
	if !strings.HasSuffix(endpoint, DT_COMMUNICATION_SUFFIX) {
		if !strings.HasSuffix(endpoint, SLASH) {
			endpoint += SLASH
		}
		endpoint += DT_COMMUNICATION_SUFFIX
	}

	return endpoint
}

func (tenantInfo *TenantInfo) findCommunicationEndpointIndex() int {
	if len(tenantInfo.Endpoints) <= 0 {
		return -1
	}
	for i, endpoint := range tenantInfo.Endpoints {
		if strings.Contains(endpoint, tenantInfo.ID) {
			return i
		}
	}
	return 0
}

const (
	SLASH                   = "/"
	DT_COMMUNICATION_SUFFIX = "communication"
)
