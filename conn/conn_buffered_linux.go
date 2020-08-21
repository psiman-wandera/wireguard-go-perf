package conn

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func (bind *nativeBind) SendBuffered(buff []byte, end Endpoint) error {
	nend := end.(*NativeEndpoint)
	if !nend.isV6 {
		if bind.sock4 == -1 {
			return syscall.EAFNOSUPPORT
		}
		msg, err := get4BufferedMsg(bind.sock4, nend, buff)
		if err != nil {
			return err
		}
		//bind.outboundMsgs = append(bind.outboundMsgs, msg)
		//send4Buffered(bind.sock4, bind.outboundMsgs)
		send4Buffered(bind.sock4, []*unix.Msghdr{msg})
	} else {
		if bind.sock6 == -1 {
			return syscall.EAFNOSUPPORT
		}
		msg, err := get4BufferedMsg(bind.sock6, nend, buff)
		if err != nil {
			return err
		}
		bind.outboundMsgs = append(bind.outboundMsgs, msg)
	}
	return nil
}

func get4BufferedMsg(sock int, end *NativeEndpoint, buff []byte) (*unix.Msghdr, error){
	// construct message header

	cmsg := struct {
		cmsghdr unix.Cmsghdr
		pktinfo unix.Inet4Pktinfo
	}{
		unix.Cmsghdr{
			Level: unix.IPPROTO_IP,
			Type:  unix.IP_PKTINFO,
			Len:   unix.SizeofInet4Pktinfo + unix.SizeofCmsghdr,
		},
		unix.Inet4Pktinfo{
			Spec_dst: end.src4().Src,
			Ifindex:  end.src4().Ifindex,
		},
	}

	return unix.CreateMsg(sock, buff, (*[unsafe.Sizeof(cmsg)]byte)(unsafe.Pointer(&cmsg))[:], end.dst4())
}

func send4Buffered(sock int, msgs []*unix.Msghdr) error {
	_, err := unix.SendMultiMsg(sock, msgs, 0)
	return err
}
