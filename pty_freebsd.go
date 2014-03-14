package pty

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

const (
	SPECNAMELEN = 63 /* max length of devicename <sys/param.h> */
)

func posix_openpt(oflag int) (fd int, err error) {
	r0, _, e1 := syscall.Syscall(syscall.SYS_POSIX_OPENPT, uintptr(oflag), 0, 0)
	fd = int(r0)
	if e1 != 0 {
		err = e1
	}
	return
}

func open() (pty, tty *os.File, err error) {
	fd, err := posix_openpt(syscall.O_RDWR)
	if err != nil {
		return nil, nil, err
	}

	p := os.NewFile(uintptr(fd), "/dev/pts")
	sname, err := ptsname(p)
	if err != nil {
		return nil, nil, err
	}

	t, err := os.OpenFile("/dev/"+sname, os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, t, nil
}

func isptmaster(fd uintptr) (bool, error) {
	var result int
	err := ioctl(fd, syscall.TIOCPTMASTER, uintptr(unsafe.Pointer(&result)))
	return (result == 0), err
}

// from <sys/filio.h>
type fiodgnameArg struct {
	Len _C_int
	Buf uintptr
}

var (
	emptyFiodgnameArg fiodgnameArg
	ioctl_FIODGNAME   = _IOW('f', 120, unsafe.Sizeof(emptyFiodgnameArg))
)

func ptsname(f *os.File) (string, error) {
	master, err := isptmaster(f.Fd())
	if err != nil {
		return "", err
	}
	if !master {
		return "", syscall.EINVAL
	}

	var (
		buf [SPECNAMELEN + 1]byte
		arg = &fiodgnameArg{_C_int(len(buf)), uintptr(unsafe.Pointer(&buf[0]))}
	)
	err = ioctl(f.Fd(), ioctl_FIODGNAME, uintptr(unsafe.Pointer(arg)))
	if err != nil {
		return "", err
	}

	for i, c := range buf {
		if c == 0 {
			return string(buf[:i]), nil
		}
	}
	return "", errors.New("FIODGNAME string not NUL-terminated")
}
