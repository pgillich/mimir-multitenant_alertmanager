package buildinfo

import (
	"path"
	"reflect"
	"runtime"
)

// Version is set by the linker.
//
//nolint:gochecknoglobals // set by the linker
var Version string

// BuildTime is set by the linker.
//
//nolint:gochecknoglobals // set by the linker
var BuildTime string

// AppName is set by the linker.
//
//nolint:gochecknoglobals // set by the linker
var AppName string

func GetAppName() string {
	if AppName == "" {
		return "multitenant_alertmanager"
	}
	return AppName
}

func ModulePath() string {
	value := reflect.ValueOf(ModulePath)
	ptr := value.Pointer()
	ffp := runtime.FuncForPC(ptr)
	modulePath := path.Dir(path.Dir(ffp.Name()))

	return modulePath
}
