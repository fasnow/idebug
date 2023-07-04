package feishu

import (
	"idebug/plugin"
)

type Req struct {
	Client      *Client
	HttpMethod  string
	ApiPath     string
	Body        interface{}
	QueryParams *plugin.QueryParams
	PathParams  *plugin.PathParams
}
