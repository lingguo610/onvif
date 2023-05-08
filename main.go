/*
 * @Author: error: git config user.name && git config user.email & please set dead value or install git
 * @Date: 2022-05-16 19:26:36
 * @LastEditors: error: git config user.name && git config user.email & please set dead value or install git
 * @LastEditTime: 2022-09-01 10:46:01
 * @FilePath: \goproject\useonvif\main.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package onvif

import (
	"fmt"

	"github.com/lingguo610/onvif"
	onvif_device "github.com/lingguo610/onvif/device"
)

const (
	login    = "admin"
	password = "Hytera1993"
)

func main() {

	var device onvif.DevInterface
	device = &onvif_device.OnvifDevice{}
	device.SetAuth(login, password, "192.168.41.14")
	i, err := device.GetMedialUri()
	if err != nil {
		fmt.Println("GetMedialUri fail, err:%v ", err)
	} else {
		fmt.Println("i:%v", i)
	}

	device.PTZContinuesMove(onvif.RIGHT)

	forever := make(chan bool)
	<-forever
}
