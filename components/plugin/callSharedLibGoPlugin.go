//go:build darwin || linux

package plugin

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	goPlugin "plugin"
	"sync"
)

type symRecord struct {
	pluginFile *goPlugin.Plugin
	pluginFun  func(...any) uintptr
}

var symRecords = sync.Map{}
var mu = sync.Mutex{}

func callSharedLib(plugin fuzzTypes.Plugin, relPath string, jsons ...[]byte) uintptr {
	pName := relPath + plugin.Name
	var pw func(...any) uintptr
	// 尝试从缓存中加载插件函数
	if pRecord, ok := symRecords.Load(pName); ok {
		pw = pRecord.(symRecord).pluginFun
	} else { // 若失败则使用open打开
		mu.Lock()
		if pRecord, ok := symRecords.Load(pName); ok {
			pw = pRecord.(symRecord).pluginFun
		} else {
			p, err := goPlugin.Open(BaseDir + pName + binSuffix)
			if err != nil {
				return uintptr(0)
			}
			sym, err := p.Lookup(pluginEntry)
			if err != nil {
				return uintptr(0)
			}
			pw, ok = sym.(func(...any) uintptr)
			if !ok {
				return uintptr(0)
			}
			symRecords.Store(pName, symRecord{p, pw})
		}
		mu.Unlock()
	}
	args := make([]any, 0)
	if len(jsons) > 0 && jsons[0] != nil {
		args = append(args, jsons[0])
	}
	if len(jsons) > 1 && jsons[1] != nil {
		args = append(args, jsons[1])
	}
	args = append(args, plugin.Args...)
	return pw(args...)
}
