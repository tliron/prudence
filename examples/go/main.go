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

	main := rest.NewFacet("main", []string{"{name}"})
	main.SetPresenter("application/json", representJson)
	main.SetPresenter("", representDefault)

	age := rest.NewFacet("age", []string{"{name}/age"})
	main.SetPresenter("application/json", representJson)
	main.SetPresenter("", representDefault)

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

func representJson(context *rest.Context) error {
	person := map[string]string{"name": context.Variables["name"]}
	bytes, _ := json.Marshal(person)
	context.Write(bytes)
	context.Write([]byte("\n"))
	return nil
}

func representDefault(context *rest.Context) error {
	fmt.Fprintf(context, "%s\n", context.ContentType)
	fmt.Fprintf(context, "%s\n", context.Variables)
	return nil
}
