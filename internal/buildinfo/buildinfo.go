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

type BuildInfoApp struct{}

func (b *BuildInfoApp) Version() string {
	return Version
}

func (b *BuildInfoApp) BuildTime() string {
	return BuildTime
}

func (b *BuildInfoApp) AppName() string {
	if AppName == "" {
		return "multitenant_alerts"
	}
	return AppName
}

func (b *BuildInfoApp) ModulePath() string {
	//return pkg_utils.ModulePath(b.ModulePath)
	return modulePath(b.ModulePath)
}

func modulePath(fn any) string {
	value := reflect.ValueOf(fn)
	ptr := value.Pointer()
	ffp := runtime.FuncForPC(ptr)
	modulePath := path.Dir(path.Dir(ffp.Name()))

	return modulePath
}

var BuildInfo = &BuildInfoApp{}
