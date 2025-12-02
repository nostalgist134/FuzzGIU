package tviewOutput

func (s *tviewScreen) addListJobItem(name, secondaryText string, callback func()) {
	// 先 AddItem（必须在 QueueUpdate 内，但这里假设调用者已处理）
	index := s.listJobs.GetItemCount()
	s.listJobs.AddItem(name, secondaryText, 0, callback)

	// 再更新 map
	s.listJobsNameToIndex.Store(name, index)
}

func (s *tviewScreen) removeListJobItemByName(name string) {
	// 注意：RemoveItem 会导致后续索引前移，所以不能只删一个！
	if _, ok := s.listJobsNameToIndex.Load(name); !ok {
		return
	}
	// 执行删除
	s.listJobs.RemoveItem(s.getListItemIndexByName(name))
	// 重建整个map
	s.rebuildNameToIndexMap()
}

func (s *tviewScreen) rebuildNameToIndexMap() {
	// 清空
	oldMap := make(map[string]int)
	s.listJobsNameToIndex.Range(func(key, _ interface{}) bool {
		oldMap[key.(string)] = 0 // 标记所有旧 key
		return true
	})
	// 删除所有旧 key
	for k := range oldMap {
		s.listJobsNameToIndex.Delete(k)
	}
	// 重新填充
	for i := 0; i < s.listJobs.GetItemCount(); i++ {
		name, _ := s.listJobs.GetItemText(i)
		s.listJobsNameToIndex.Store(name, i)
	}
}

func (s *tviewScreen) getListItemIndexByName(name string) int {
	if val, ok := s.listJobsNameToIndex.Load(name); ok {
		return val.(int)
	}
	return -1
}

func (s *tviewScreen) updateItemName(old, new string) {
	if ind := s.getListItemIndexByName(old); ind != -1 {
		_, secondary := screen.listJobs.GetItemText(ind)
		screen.listJobs.SetItemText(ind, new, secondary)
		s.listJobsNameToIndex.Store(new, ind)
	}
}
