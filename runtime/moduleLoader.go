package runtime

import (
	"os"
	"path/filepath"

	"github.com/Seeingu/coldmoon/coldmoon"
)

// FilesystemModuleLoader resolves and reads ECMAScript modules from disk.
type FilesystemModuleLoader struct{}

// RegisterFilesystemModuleLoader installs the terminal host's filesystem module
// resolver on realm's Agent.
func RegisterFilesystemModuleLoader(realm *coldmoon.Realm) {
	realm.Agent.HostHooks.HostLoadModule = FilesystemModuleLoader{}
}

// Resolve resolves specifier relative to hostDefined.BaseDir and returns an
// absolute, cleaned path as the canonical identity.
func (FilesystemModuleLoader) Resolve(
	_ coldmoon.ImportedModuleReferrer,
	specifier string,
	hostDefined coldmoon.HostDefined,
) (coldmoon.ModuleResolution, error) {
	filePath := specifier
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(hostDefined.BaseDir, filePath)
	}
	identity, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return coldmoon.ModuleResolution{}, err
	}
	return coldmoon.ModuleResolution{
		Identity: identity,
		HostDefined: coldmoon.HostDefined{
			FileName: filepath.Base(identity),
			BaseDir:  filepath.Dir(identity),
		},
	}, nil
}

// Load reads the source for a previously resolved filesystem identity.
func (FilesystemModuleLoader) Load(resolution coldmoon.ModuleResolution) (string, error) {
	sourceText, err := os.ReadFile(resolution.Identity)
	if err != nil {
		return "", err
	}
	return string(sourceText), nil
}
