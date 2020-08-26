package conn

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
)

func BenchmarkNativeBind_Send(b *testing.B) {
	b.ReportAllocs()
	bind, _, err := createBind(uint16(getFreePort()))
	if err != nil {
		b.Fatal(err)
	}
	end, _ := CreateEndpoint("127.0.0.1:" + fmt.Sprintf("%d", getFreePort()))
	token := make([]byte, 1452)
	rand.Read(token)
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			err := bind.Send(token, end)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkNativeBind_BufferedSend(b *testing.B) {
	b.ReportAllocs()
	bind, _, err := createBind(uint16(getFreePort()))
	if err != nil {
		b.Fatal(err)
	}
	end, _ := CreateEndpoint("127.0.0.1:" + fmt.Sprintf("%d", getFreePort()))
	token := make([]byte, 1452)
	rand.Read(token)
	for n := 0; n < b.N; n++ {
		for i := 0; i < 1000; i++ {
			err := bind.SendBuffered(token, end)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func getFreePort() int {
	l, _ := net.ListenPacket("udp", "localhost:0")
	defer l.Close()
	return l.LocalAddr().(*net.UDPAddr).Port
}
