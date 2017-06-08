package rproxy

////////////////////////////////////////
// TestConfigHolder

type TestConfigHolder struct {
	config *RedisProxyConfig
}

func (ch *TestConfigHolder) GetConfig() *RedisProxyConfig {
	return ch.config
}

func (ch *TestConfigHolder) ReloadConfig() {}

////////////////////////////////////////
// TestRequest

type TestRequest struct {
	contr *ProxyController
	done  bool
}

func NewTestRequest(contr *ProxyController) *TestRequest {
	return &TestRequest{contr: contr}
}

func (r *TestRequest) Do() {
	r.contr.CallUplink(func() (*RespMsg, error) {
		return nil, nil
	})
	r.done = true
}
