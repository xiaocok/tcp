/**
 * Copyright gitteamer 2020
 * @date: 2020/11/10
 * @note: error
 */
package tcp

import (
	"errors"
	"fmt"
)

/**
 * error
 */
var (
	errRemoteForceDisconnect = errors.New("An existing connection was forcibly closed by the remote host.")
	errRecvEOF = errors.New("receive EOF from connect.")
	errDataLenOvertakeMaxLen = errors.New("receive data length overtake max data length.")
)

/**
 * @describe:	Do some function call that program might crash
 * @Reference:
 * Copyright 2018 fatedier, fatedier@gmail.com
 * https://github.com/fatedier/frp
 * http://www.apache.org/licenses/LICENSE-2.0
 */
func panicToError(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic error: %v", r)
		}
	}()

	fn()
	return
}

/**
 * author gitteamer 2020/11/10
 * check io.EOF error
 */
/*func tcpErrorCheck(err error) error {
	if err == nil {
		return nil
	}

	// the connection close by remote
	if err == io.EOF {
		return errRecvEOF
	}

	return err
}*/
