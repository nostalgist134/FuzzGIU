package input

func HandleInput() bool {
	globCq.mu.Lock()
	defer globCq.mu.Unlock()
	if globCq.listCursor != -1 {
		return true
	}
	return false
}
