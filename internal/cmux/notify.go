package cmux

// CreateNotification triggers a cmux notification popover.
func CreateNotification(title, subtitle, body string) error {
	return Call("notification.create", map[string]string{
		"title":    title,
		"subtitle": subtitle,
		"body":     body,
	}, nil)
}
