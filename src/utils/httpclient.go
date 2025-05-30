package utils

import "net/http"

type IClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var httpClient IClient

func HttpClient() IClient {
	return &http.Client{}
	// if httpClient == nil {
	// 	httpClient = &http.Client{}
	// }
	// return httpClient
}

func SetHttpClient(client IClient) {
	httpClient = client
}
