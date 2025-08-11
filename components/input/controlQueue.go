package input

func (cq *controlQueue) append(c Input) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	if cq.listCursor == len(cq.list)-1 {
		cq.list = append(cq.list, c)
	} else {
		cq.list[cq.listCursor] = c
	}
	cq.listCursor++
}
