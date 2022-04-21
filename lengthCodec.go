/*
 * Copyright (c) 2021 BlueStorm
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFINGEMENT IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package codec

import (
	"bytes"
	"encoding/binary"
)

type Packer struct {
	buffer      bytes.Buffer
	headLen     int
	bodyLen     int
	Datagram    []byte
	Receiver    func(body []byte)
	temp        bytes.Buffer
	tempLen     int
	tempHead    bytes.Buffer
	tempHeadLen int
	offset      int
	addHead     bool
}

func NewPacketFieldLength(lengthFieldLength int) *Packer {
	return &Packer{headLen: lengthFieldLength}
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// size (2/4/8 bytes)  | body (size bytes)
//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (pk *Packer) PacketFieldFrameLength(body []byte) []byte {
	pkgLen := len(body)
	var headData = make([]byte, pk.headLen)
	switch pk.headLen {
	case 2:
		binary.BigEndian.PutUint16(headData[:pk.headLen], uint16(pkgLen))
	case 4:
		binary.BigEndian.PutUint32(headData[:pk.headLen], uint32(pkgLen))
	case 8:
		binary.BigEndian.PutUint64(headData[:pk.headLen], uint64(pkgLen))
	}
	return headData
}

func (pk *Packer) PacketFieldLengthEncode(body []byte) (data []byte) {
	pk.buffer.Write(pk.PacketFieldFrameLength(body))
	pk.buffer.Write(body)
	data = pk.buffer.Bytes()
	pk.buffer.Reset()
	return
}

func (pk *Packer) PacketFieldLengthDecode(body []byte) {
	if pk.Receiver == nil {
		return
	}
	pk.Datagram = body
	if pk.tempLen > 0 {
		count := len(pk.Datagram)
		if pk.tempLen <= count {
			pk.temp.Write(pk.Datagram[:pk.tempLen])
			pk.copy(pk.temp.Bytes())
			if pk.tempLen == count {
				pk.reset(0)
				return
			}
			pk.reset(pk.tempLen)
		} else {
			pk.temp.Write(pk.Datagram)
			pk.tempLen = pk.tempLen - len(pk.Datagram)
			return
		}
	}
	if pk.bodyLen == 0 {
		if pk.fieldHeadRead() {
			return
		}
	}
	if pk.bodyLen > 0 {
		for {
			size := len(pk.Datagram[pk.offset:])
			if size < pk.bodyLen {
				if size > 0 {
					data := pk.Datagram[pk.offset : pk.offset+size]
					pk.temp.Write(data)
				}
				pk.offset = 0
				pk.tempLen = pk.bodyLen - size
				pk.Datagram = pk.Datagram[:0]
				return
			} else {
				data := pk.Datagram[pk.offset : pk.offset+pk.bodyLen]
				pk.copy(data)
				if pk.fieldHeadRead() {
					pk.reset(0)
					return
				}
			}
		}
	}
}

func (pk *Packer) fieldHeadRead() bool {
	var data []byte
	var next int
	if pk.tempHeadLen > 0 {
		pk.tempHead.Write(pk.Datagram[:pk.tempHeadLen])
		data = pk.tempHead.Bytes()
		pk.tempHead.Reset()
		next = pk.tempHeadLen
		pk.tempHeadLen = 0
	} else {
		offset := pk.offset + pk.bodyLen
		count := len(pk.Datagram)
		if offset >= count {
			return true
		}
		next = offset + pk.headLen
		if next > count {
			pk.tempHead.Write(pk.Datagram[offset:])
			pk.tempHeadLen = pk.headLen - pk.tempHead.Len()
			return true
		}
		data = pk.Datagram[offset:next]
	}
	switch pk.headLen {
	case 2:
		pk.bodyLen = int(binary.BigEndian.Uint16(data))
	case 4:
		pk.bodyLen = int(binary.BigEndian.Uint32(data))
	case 8:
		pk.bodyLen = int(binary.BigEndian.Uint64(data))
	default:
		return true
	}
	pk.offset = next
	return false
}

func (pk *Packer) copy(data []byte) {
	body := make([]byte, pk.bodyLen)
	copy(body, data)
	go pk.Receiver(body)
}

func (pk *Packer) reset(offset int) {
	pk.offset = offset
	pk.bodyLen = 0
	pk.tempLen = 0
	pk.temp.Reset()
	if pk.offset == 0 {
		pk.Datagram = pk.Datagram[:0]
	}
}
