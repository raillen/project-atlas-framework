package projectatlas

import (
	"embed"
	"io/fs"
)

//go:embed schemas/*.json
var schemaAssets embed.FS

//go:embed src/project_atlas/resources/adapters/*.md
var adapterAssets embed.FS

//go:embed src/project_atlas/resources/catalog/*.json
var catalogAssets embed.FS

//go:embed src/project_atlas/resources/workforce
var workforceAssets embed.FS

func EmbeddedSchemas() fs.FS  { return subFS(schemaAssets, "schemas") }
func EmbeddedAdapters() fs.FS { return subFS(adapterAssets, "src/project_atlas/resources/adapters") }
func EmbeddedCatalog() fs.FS  { return subFS(catalogAssets, "src/project_atlas/resources/catalog") }
func EmbeddedWorkforce() fs.FS {
	return subFS(workforceAssets, "src/project_atlas/resources/workforce")
}

func subFS(source embed.FS, path string) fs.FS {
	sub, err := fs.Sub(source, path)
	if err != nil {
		return source
	}
	return sub
}
