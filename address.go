package tcp

import "strings"

func NewAddr(address string) *Addr {
	addr := new(Addr)
	addr.parseAddr(address)
	return addr
}

/**
 * address info
 */
type Addr struct {
	Ip   string
	Port string
}

/**
 * parse tcp addr: IP:PORT to IP and Port
 * @param string address: 	ip:port
 */
func (a *Addr) parseAddr(address string) {

	params := strings.Split(address, ":")
	if len(params) == 1 {
		a.Ip = params[0]
	} else {
		if params[0] == "" {
			a.Ip = "127.0.0.1"
		} else {
			a.Ip = params[0]
		}
		a.Port = params[1]
	}
}

/**
 * get address: ip:port
 * @return string
 */
func (a *Addr) GetAddress() string {
	return a.Ip + ":" + a.Port
}
