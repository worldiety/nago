// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

//go:build linux

package sbox

import (
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

// nativeEndian is the host byte order, used to build netlink messages.
var nativeEndian binary.ByteOrder = func() binary.ByteOrder {
	var i uint16 = 1
	if *(*byte)(unsafe.Pointer(&i)) == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}()

// netInterfaceByName returns the interface index for the named link inside the
// current network namespace.
func netInterfaceByName(name string) (int, error) {
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		return 0, fmt.Errorf("interface %q: %w", name, err)
	}
	return ifi.Index, nil
}

// setLinkUp sets IFF_UP on the interface identified by index using a netlink
// RTM_NEWLINK message.
func setLinkUp(index int) error {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		return fmt.Errorf("netlink socket: %w", err)
	}
	defer unix.Close(fd)

	if err := unix.Bind(fd, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return fmt.Errorf("netlink bind: %w", err)
	}

	// struct ifinfomsg is 16 bytes:
	//   __u8  ifi_family; __u8 __ifi_pad; __u16 ifi_type;
	//   __s32 ifi_index; __u32 ifi_flags; __u32 ifi_change;
	const ifinfomsgLen = 16
	msgLen := unix.NLMSG_HDRLEN + ifinfomsgLen
	buf := make([]byte, msgLen)

	// nlmsghdr
	nativeEndian.PutUint32(buf[0:4], uint32(msgLen))                            // nlmsg_len
	nativeEndian.PutUint16(buf[4:6], uint16(unix.RTM_NEWLINK))                  // nlmsg_type
	nativeEndian.PutUint16(buf[6:8], uint16(unix.NLM_F_REQUEST|unix.NLM_F_ACK)) // nlmsg_flags
	nativeEndian.PutUint32(buf[8:12], 1)                                        // nlmsg_seq
	nativeEndian.PutUint32(buf[12:16], 0)                                       // nlmsg_pid

	// ifinfomsg starts at NLMSG_HDRLEN (16)
	off := unix.NLMSG_HDRLEN
	buf[off] = unix.AF_UNSPEC // ifi_family
	// ifi_index at off+4
	nativeEndian.PutUint32(buf[off+4:off+8], uint32(index))
	// ifi_flags at off+8: IFF_UP
	nativeEndian.PutUint32(buf[off+8:off+12], unix.IFF_UP)
	// ifi_change at off+12: IFF_UP (only change the UP bit)
	nativeEndian.PutUint32(buf[off+12:off+16], unix.IFF_UP)

	if err := unix.Sendto(fd, buf, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return fmt.Errorf("netlink send: %w", err)
	}

	// Read the ACK.
	resp := make([]byte, 4096)
	n, _, err := unix.Recvfrom(fd, resp, 0)
	if err != nil {
		return fmt.Errorf("netlink recv: %w", err)
	}
	if n < unix.NLMSG_HDRLEN {
		return fmt.Errorf("netlink short response")
	}
	msgType := nativeEndian.Uint16(resp[4:6])
	if msgType == unix.NLMSG_ERROR {
		// error code is the first 4 bytes after the header
		errno := int32(nativeEndian.Uint32(resp[unix.NLMSG_HDRLEN : unix.NLMSG_HDRLEN+4]))
		if errno != 0 {
			return fmt.Errorf("netlink RTM_NEWLINK: errno %d", -errno)
		}
	}
	return nil
}
