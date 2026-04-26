package cmux

// OpenBrowser asks cmux to open a URL via the browser.open RPC.
func OpenBrowser(url string) error {
	return Call("browser.open", map[string]string{"url": url}, nil)
}
