package js

//
// ConsoleAPI
//

type ConsoleAPI struct{}

func (self ConsoleAPI) Log(message string) {
	log.Info(message)
}
