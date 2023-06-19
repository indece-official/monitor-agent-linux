package agent

func (c *Controller) boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}
