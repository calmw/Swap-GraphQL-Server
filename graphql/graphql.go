package graphql

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
	"time"
)

type Feed struct {
	ID string `graphql:"id"`
}

var FeedType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FeedType",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
	},
})

var RootSubscription = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootSubscription",
	Fields: graphql.Fields{
		"feed": &graphql.Field{
			Type: FeedType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				c := make(chan interface{})

				go func() {
					var i int

					for {
						i++

						feed := Feed{ID: fmt.Sprintf("%d", i)}

						select {
						case <-p.Context.Done():
							log.Println("[RootSubscription] [Subscribe] subscription canceled")
							close(c)
							return
						default:
							c <- feed
						}

						time.Sleep(250 * time.Millisecond)

						if i == 21 {
							close(c)
							return
						}
					}
				}()

				return c, nil
			},
		},
	},
})

var RootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"ping": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				fmt.Println(p.Source, 111)
				return struct {
					Name string `json:"name"`
					Age  string `json:"age"`
				}{"ss", "18"}, nil
			},
		},
		"ping2": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {

				fmt.Println(p.Source, 222)

				return "ok2", nil
			},
		},
	},
})

var schema graphql.Schema

func Handle1() *handler.Handler {
	schemaConfig := graphql.SchemaConfig{
		Query:        RootQuery,
		Subscription: RootSubscription,
	}

	s, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatal(err)
	}

	//schema = s

	h := handler.New(&handler.Config{
		Schema:     &s,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	return h

}
