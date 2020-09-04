package conn

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

const maxMessages = 100
const structSize = int(unsafe.Sizeof(cmsg{}))

var recvMessages = make([]*unix.ReceiveResp, 0, maxMessages)
var rrs = make([]*unix.ReceiveResp, maxMessages)

type cmsg struct {
	cmsghdr unix.Cmsghdr
	pktinfo unix.Inet4Pktinfo
}

func init() {
	for i := 0; i < maxMessages; i++ {
		var rBuff [unix.MaxSegmentSize]byte
		rr := unix.ReceiveResp{P: rBuff[:]}
		//rr := unix.ReceiveResp{P: rBuff[:], Oob: (*[structSize]byte)(unsafe.Pointer(&cmsg{}))[:]}
		rrs[i] = &rr
	}
}

func receive4msgs(sock int, buff []byte, end *NativeEndpoint) (int, error) {
	if len(recvMessages) == 0 {
		for i := 0; i < maxMessages; i++ {
			rrs[i].Oob = (*[structSize]byte)(unsafe.Pointer(&cmsg{}))[:]
		}
		size, err := unix.Recvmmsg(sock, rrs, unix.MSG_WAITFORONE)
		//fmt.Printf("Number of packets: %d\n", size)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return 0, err
		}

		for i := 0; i < size; i++ {
			recvMessages = append(recvMessages, rrs[i])
		}
	}

	var r *unix.ReceiveResp
	r, recvMessages = recvMessages[0], recvMessages[1:]

	if r.Err != nil {
		fmt.Printf("Error: %v", r.Err)
		return 0, r.Err
	}
	copy(buff[:r.Size], r.P[:])

	end.isV6 = false

	if newDst4, ok := r.From.(*unix.SockaddrInet4); ok {
		*end.dst4() = *newDst4
	}

	var oob [structSize]byte
	copy(oob[:], r.Oob)
	cmsg := (*cmsg)(unsafe.Pointer(&oob))
	if cmsg.cmsghdr.Level == unix.IPPROTO_IP &&
		cmsg.cmsghdr.Type == unix.IP_PKTINFO &&
		cmsg.cmsghdr.Len >= unix.SizeofInet4Pktinfo {
		end.src4().Src = cmsg.pktinfo.Spec_dst
		end.src4().Ifindex = cmsg.pktinfo.Ifindex
	}

	return r.Size, r.Err
}
