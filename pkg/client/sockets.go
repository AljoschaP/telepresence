package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
)

// DialSocket dials the given socket and returns the resulting connection.
func DialSocket(ctx context.Context, socketName string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return dialSocket(ctx, socketName, opts...)
}

// ListenSocket returns a listener for the given socket and returns the resulting connection.
func ListenSocket(ctx context.Context, processName, socketName string) (net.Listener, error) {
	return listenSocket(ctx, processName, socketName)
}

// RemoveSocket removes any representation of the socket from the filesystem.
func RemoveSocket(listener net.Listener) error {
	return removeSocket(listener)
}

// SocketExists returns true if a socket is found with the given name.
func SocketExists(name string) (bool, error) {
	return socketExists(name)
}

// WaitUntilSocketVanishes waits until the socket at the given path is removed
// and returns when that happens. The wait will be max ttw (time to wait) long.
// An error is returned if that time is exceeded before the socket is removed.
func WaitUntilSocketVanishes(name, path string, ttw time.Duration) error {
	giveUp := time.Now().Add(ttw)
	for giveUp.After(time.Now()) {
		if exists, err := SocketExists(path); err != nil || !exists {
			return err
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timeout while waiting for %s to exit", name)
}

// WaitUntilSocketAppears waits until the socket at the given path comes into
// existence and returns when that happens. The wait will be max ttw (time to wait) long.
func WaitUntilSocketAppears(name, path string, ttw time.Duration) error {
	giveUp := time.Now().Add(ttw)
	for giveUp.After(time.Now()) {
		if exists, err := SocketExists(path); err != nil || exists {
			return err
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timeout while waiting for %s to start", name)
}

// WaitUntilRunning waits until the socket at the given path comes into
// existence and a dial is successful and returns when that happens. The wait will
// be max ttw (time to wait) long.
func WaitUntilRunning(ctx context.Context, name, path string, ttw time.Duration) error {
	giveUp := time.Now().Add(ttw)
	if err := WaitUntilSocketAppears(name, path, ttw); err != nil {
		return err
	}
	for giveUp.After(time.Now()) {
		running, err := IsRunning(ctx, path)
		if err != nil || running {
			return err
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("timeout while waiting for %s to respond", name)
}

// IsRunning makes an attempt to dial the given socket and returns true if that
// succeeds. If the attempt doesn't succeed the method returns false. No error is
// returned when the failed attempt is caused by a non-existing socket.
func IsRunning(ctx context.Context, path string) (bool, error) {
	conn, err := DialSocket(ctx, path)
	switch {
	case err == nil:
		conn.Close()
		return true, nil
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	default:
		return false, err
	}
}
