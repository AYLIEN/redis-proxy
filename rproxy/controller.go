package rproxy

import "log"

const (
	PROXY_RUNNING = iota
	PROXY_PAUSING
	PROXY_PAUSED
	PROXY_RELOADING

	CMD_PAUSE = iota
	CMD_UNPAUSE
	CMD_RELOAD
)

type ProxyController struct {
	requestPermissionChannel chan chan bool
	releasePermissionChannel chan bool
	infoChannel              chan chan *ControllerInfo
	commandChannel           chan int
	proxy                    *RedisProxy
}

func NewProxyController() *ProxyController {
	return &ProxyController{
		requestPermissionChannel: make(chan chan bool),
		releasePermissionChannel: make(chan bool), // TODO: buffer responses?
		infoChannel:              make(chan chan *ControllerInfo),
		commandChannel:           make(chan int)}
}

func (controller *ProxyController) run() {
	activeRequests := 0
	state := PROXY_RUNNING
	requestPermissionChannel := controller.requestPermissionChannel

	for {
		requestPermissionChannel = nil
		switch state {
		case PROXY_RUNNING:
			requestPermissionChannel = controller.requestPermissionChannel
		case PROXY_PAUSING:
			if activeRequests == 0 {
				state = PROXY_PAUSED
				continue
			}
		case PROXY_RELOADING:
			if activeRequests == 0 {
				controller.proxy.ReloadConfig()
				state = PROXY_RUNNING
				continue
			}
		case PROXY_PAUSED:
			// nothing
		}
		select {
		// In states other than PROXY_RUNNING
		// requestPermissionChannel is nil, so the controller
		// will not receive any requests for permission.
		case permCh := <-requestPermissionChannel:
			permCh <- true
			activeRequests++
		case _ = <-controller.releasePermissionChannel:
			activeRequests--

		case stateCh := <-controller.infoChannel:
			stateCh <- &ControllerInfo{
				ActiveRequests: activeRequests,
				State:          state,
				Config:         controller.proxy.config}

		case cmd := <-controller.commandChannel:
			switch cmd {
			case CMD_PAUSE:
				state = PROXY_PAUSING
			case CMD_UNPAUSE:
				state = PROXY_RUNNING
			case CMD_RELOAD:
				state = PROXY_RELOADING
			default:
				log.Print("Unknown controller command:", cmd)
			}
		}
	}
}

func (controller *ProxyController) enterExecution() {
	ch := make(chan bool)
	controller.requestPermissionChannel <- ch
	<-ch
}

func (controller *ProxyController) leaveExecution() {
	controller.releasePermissionChannel <- true
}

func (controller *ProxyController) ExecuteCall(block func() ([]byte, error)) ([]byte, error) {
	controller.enterExecution()
	defer controller.leaveExecution()

	return block()
}

func (controller *ProxyController) GetInfo() *ControllerInfo {
	ch := make(chan *ControllerInfo)
	controller.infoChannel <- ch
	return <-ch
}

func (controller *ProxyController) Pause() {
	controller.commandChannel <- CMD_PAUSE
}

func (controller *ProxyController) Unpause() {
	controller.commandChannel <- CMD_UNPAUSE
}

func (controller *ProxyController) Reload() {
	controller.commandChannel <- CMD_RELOAD
}

////////////////////////////////////////
// ControllerInfo

type ControllerInfo struct {
	ActiveRequests int
	State          int
	Config         *RedisProxyConfig
}

func (ci *ControllerInfo) StateStr() string {
	switch ci.State {
	case PROXY_RUNNING:
		return "running"
	case PROXY_PAUSING:
		return "pausing"
	case PROXY_PAUSED:
		return "paused"
	case PROXY_RELOADING:
		return "reloading"
	default:
		return "unknown"
	}
}