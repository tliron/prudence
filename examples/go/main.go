package main

import (
	"encoding/json"
	"fmt"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/rest"

	_ "github.com/tliron/kutil/logging/simple"
)

func main() {
	logging.Configure(2, nil)
	var err error

	jsonRepresentation := &rest.Representation{
		Present: presentJson,
	}

	defaultRepresentation := &rest.Representation{
		Present: presentDefault,
	}

	main := rest.NewFacet("main")
	main.PathTemplates, err = rest.NewPathTemplates("{name}")
	util.FailOnError(err)
	main.Representations["application/json"] = jsonRepresentation
	main.Representations[""] = defaultRepresentation

	age := rest.NewFacet("age")
	age.PathTemplates, err = rest.NewPathTemplates("{name}/age")
	util.FailOnError(err)
	age.Representations["application/json"] = jsonRepresentation
	age.Representations[""] = defaultRepresentation

	person := rest.NewResource("person")
	person.AddFacet(main)
	person.AddFacet(age)

	router := rest.NewRouter("myapp")
	route := rest.NewRoute("")
	route.PathTemplates, err = rest.NewPathTemplates("person/*")
	util.FailOnError(err)
	route.Handler = person.Handle
	router.AddRoute(route)
	route = rest.NewRoute("")
	route.Handler = rest.DefaultNotFound.Handle
	router.AddRoute(route)

	server := rest.NewServer("")
	server.Address = "localhost:8080"
	server.Handler = router.Handle

	err = server.Start()
	util.FailOnError(err)
}

func presentJson(context *rest.Context) error {
	person := map[string]interface{}{"name": context.Variables["name"]}
	bytes, _ := json.Marshal(person)
	context.Write(bytes)
	context.WriteString("\n")
	return nil
}

func presentDefault(context *rest.Context) error {
	context.WriteString(context.ContentType)
	fmt.Fprintf(context, "\n%s\n", context.Variables)
	return nil
}
