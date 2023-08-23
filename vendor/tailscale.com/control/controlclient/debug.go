// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package controlclient

import (
	"bytes"
	"compress/gzip"
	"context"
	"log"
	"net/http"
	"time"

	"tailscale.com/util/goroutines"
)

func dumpGoroutinesToURL(c *http.Client, targetURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	zbuf := new(bytes.Buffer)
	zw := gzip.NewWriter(zbuf)
	zw.Write(goroutines.ScrubbedGoroutineDump(true))
	zw.Close()

	req, err := http.NewRequestWithContext(ctx, "PUT", targetURL, zbuf)
	if err != nil {
		log.Printf("dumpGoroutinesToURL: %v", err)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	t0 := time.Now()
	_, err = c.Do(req)
	d := time.Since(t0).Round(time.Millisecond)
	if err != nil {
		log.Printf("dumpGoroutinesToURL error: %v to %v (after %v)", err, targetURL, d)
	} else {
		log.Printf("dumpGoroutinesToURL complete to %v (after %v)", targetURL, d)
	}
}
