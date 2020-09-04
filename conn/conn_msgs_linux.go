package conn

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

const maxMessages = 100
var recvMessages = make([]*unix.ReceiveResp, 0, maxMessages)

type cmsg struct {
	cmsghdr unix.Cmsghdr
	pktinfo unix.Inet4Pktinfo
}

const structSize = int(unsafe.Sizeof(cmsg{}))

func (f *cmsg) BytesPtr() *[structSize]byte {
	return (*[structSize]byte)(unsafe.Pointer(f))
}

func newCmsg(b *[structSize]byte) *cmsg {
	return (*cmsg)(unsafe.Pointer(b))
}

func receive4msgs(sock int, buff []byte, end *NativeEndpoint) (int, error) {

	if len(recvMessages) == 0 {
		rrs := make([]*unix.ReceiveResp, 0)
		for i := 0; i < maxMessages; i++ {
			cmsg := cmsg{}
			var rBuff [unix.MaxSegmentSize]byte
			rr := unix.ReceiveResp{Oob: cmsg.BytesPtr()[:], P: rBuff[:]}
			rrs = append(rrs, &rr)
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
	cmsg := newCmsg(&oob)
	if cmsg.cmsghdr.Level == unix.IPPROTO_IP &&
		cmsg.cmsghdr.Type == unix.IP_PKTINFO &&
		cmsg.cmsghdr.Len >= unix.SizeofInet4Pktinfo {
		end.src4().Src = cmsg.pktinfo.Spec_dst
		end.src4().Ifindex = cmsg.pktinfo.Ifindex
	}

	return r.Size, r.Err
}
