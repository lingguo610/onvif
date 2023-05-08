
package onvif

import "github.com/lingguo610/onvif/device"


type CommandType int

const (
	LEFT CommandType = iota
	RIGHT
	UP
	DOWN
	ZOOM_IN
	ZOOM_OUT
	STOP
)

//onvif设备对外提供的接口
type DevInterface interface {
	SetAuth(user, passwd, devIp string)
	PTZContinuesMove(CommandType) error
	GetMediaUri() (string, error)
}

func NewDevice() DevInterface {
    return &device.OnvifDevice{}
}
