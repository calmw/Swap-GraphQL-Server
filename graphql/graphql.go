package graphql

import (
	"Swap-Server/models"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
)

func Handle1() *handler.Handler {
	// 定义一个简单的GraphQL对象类型
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Swap",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: graphql.String,
			},
			"from_token": &graphql.Field{
				Type: graphql.String,
			},
			"from_token_symbol": &graphql.Field{
				Type: graphql.String,
			},
			"from_amount": &graphql.Field{
				Type: graphql.String,
			},
			"to_token_symbol": &graphql.Field{
				Type: graphql.String,
			},
			"to_token": &graphql.Field{
				Type: graphql.String,
			},
			"to_amount": &graphql.Field{
				Type: graphql.String,
			},
			"tx_hash": &graphql.Field{
				Type: graphql.String,
			},
			"block_number": &graphql.Field{
				Type: graphql.String,
			},
			"create_time": &graphql.Field{
				Type: graphql.String,
			},
			"utc_date_time": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// 定义GraphQL查询类型
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"swaps": &graphql.Field{
				//Type: userType,
				Type: graphql.NewList(userType), // 返回数组
				Args: graphql.FieldConfigArgument{
					"user": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// 这里是获取用户的逻辑
					user := parseParameter(p.Args["user"])
					fromTokenSymbol := parseParameter(p.Args["from_token_symbol"])
					toTokenSymbol := parseParameter(p.Args["to_token_symbol"])
					//fmt.Println(user, fromTokenSymbol, toTokenSymbol)
					data, err := models.NewOrder().Query(user, fromTokenSymbol, toTokenSymbol)
					if err != nil {
						return "", err
					} else {
						return data, nil
					}
					// 这里是获取用户的逻辑
					//res := make([]map[string]interface{}, 0)
					//res = append(res, map[string]interface{}{
					//	"User":       user,
					//	"from_token": "John Doe",
					//	"to_token":   "john@example.com",
					//})
					//fmt.Println(res)
					//return res, nil
				},
			},
		},
	})

	// 定义GraphQL Schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		log.Fatalf("failed to create new schema: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	})

	return h
}

func parseParameter(p interface{}) string {
	if p == nil {
		return ""
	}
	return p.(string)
}
