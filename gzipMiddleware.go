package whttp

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
)

func GZIPMiddleware(level int) func(*HTTPContext) {
	return func(c *HTTPContext) {
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			f := func(b *bytes.Buffer) *bytes.Buffer {
				c.Writer.Header().Set("Content-Encoding", "gzip")
				c.Writer.Header().Set("Vary", "Accept-Encoding")
				buf := new(bytes.Buffer)
				gzip, err := gzip.NewWriterLevel(buf, level)
				if err != nil {
					c.Error("gzipMiddleware new failed", "error", err)
					return nil
				}
				_, err = io.Copy(gzip, b)
				if err != nil {
					c.Error("gzipMiddleware copy failed", "error", err)
					return nil
				}
				err = gzip.Close()
				if err != nil {
					c.Error("gzipMiddleware close failed", "error", err)
					return nil
				}

				return buf
			}
			c.HookBeforWriteHeader = append(c.HookBeforWriteHeader, f)
		}
		c.Next()
	}

}
