package evan

func NewSugarClient(client *Client) *SugarClient {
	return &SugarClient{
		client: client,
	}
}

type SugarClient struct {
	client *Client
}
