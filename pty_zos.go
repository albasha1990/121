//go:build zos
// +build zos

package pty

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func open() (pty, tty *os.File, err error) {
	ptmxfd, err := unix.Posix_openpt(os.O_RDWR|syscall.O_NOCTTY)
	if err != nil {
		return nil, nil, err
	}

	// Needed for z/OS so that the characters are not garbled if ptyp* is untagged
	cvtreq := unix.F_cnvrt{Cvtcmd: unix.SETCVTON, Pccsid: 0, Fccsid: 1047}
	if _, err = unix.Fcntl(uintptr(ptmxfd), unix.F_CONTROL_CVT, &cvtreq); err != nil {
		return nil, nil, err
	}


	p := os.NewFile(uintptr(ptmxfd), "/dev/ptmx")
	if p == nil {
		return nil, nil, err
	}

	// In case of error after this point, make sure we close the ptmx fd.
	defer func() {
		if err != nil {
			_ = p.Close() // Best effort.
		}
	}()

	sname, err := unix.Ptsname(ptmxfd)
	if err != nil {
		return nil, nil, err
	}

	_, err = unix.Grantpt(ptmxfd)
	if err != nil {
		return nil, nil, err
	}

	if _, err = unix.Unlockpt(ptmxfd); err != nil {
		return nil, nil, err
	}

	ptsfd, err := syscall.Open(sname, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}

	if _, err = unix.Fcntl(uintptr(ptsfd), unix.F_CONTROL_CVT, &cvtreq); err != nil {
		return nil, nil, err
	}

	t := os.NewFile(uintptr(ptsfd), sname)
	if err != nil {
		return nil, nil, err
	}

	return p, t, nil
}

