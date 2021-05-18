package render

/*
func init() {
	Register("typescript", RenderTypeScript)
	Register("ts", RenderTypeScript)
	Register("tsx", RenderTSX)
	Register("jsx", RenderJSX)
}

// RenderFunc signature
func RenderTypeScript(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	result := api.Transform(content, api.TransformOptions{
		Loader: api.LoaderTS,
		Target: api.ES2015,
	})

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("%v", result.Errors)
	} else {
		return util.BytesToString(result.Code), nil
	}
}

func RenderTSX(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	result := api.Transform(content, api.TransformOptions{
		Loader: api.LoaderTSX,
		Target: api.ES2015,
	})

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("%v", result.Errors)
	} else {
		return util.BytesToString(result.Code), nil
	}
}

func RenderJSX(content string, getRelativeURL common.GetRelativeURL) (string, error) {
	api.Build(api.BuildOptions{
		EntryPoints: []string{"examples/hello-world/myapp/person/profile/html.jsx"},
		Outfile:     "examples/hello-world/myapp/person/profile/html.js",
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		NodePaths:   []string{"npm/node_modules"},
		Target:      api.ES2015,
	})

	r, err := os.ReadFile("examples/hello-world/myapp/person/profile/html.js")
	return string(r), err

	result := api.Transform(content, api.TransformOptions{
		Loader: api.LoaderJSX,
		Target: api.ES2015,
	})

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("%v", result.Errors)
	} else {
		return util.BytesToString(result.Code), nil
	}
}
*/
