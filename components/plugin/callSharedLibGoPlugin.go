//go:build darwin || linux

package plugin

import (
	goPlugin "plugin"
)

func CallSharedLib(plugin Plugin, relPath string, jsons ...[]byte) uintptr {
	pName := relPath + plugin.Name
	p, err := goPlugin.Open(BaseDir + pName + binSuffix)
	if err != nil {
		return uintptr(0)
	}
	sym, err := p.Lookup(pluginEntry)
	if err != nil {
		return uintptr(0)
	}
	pw, ok := sym.(func(...any) uintptr)
	if !ok {
		return uintptr(0)
	}
	args := make([]any, 0)
	if len(jsons) > 0 && jsons[0] != nil {
		args = append(args, jsons[0])
		if jsons[1] != nil {
			args = append(args, jsons[1])
		}
	}
	args = append(args, plugin.Args...)
	return pw(args...)
}
