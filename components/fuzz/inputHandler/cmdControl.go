package inputHandler

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/rp"
	"strconv"
	"strings"
)

var ErrMissingControl = errors.New("missing poolCtrl operation")
var ErrMissingSize = errors.New("missing size to resize the routine pool")

func controlPause() {
	rp.CurrentRp.Pause()
}

func controlResume() {
	rp.CurrentRp.Resume()
}

func controlResize(size int) error {
	if size > 0 {
		rp.CurrentRp.Resize(size)
		curFuz := fuzzCommon.GetCurFuzz()
		curFuz.Misc.PoolSize = size
		output.UpdateScreenInfoPage(curFuz)
		return nil
	}
	return fmt.Errorf("invalid size %d", size)
}

func poolCtrl(args []string, _ []byte) (any, error) {
	if len(args) == 0 {
		return nil, ErrMissingControl
	}
	switch strings.ToLower(args[0]) {
	case "pause":
		controlPause()
	case "resume":
		controlResume()
	case "resize":
		if len(args) == 1 {
			return nil, ErrMissingSize
		}
		size, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}
		if err = controlResize(size); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown operation over rp: %s", args[0])
	}
	return bytesOk, nil
}
