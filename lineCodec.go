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
	"encoding/binary"
)

func NewPacketFieldLine() *Packer {
	return &Packer{headLen: 6}
}

//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//        size (6 bytes)           |
//2 bytes (\r\n) | size (4 bytes)  | body (size bytes)
//+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func (pk *Packer) PacketFieldLineEncode(buffer []byte) []byte {
	defer pk.buffer.Reset()
	pk.buffer.WriteByte('\r')
	pk.buffer.WriteByte('\n')
	pkgLen := uint32(len(buffer))
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:4], pkgLen)
	pk.buffer.Write(buf[:4])
	pk.buffer.Write(buffer)
	return pk.buffer.Bytes()
}

func (pk *Packer) PacketFieldLineDecode(buffer []byte) {
	if pk.Receiver == nil {
		return
	}
	pk.Datagram = buffer
	if pk.tempHeadLen > 0 {
		if len(pk.Datagram) >= pk.tempHeadLen {
			pack := pk.Datagram[:pk.tempHeadLen]
			pk.tempHead.Write(pack)
			spk := pk.Datagram[len(pack):]
			pk.addHead = true
			if len(spk) == 0 {
				pk.tempHeadLen = 0
				return
			} else {
				pk.Datagram = spk
				pk.tempHeadLen = 0
			}
		} else {
			pk.tempHeadLen = pk.tempHeadLen - len(pk.Datagram)
			pk.tempHead.Write(pk.Datagram)
			pk.Datagram = pk.Datagram[:0]
			return
		}
	}
	if pk.addHead {
		pk.addHead = false
		pk.tempHead.Write(pk.Datagram)
		pk.Datagram = pk.tempHead.Bytes()
		pk.tempHead.Reset()
	}
	if pk.tempLen > 0 {
		if len(pk.Datagram) >= pk.tempLen {
			pk.temp.Write(pk.Datagram[:pk.tempLen])
			body := make([]byte, pk.temp.Len())
			copy(body, pk.temp.Bytes())
			go pk.Receiver(body)
			pk.temp.Reset()
			pk.Datagram = pk.Datagram[pk.tempLen:]
			pk.tempLen = 0
		} else {
			pk.tempLen = pk.tempLen - len(pk.Datagram)
			pk.temp.Write(pk.Datagram)
			pk.Datagram = pk.Datagram[:0]
			return
		}
	}
	for {
		length := len(pk.Datagram)
		if length == 0 {
			return
		}
		if length < pk.headLen {
			pk.tempHeadLen = pk.headLen - length
			pk.tempHead.Write(pk.Datagram)
			pk.Datagram = pk.Datagram[:0]
			return
		}
		if pk.Datagram[0] == '\r' && pk.Datagram[1] == '\n' {
			dataCount := int(binary.BigEndian.Uint32(pk.Datagram[pk.headLen-4 : pk.headLen]))
			remaining := length - pk.headLen
			if dataCount <= remaining {
				body := make([]byte, pk.headLen+dataCount-pk.headLen)
				copy(body, pk.Datagram[pk.headLen:pk.headLen+dataCount])
				go pk.Receiver(body)
				pk.Datagram = pk.Datagram[pk.headLen+dataCount:]
			} else {
				pk.temp.Write(pk.Datagram[pk.headLen:])
				pk.tempLen = dataCount - remaining
				break
			}
		} else {
			break
		}
	}

	pk.Datagram = pk.Datagram[:0]
}
