package input

func (stk *inputStack) init(size int) {
	stk.list = make([]*Input, size)
	stk.cursor = -1
}

func (stk *inputStack) push(c *Input) {
	stk.mu.Lock()
	defer stk.mu.Unlock()
	if stk.cursor == len(stk.list)-1 {
		stk.list = append(stk.list, c)
	} else {
		stk.list[stk.cursor+1] = c
	}
	stk.cursor++
}

func (stk *inputStack) pop() *Input {
	stk.mu.Lock()
	defer stk.mu.Unlock()
	if stk.cursor == -1 {
		return nil
	}
	stk.cursor--
	return stk.list[stk.cursor+1]
}
