package util

// DiffIDs 对比旧ID切片和新ID切片，返回新增和删除的ID
func DiffIDs(oldIDs, newIDs []uint32) (added, removed []uint32) {
	// 旧ID映射表（用于快速查找）
	oldMap := make(map[uint32]struct{}, len(oldIDs))
	for _, id := range oldIDs {
		oldMap[id] = struct{}{}
	}

	// 新ID映射表
	newMap := make(map[uint32]struct{}, len(newIDs))
	for _, id := range newIDs {
		newMap[id] = struct{}{}
	}

	// 新增：存在于新ID但不存在于旧ID
	for id := range newMap {
		if _, exists := oldMap[id]; !exists {
			added = append(added, id)
		}
	}

	// 删除：存在于旧ID但不存在于新ID
	for id := range oldMap {
		if _, exists := newMap[id]; !exists {
			removed = append(removed, id)
		}
	}

	return added, removed
}
