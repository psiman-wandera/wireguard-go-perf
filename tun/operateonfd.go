// +build !windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2020 WireGuard LLC. All Rights Reserved.
 */

package tun

import (
	"fmt"
)

func (tun *NativeTun) operateOnFd(fn func(fd uintptr)) {
	for i, _ := range tun.tunFiles {
		sysconn, err := tun.tunFiles[i].SyscallConn()
		if err != nil {
			tun.errors <- fmt.Errorf("unable to find sysconn for tunfile: %s", err.Error())
			return
		}
		err = sysconn.Control(fn)
		if err != nil {
			tun.errors <- fmt.Errorf("unable to control sysconn for tunfile: %s", err.Error())
		}
	}
}
