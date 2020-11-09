package tcp

import (
	"errors"
	"fmt"
	"io"
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

func tcpErrorCheck(err error) error {
	if err == nil {
		return nil
	}

	// the connection close by remote
	if err == io.EOF {
		return errRecvEOF
	}

	/*if runtime.GOOS == "windows" {
		// client disconnect. If panicErr is nil, then err is errDisconnect, otherwise err is original error.
		panicErr := panicToError(func() {
			if errno, ok := err.(*net.OpError).Err.(*os.SyscallError).Err.(syscall.Errno); ok && syscall.Errno(errno) == syscall.WSAECONNRESET {
				err = errRemoteForceDisconnect
			}
		})
		// record the panicErr
		if panicErr != nil {
			log.Error(panicErr.Error())
		}
		return err
	}*/

	return err
}
