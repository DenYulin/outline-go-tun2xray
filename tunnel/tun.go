package tunnel

import (
	"errors"
	"github.com/eycorsican/go-tun2socks/common/log"
	"os"

	_ "github.com/eycorsican/go-tun2socks/common/log/simple" // Import simple log for the side effect of making logs printable.
	"golang.org/x/sys/unix"
)

const vpnMtu = 1500

// MakeTunFile returns an os.File object from a TUN file descriptor `fd`.
// The returned os.File holds a separate reference to the underlying file,
// so the file will not be closed until both `fd` and the os.File are
// separately closed.  (UNIX only.)
func MakeTunFile(fd int) (*os.File, error) {
	if fd < 0 {
		log.Errorf("Must provide a valid TUN file descriptor")
		return nil, errors.New("must provide a valid TUN file descriptor")
	}
	// Make a copy of `fd` so that os.File's finalizer doesn't close `fd`.
	newFd, err := unix.Dup(fd)
	if err != nil {
		log.Errorf("make a copy of `fd` error, error: %s", err.Error())
		return nil, err
	}
	file := os.NewFile(uintptr(newFd), "")
	if file == nil {
		log.Errorf("Failed to open TUN file descriptor, fd: %d", newFd)
		return nil, errors.New("failed to open TUN file descriptor")
	}

	log.Infof("Success to make a new tun file, fd: %d", fd)
	return file, nil
}

// ProcessInputPackets reads packets from a TUN device `tun` and writes them to `tunnel`.
func ProcessInputPackets(tunnel Tunnel, tun *os.File) {
	buffer := make([]byte, vpnMtu)
	for tunnel.IsConnected() {
		len, err := tun.Read(buffer)
		if err != nil {
			log.Warnf("Failed to read packet from TUN: %v", err)
			continue
		}
		if len == 0 {
			log.Infof("Read EOF from TUN")
			continue
		}
		if code, err := tunnel.Write(buffer); err != nil {
			log.Errorf("Failed to write msg to tunnel, code: %d, error: %+v", code, err)
		}
	}
}
