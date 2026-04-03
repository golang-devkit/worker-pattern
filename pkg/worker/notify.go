package worker

import (
	"net"
	"os"
)

// sdNotify gửi thông báo đến systemd
func sdNotify(state string) error {
	socketPath := os.Getenv("NOTIFY_SOCKET")
	if socketPath == "" {
		// Không chạy dưới systemd, bỏ qua
		return nil
	}

	// Kết nối đến Unix socket của systemd
	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	// Gửi thông báo
	_, err = conn.Write([]byte(state))
	return err
}
