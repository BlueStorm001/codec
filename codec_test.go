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
	"log"
	"testing"
	"time"
)

func TestBody(t *testing.T) {
	var body []byte
	packet := NewPacketFieldLength(4)
	packet.Receiver = func(body []byte) {
		if t != nil {
			log.Println(string(body))
		}
	}
	body = append(body, packet.PacketFieldLengthEncode([]byte("CCC"))...)
	test1 := []byte("Hello Good ")
	test2 := []byte("He")
	body = append(body, packet.PacketFieldFrameLength(test1)...)
	body = append(body, test2...)

	packet.PacketFieldLengthDecode(body)

	body = body[:0]
	test2 = []byte("llo")
	body = append(body, test2...)
	packet.PacketFieldLengthDecode(body)

	body = body[:0]
	test2 = []byte(" Good ")
	body = append(body, test2...)
	//packet.PacketFieldLengthDecode(body)

	test1 = []byte("Yes, we're fine")
	test2 = []byte("Yes, we")
	body = append(body, packet.PacketFieldFrameLength(test1)...)
	body = append(body, test2...)

	packet.PacketFieldLengthDecode(body)

	body = body[:0]
	test2 = []byte("'re fine")
	body = append(body, test2...)
	packet.PacketFieldLengthDecode(body)
}

func TestHead(t *testing.T) {
	var body []byte
	packet := NewPacketFieldLength(4)
	packet.Receiver = func(body []byte) {
		if t != nil {
			log.Println(string(body))
		}
	}
	body = append(body, packet.PacketFieldLengthEncode([]byte("CCC"))...)

	test1 := []byte("Hello Good ")
	headBuf := packet.PacketFieldFrameLength(test1)
	body = append(body, headBuf[:2]...)
	packet.PacketFieldLengthDecode(body)
	body = body[:0]
	body = append(body, headBuf[2:]...)
	body = append(body, test1...)
	packet.PacketFieldLengthDecode(body)
	body = body[:0]

	test1 = []byte("We all have a home")
	headBuf = packet.PacketFieldFrameLength(test1)
	body = append(body, headBuf...)
	packet.PacketFieldLengthDecode(body)
	body = body[:0]
	body = append(body, test1...)
	packet.PacketFieldLengthDecode(body)
}

func TestPacket(t *testing.T) {
	TestBody(t)
	TestHead(t)
	time.Sleep(time.Second * 1)
	log.Println("================================")
	time.Sleep(time.Second * 1)
	TestHead(t)
	TestBody(t)
	time.Sleep(time.Second * 2)
}

func BenchmarkPacket(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			TestBody(nil)
			TestHead(nil)
		}
	})
}
