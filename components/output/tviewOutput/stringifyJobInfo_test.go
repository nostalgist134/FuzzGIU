package tviewOutput

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	"testing"
)

func TestStringify(t *testing.T) {
	f := convention.GetFullStruct("*fuzzTypes.Fuzz").(*fuzzTypes.Fuzz)
	f.Control.OutSetting.ToWhere |= outputFlag.OutToHttp | outputFlag.OutToTview
	f.Preprocess.Preprocessors = []fuzzTypes.Plugin{{"NISHIGIU", []any{3, 4, "woshigiu"}}, {"WOSHIGIU", []any{"", 1}}}
	fmt.Println(stringifyJobInfo(f))
}

func TestGetColorByType(t *testing.T) {
	fmt.Println(getColorByType(3))
}
