package conn

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

func receive4msgs(sock int, buff []byte, end *NativeEndpoint) (int, error) {

	// construct message header

	var cmsg struct {
		cmsghdr unix.Cmsghdr
		pktinfo unix.Inet4Pktinfo
	}

	size, _, _, newDst, err := unix.Recvmsgs(sock, buff, (*[unsafe.Sizeof(cmsg)]byte)(unsafe.Pointer(&cmsg))[:], unix.MSG_DONTWAIT)

	if err != nil {
		return 0, err
	}
	end.isV6 = false

	if newDst4, ok := newDst.(*unix.SockaddrInet4); ok {
		*end.dst4() = *newDst4
	}

	// update source cache

	if cmsg.cmsghdr.Level == unix.IPPROTO_IP &&
		cmsg.cmsghdr.Type == unix.IP_PKTINFO &&
		cmsg.cmsghdr.Len >= unix.SizeofInet4Pktinfo {
		end.src4().Src = cmsg.pktinfo.Spec_dst
		end.src4().Ifindex = cmsg.pktinfo.Ifindex
	}

	return size, nil
}
