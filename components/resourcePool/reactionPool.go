package resourcePool

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
)

var reactionPool = sync.Pool{
	New: func() any { return new(fuzzTypes.Reaction) },
}

// GetReaction 从池中获取一个新的Reaction结构
func GetReaction() *fuzzTypes.Reaction {
	newReaction := (reactionPool.Get()).(*fuzzTypes.Reaction)
	*newReaction = fuzzTypes.Reaction{}
	return newReaction
}

// PutReaction 将用完的Reaction结构放回池
func PutReaction(r *fuzzTypes.Reaction) {
	if r == nil {
		return
	}
	*r = fuzzTypes.Reaction{}
	reactionPool.Put(r)
}
