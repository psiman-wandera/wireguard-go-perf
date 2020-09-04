package conn

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"unsafe"

	"golang.org/x/sys/unix"
)

const maxMessages = 100
const structSize = int(unsafe.Sizeof(cmsg{}))

var rrs = make([]unix.ReceiveResp, maxMessages)
var rrsSize = 0
var rrsIdx = 0

type cmsg struct {
	cmsghdr unix.Cmsghdr
	pktinfo unix.Inet4Pktinfo
}

func init() {
	//ListenAndServe()
	for i := 0; i < maxMessages; i++ {
		var rBuff [unix.MaxSegmentSize]byte
		rr := unix.ReceiveResp{P: rBuff[:]}
		rrs[i] = rr
	}
}

// ListenAndServe starts server based on provided configuration and registers request handlers.
func ListenAndServe() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", 8079), nil)
}

func receive4msgs(sock int, buff []byte, end *NativeEndpoint) (int, error) {
	if rrsIdx == rrsSize {
		for i := 0; i < maxMessages; i++ {
			rrs[i].Oob = (*[structSize]byte)(unsafe.Pointer(&cmsg{}))[:]
		}
		size, err := unix.Recvmmsg(sock, rrs, unix.MSG_WAITFORONE)
		//fmt.Printf("Number of packets: %d\n", size)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return 0, err
		}

		rrsSize = size
		rrsIdx = 0
	}

	var r unix.ReceiveResp
	r = rrs[rrsIdx]
	rrsIdx++

	if r.Err != nil {
		fmt.Printf("Error: %v", r.Err)
		return 0, r.Err
	}
	copy(buff[:r.Size], r.P[:])

	end.isV6 = false

	if newDst4, ok := r.From.(*unix.SockaddrInet4); ok {
		*end.dst4() = *newDst4
	}

	cmsg := (*cmsg)(unsafe.Pointer(&r.Oob))
	if cmsg.cmsghdr.Level == unix.IPPROTO_IP &&
		cmsg.cmsghdr.Type == unix.IP_PKTINFO &&
		cmsg.cmsghdr.Len >= unix.SizeofInet4Pktinfo {
		end.src4().Src = cmsg.pktinfo.Spec_dst
		end.src4().Ifindex = cmsg.pktinfo.Ifindex
	}

	return r.Size, r.Err
}
