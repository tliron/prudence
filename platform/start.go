package platform

type Startable interface {
	Start() error
	Stop() error
}

func Start(startables []Startable) error {
	for _, startable := range startables {
		startable_ := startable // closure capture
		go func() {
			if err := startable_.Start(); err != nil {
				log.Errorf("%s", err.Error())
			}
		}()
	}

	// Block forever
	<-make(chan struct{})

	return nil
}
