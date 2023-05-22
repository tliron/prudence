package main

import (
	"encoding/json"
	"fmt"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	"github.com/tliron/prudence/rest"

	_ "github.com/tliron/commonlog/simple"
)

func main() {
	util.ExitOnSIGTERM()

	commonlog.Configure(2, nil)
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
	main.Representations.Add([]string{"application/json"}, nil, jsonRepresentation)
	main.Representations.Add(nil, nil, defaultRepresentation)

	age := rest.NewFacet("age")
	age.PathTemplates, err = rest.NewPathTemplates("{name}/age")
	util.FailOnError(err)
	age.Representations.Add([]string{"application/json"}, nil, jsonRepresentation)
	age.Representations.Add(nil, nil, defaultRepresentation)

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
	server.Address = ":8080"
	server.Handler = router.Handle
	err = server.Start()
	util.FailOnError(err)

	util.Exit(0)
}

// rest.RepresentationFunc signature
func presentJson(context *rest.Context) error {
	person := map[string]interface{}{"name": context.Variables["name"]}
	bytes, _ := json.Marshal(person)
	context.Write(bytes)
	context.WriteString("\n")
	return nil
}

// rest.RepresentationFunc signature
func presentDefault(context *rest.Context) error {
	context.WriteString(context.Response.ContentType)
	fmt.Fprintf(context, "\n%s\n", context.Variables)
	return nil
}
