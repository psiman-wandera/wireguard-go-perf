package conn

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const maxMessages = 200
const maxTimeoutMs = 3
const diffPerc = 0.3
const maxSegmentSize = (1 << 16) - 1 // largest possible UDP datagram
var currMessages = 100

const structSize = int(unsafe.Sizeof(cmsg{}))

var rrs = make([]*unix.ReceiveResp, maxMessages)
var rrsSize = 0
var rrsIdx = 0

type cmsg struct {
	cmsghdr unix.Cmsghdr
	pktinfo unix.Inet4Pktinfo
}

func init() {
	//ListenAndServe()
	for i := 0; i < maxMessages; i++ {
		var rBuff [maxSegmentSize]byte
		rr := unix.ReceiveResp{P: rBuff[:], Oob: (*[structSize]byte)(unsafe.Pointer(&cmsg{}))[:]}
		rrs[i] = &rr
	}
}

// ListenAndServe starts server based on provided configuration and registers request handlers.
func ListenAndServe() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", 8079), nil)
}

func receive4msgs(sock int, buff []byte, end *NativeEndpoint) (int, error) {
	if rrsIdx == rrsSize {
		var size int
		for {
			ts := unix.NsecToTimespec(maxTimeoutMs * time.Millisecond.Nanoseconds())
			for i := 0; i < rrsSize; i++ {
				rrs[i].Oob = rrs[i].Oob[:0]
			}
			var err error
			size, err = unix.Recvmmsg(sock, rrs[:currMessages], unix.MSG_WAITALL, &ts)
			if err != nil {
				fmt.Printf("Error overall: %v\n", err)
				if err == syscall.EAGAIN || err == syscall.Errno(512) {
					continue
				}
				return 0, err
			}
			break
		}

		if size < currMessages {
			currMessages = int(float32(currMessages) * (1 - diffPerc))
			if currMessages < 10 {
				currMessages = 10
			}
		} else {
			currMessages = int(float32(currMessages) * (1 + diffPerc))
			if currMessages > maxMessages {
				currMessages = maxMessages
			}
		}

		rrsSize = size
		rrsIdx = 0
	}

	var r unix.ReceiveResp
	r = *rrs[rrsIdx]
	rrsIdx++

	if r.Err != nil {
		fmt.Printf("Error in: %v\n", r.Err)
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
