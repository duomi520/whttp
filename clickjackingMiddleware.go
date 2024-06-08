package whttp

// ClickjackingMiddleware 点击劫持 是指攻击者使用多个透明或不透明层来诱使用户在打算点击顶层页面时点击另一个页面上的按钮或链接。
// 因此攻击者正在劫持针对其页面的点击，并将它们路由到另一个页面，该页面很可能是另一个应用程序或域。
func ClickjackingMiddleware() func(*HTTPContext) {
	return func(c *HTTPContext) {
		// 使用内容安全策略（CSP）frame-ancestors 指令进行防御
		// 此设置阻止任何域使用框架对页面进行引用
		c.Writer.Header().Set("frame-ancestors", "none")
		// 使用X-Frame-Options HTTP 响应标头进行防御
		// 此设置阻止任何域使用框架对页面进行引用
		c.Writer.Header().Set("X-Frame-Optoins", "DENY")
		c.Next()
	}
}

// https://www.cnblogs.com/xiaodi-js/p/16718859.html
