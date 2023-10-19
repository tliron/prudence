package main

import (
	"fmt"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/rest"

	_ "github.com/tliron/commonlog/simple"
)

func main() {
	util.ExitOnSignals()
	util.InitializeColorization("true")
	commonlog.Configure(2, nil)
	start()
	util.Exit(0)
}

func start() {
	var err error

	// Server

	server := rest.NewServer("myserver")
	server.Port = 8080

	// Router

	router := rest.NewRouter("myapp")
	server.Handler = router.Handle
	router.Variables["app"] = "myapp"

	// Resource

	person := rest.NewResource("person")
	route := rest.NewRoute("")
	router.AddRoute(route)
	route.PathTemplates, err = rest.NewPathTemplates("person//*")
	util.FailOnError(err)
	route.Variables["resource"] = "person"
	route.Handler = person.Handle

	// NotFound

	route = rest.NewRoute("notFound")
	router.AddRoute(route)
	route.Handler = rest.HandleNotFound

	// Facets

	main := rest.NewFacet("main")
	person.AddFacet(main)
	main.PathTemplates, err = rest.NewPathTemplates("{name}//")
	util.FailOnError(err)

	age := rest.NewFacet("age")
	person.AddFacet(age)
	age.PathTemplates, err = rest.NewPathTemplates("{name}/age//")
	util.FailOnError(err)

	// Representations

	textRepresentation := rest.NewRepresentation("text")
	textRepresentation.Present = presentText
	main.Representations.Add([]string{"text/plain"}, []string{"en"}, textRepresentation)
	age.Representations.Add([]string{"text/plain"}, []string{"en"}, textRepresentation)

	dataRepresentation := rest.NewRepresentation("json")
	dataRepresentation.Present = presentData
	main.Representations.Add(rest.DataContentTypes, []string{"en"}, dataRepresentation)
	age.Representations.Add(rest.DataContentTypes, []string{"en"}, dataRepresentation)

	// Start!

	util.FailOnError(server.Start())
}

// ([rest.RepresentationHook] signature)
func presentText(restContext *rest.Context) error {
	fmt.Fprintf(restContext.Writer, "%s\n", restContext.Variables)
	return nil
}

// ([rest.RepresentationHook] signature)
func presentData(restContext *rest.Context) error {
	person := map[string]any{"name": restContext.Variables["name"]}
	restContext.Transcribe(person, "", "  ")
	return nil
}
