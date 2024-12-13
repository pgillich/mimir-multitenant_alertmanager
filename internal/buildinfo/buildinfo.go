package buildinfo

import (
	"os"

	"golang.org/x/mod/modfile"
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
	var err error
	var data []byte
	files := []string{"go.mod", "../go.mod", "../../go.mod"}
	for _, file := range files {
		data, err = os.ReadFile(file)
		if err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}
	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		panic(err)
	}
	return modFile.Module.Mod.Path
}
