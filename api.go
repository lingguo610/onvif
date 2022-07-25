/*
 * @Author: error: git config user.name && git config user.email & please set dead value or install git
 * @Date: 2022-07-25 13:49:49
 * @LastEditors: error: git config user.name && git config user.email & please set dead value or install git
 * @LastEditTime: 2022-07-25 16:57:59
 * @FilePath: \onvif\api.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
/*
 * @Author: error: git config user.name && git config user.email & please set dead value or install git
 * @Date: 2022-07-25 13:31:31
 * @LastEditors: error: git config user.name && git config user.email & please set dead value or install git
 * @LastEditTime: 2022-07-25 16:21:18
 * @FilePath: \go-onvif\api.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package onvif

type CommandType int

const (
	LEFT CommandType = iota
	RIGHT
	UP
	DOWN
	ZOOM_IN
	ZOOM_OUT
)

//onvif设备对外提供的接口
type DevInterface interface {
	SetAuth(user, passwd, devIp string)
	PTZContinuesMove(CommandType) error
	GetMediaUri() (string, error)
}
