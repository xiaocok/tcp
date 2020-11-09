package tcp

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

/**
 * gob encode
 * @param *Message msg:		a struct pointer for encode
 * @return *bytes.Buffer:	a bytes buffer by encode, could get bytes by use buffer.Bytes()
 * @return error
 */
func encode(msg *Message) (*bytes.Buffer, error) {
	// buffer struct realization io.Writer interface
	var buf bytes.Buffer

	// gob encoder obj, need realization io.Writer interface
	enc := gob.NewEncoder(&buf)

	// gob encode data
	if err := enc.Encode(msg); err != nil {
		return nil, err
	}

	// return encode buffer
	return &buf, nil
}

/**
 * gob decode
 * @param []byte data:	the bytes data for decode
 * @param *Message msg:	the receive obj struct pointer for decode
 * @return error
 */
func decode(data []byte, msg *Message) error {
	// buffer struct realization io.Writer interface
	var buffer = bytes.NewBuffer(data)

	// new gob decoder obj, need realization io.Writer interface
	dec := gob.NewDecoder(buffer)

	// decode data, obj is receive struct pointer
	if err := dec.Decode(msg); err != nil {
		return err
	}
	return nil
}

/**
 * pack message pointer send to client
 * @param interface{} msg:	message pointer for send
 * @return *bytes.Buffer:	the data buffer obj by pack
 * @return error
 */
func pack(msg *Message) (*bytes.Buffer, error) {
	// encode data
	buf, err := encode(msg)
	if err != nil {
		return nil, err
	}
	n := int64(buf.Len())
	if n > maxMessageLength {
		return nil, errDataLenOvertakeMaxLen
	}

	// write data length
	buffer := bytes.NewBuffer(nil)
	binary.Write(buffer, binary.BigEndian, n)

	// write data
	buffer.Write(buf.Bytes())
	return buffer, nil
}

/**
 * read data form connect
 * @param io.Reader r:	the io.Reader interface obj for read data
 * @param *Message mes:	the message struct pointer for receive the decode data
 */
func read(r io.Reader, msg *Message) error {
	// read data length
	var length int64
	err := binary.Read(r, binary.BigEndian, &length)
	// check tcp error
	if err = tcpErrorCheck(err); err != nil {
		return err
	}

	if length < 0 {
		return fmt.Errorf("read error, data length < 0, length=%d", length)
	}
	if length > maxMessageLength {
		return errDataLenOvertakeMaxLen
	}

	// rend data
	buffer := make([]byte, length)
	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return fmt.Errorf("read error=%s", err.Error())
	}
	if int64(n) != length {
		return fmt.Errorf("read error")
	}

	// decode data
	return decode(buffer, msg)
}
