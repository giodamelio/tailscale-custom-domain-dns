// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package pidowner

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

func ownerOfPID(pid int) (userID string, err error) {
	procHnd, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err == syscall.Errno(0x57) { // invalid parameter, for PIDs that don't exist
		return "", ErrProcessNotFound
	}
	if err != nil {
		return "", fmt.Errorf("OpenProcess: %T %#v", err, err)
	}
	defer windows.CloseHandle(procHnd)

	var tok windows.Token
	if err := windows.OpenProcessToken(procHnd, windows.TOKEN_QUERY, &tok); err != nil {
		return "", fmt.Errorf("OpenProcessToken: %w", err)
	}

	tokUser, err := tok.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("GetTokenUser: %w", err)
	}

	sid := tokUser.User.Sid
	return sid.String(), nil
}
