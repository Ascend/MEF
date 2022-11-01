// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common the constants used
package common

import "time"

const (
	// WriteDeadline  the write deadline on the websocket connection
	WriteDeadline = 15 * time.Second

	// ReadDeadline  the read deadline on the websocket connection
	ReadDeadline = 15 * time.Second

	// ReadBufferSize the size of buffers reading from the websocket connection
	ReadBufferSize = 1024

	// WriteBufferSize the size of buffers writing to the websocket connection
	WriteBufferSize = 1024
)
