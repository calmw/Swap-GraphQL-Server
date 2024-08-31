package graphql

import (
	"Swap-Server/models"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
	"strings"
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
				Description: "get user swap records list",
				Type:        graphql.NewList(userType), // 返回数组
				Args: graphql.FieldConfigArgument{
					"user": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"from_token_symbol": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"to_token_symbol": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"order_by": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					user := parseParameterWithLowerCase(p.Args["user"])
					fromTokenSymbol := parseParameter(p.Args["from_token_symbol"])
					toTokenSymbol := parseParameter(p.Args["to_token_symbol"])
					fromTokenAddress := parseParameterWithLowerCase(p.Args["from_token_address"])
					toTokenAddress := parseParameterWithLowerCase(p.Args["to_token_address"])
					orderBy := parseParameter(p.Args["order_by"])
					fmt.Println(fmt.Sprintf("user:%s, fromTokenSymbol:%s, toTokenSymbol:%s, fromTokenAddress:%s, toTokenAddress:%s, orderBy:%s",
						user, fromTokenSymbol, toTokenSymbol, fromTokenAddress, toTokenAddress, orderBy,
					))
					data, err := models.NewOrder().Query(user, fromTokenSymbol, toTokenSymbol, fromTokenAddress, toTokenAddress, orderBy)
					if err != nil {
						return "", err
					} else {
						return data, nil
					}
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

func parseParameterWithLowerCase(p interface{}) string {
	if p == nil {
		return ""
	}
	return strings.ToLower(p.(string))
}
func parseParameter(p interface{}) string {
	if p == nil {
		return ""
	}
	return p.(string)
}
