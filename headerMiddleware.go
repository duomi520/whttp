package whttp

func HeaderMiddleware(h map[string]string) func(*HTTPContext) {
	return func(c *HTTPContext) {
		for k, v := range h {
			c.Writer.Header().Add(k, v)
		}
		c.Next()
	}
}

// https://zhuanlan.zhihu.com/p/118381660
