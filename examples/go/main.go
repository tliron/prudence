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

	jsonRepresentation := &rest.Representation{
		Present: presentJson,
	}

	defaultRepresentation := &rest.Representation{
		Present: presentDefault,
	}

	main := rest.NewFacet("main", []string{"{name}"})
	main.Representations["application/json"] = jsonRepresentation
	main.Representations[""] = defaultRepresentation

	age := rest.NewFacet("age", []string{"{name}/age"})
	age.Representations["application/json"] = jsonRepresentation
	age.Representations[""] = defaultRepresentation

	person := rest.NewResource("person")
	person.AddFacet(main)
	person.AddFacet(age)

	router := rest.NewRouter("myapp")
	router.AddRoute(rest.NewRoute("", []string{"persons/*"}, person.Handle))
	router.AddRoute(rest.NewRoute("", nil, rest.DefaultNotFound.Handle))

	server := rest.NewServer("127.0.0.1:8080", router.Handle)

	err := server.Start()
	util.FailOnError(err)
}

func presentJson(context *rest.Context) error {
	person := map[string]interface{}{"name": context.Variables["name"]}
	bytes, _ := json.Marshal(person)
	context.Write(bytes)
	context.Write([]byte("\n"))
	return nil
}

func presentDefault(context *rest.Context) error {
	fmt.Fprintf(context, "%s\n", context.ContentType)
	fmt.Fprintf(context, "%s\n", context.Variables)
	return nil
}
