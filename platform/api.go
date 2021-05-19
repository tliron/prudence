package platform

var apis = make(map[string]interface{})

func RegisterAPI(name string, api interface{}) {
	apis[name] = api
}

func OnAPIs(f func(name string, api interface{}) bool) {
	for name, api := range apis {
		if !f(name, api) {
			return
		}
	}
}
