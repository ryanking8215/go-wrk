package loader

import (
	"io"
	"net/http"
	"os"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
)

type ScriptContext struct {
	Config   Config
	request  func()
	response func(status int, header http.Header, body string)
	stop     func() bool
	delay    func() int // in ms
}

func loadFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func LoadScript(cfg Config, fn string) (*ScriptContext, error) {
	content, err := loadFile(fn)
	if err != nil {
		return nil, err
	}

	s := &ScriptContext{Config: cfg.Clone()}

	vm := goja.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)

	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	if err := vm.Set("wrk", &s.Config); err != nil {
		return nil, err
	}
	if _, err := vm.RunString(content); err != nil {
		return nil, err
	}

	if ret := vm.Get("request"); ret != nil {
		if err := vm.ExportTo(ret, &s.request); err != nil {
			return nil, err
		}
	}
	if ret := vm.Get("response"); ret != nil {
		if err := vm.ExportTo(ret, &s.response); err != nil {
			return nil, err
		}
	}
	if ret := vm.Get("stop"); ret != nil {
		if err := vm.ExportTo(ret, &s.stop); err != nil {
			return nil, err
		}
	}
	if ret := vm.Get("delay"); ret != nil {
		if err := vm.ExportTo(ret, &s.delay); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// func (s *ScriptContext) Clone() *ScriptContext {
// 	if s == nil {
// 		return nil
// 	}

// 	cloned := &ScriptContext{}
// 	cloned.Config = s.Config.Clone()
// 	cloned.request = s.request
// 	cloned.response = s.response
// 	cloned.stop = s.stop
// 	cloned.delay = s.delay
// 	return cloned
// }
