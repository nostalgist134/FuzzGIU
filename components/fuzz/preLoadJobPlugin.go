package fuzz

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
)

var embeddedGen = map[string]bool{
	"int":       true,
	"permute":   true,
	"permuteex": true,
	"nil":       true,
}

var embeddedPlProc = map[string]bool{
	"urlencode":    true,
	"base64":       true,
	"addslashes":   true,
	"stripslashes": true,
	"suffix":       true,
	"repeat":       true,
}

var embeddedIterator = map[string]bool{
	"sniper":          true,
	"clusterbomb":     true,
	"pitchfork":       true,
	"pitchfork-cycle": true,
}

// checkPlugin 根据插件信息检查参数和插件类型是否有错（仅当插件信息不为空时检查）
func checkPlugin(pi *convention.PluginInfo, expectTypeInd int, p fuzzTypes.Plugin, argOffs int) error {
	if pi == nil {
		return nil
	}

	expectType := convention.PluginTypes[expectTypeInd]

	// 判断插件类型是否相符
	if pi.Type != expectType {
		return fmt.Errorf("incorrect plugin type %s, expect %s", pi.Type, expectType)
	}

	// 判断插件参数列表是否相符
	if len(pi.Params) != len(p.Args)+argOffs {
		return fmt.Errorf("incorrect argument count for %s, expect %d, got %d",
			expectType, len(pi.Params)-argOffs, len(p.Args))
	}
	return nil
}

// preLoadJobPlugin 预加载需要使用的插件
func preLoadJobPlugin(job *fuzzTypes.Fuzz) error {
	var errTotal error

	for _, plTmp := range job.Preprocess.PlMeta {
		// 加载payload生成器插件
		if plTmp.Generators.Type == "plugin" { // 仅当生成器类型为plugin时才加载插件
			for _, gen := range plTmp.Generators.Gen {
				// 避免将内置生成器当成插件加载
				if _, ok := embeddedGen[gen.Name]; ok {
					continue
				}
				pi, err := plugin.PreLoad(gen, plugin.RelPathPlGen)
				if err != nil {
					errTotal = errors.Join(errTotal, err)
				}

				if err = checkPlugin(pi, convention.IndPTypePlGen, gen, 0); err != nil {
					errTotal = errors.Join(errTotal, err)
				}
			}
		}

		// 加载payload处理器插件
		for _, plProc := range plTmp.Processors {
			// 避免将内置处理器当成插件加载
			if _, ok := embeddedPlProc[plProc.Name]; ok {
				continue
			}
			pi, err := plugin.PreLoad(plProc, plugin.RelPathPlProc)
			if err != nil {
				errTotal = errors.Join(errTotal, err)
			}

			if err = checkPlugin(pi, convention.IndPTypePlProc, plProc, 1); err != nil {
				errTotal = errors.Join(errTotal, err)
			}
		}
	}

	// 加载preprocessor插件
	for _, preproc := range job.Preprocess.Preprocessors {
		pi, err := plugin.PreLoad(preproc, plugin.RelPathPreprocessor)
		if err != nil {
			errTotal = errors.Join(errTotal, err)
		}

		if err = checkPlugin(pi, convention.IndPTypePreproc, preproc, 1); err != nil {
			errTotal = errors.Join(errTotal, err)
		}
	}

	// requester插件由于可能是易变的（url中可能包含fuzz关键字），预加载实现难度过大，因此略去

	// 加载reactor插件
	if job.React.Reactor.Name != "" {
		reactor := job.React.Reactor
		pi, err := plugin.PreLoad(reactor, plugin.RelPathReactor)
		if err != nil {
			errTotal = errors.Join(errTotal, err)
		}

		if err = checkPlugin(pi, convention.IndPTypeReact, reactor, 2); err != nil {
			errTotal = errors.Join(errTotal, err)
		}
	}

	// 加载iterator插件
	if iterName := job.Control.IterCtrl.Iterator.Name; iterName != "" {
		iterator := job.Control.IterCtrl.Iterator
		if _, ok := embeddedIterator[iterName]; !ok {
			pi, err := plugin.PreLoad(iterator, plugin.RelPathIterator)
			if err != nil {
				errTotal = errors.Join(errTotal, err)
			}

			if err = checkPlugin(pi, convention.IndPTypeIterator, iterator, 2); err != nil {
				errTotal = errors.Join(errTotal, err)
			}
		}
	}

	return errTotal
}
