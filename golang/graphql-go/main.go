package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"../common"
	"errors"
	"github.com/graphql-go/graphql"
)

// TodoItemType ...
var TodoItemType = graphql.NewObject(graphql.ObjectConfig{
	Name: "TodoItem",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"done": &graphql.Field{
			Type: graphql.Boolean,
		},
		"listId": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

// TodoListType ...
var TodoListType = graphql.NewObject(graphql.ObjectConfig{
	Name: "TodoList",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var rootMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootMutation",
	Fields: graphql.Fields{
		"createTodoList": &graphql.Field{
			Type: TodoListType,
			Description: "Create a new todo list",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("could not find 'name' parameter")
				}
				id, ok := params.Args["id"].(int)
				if !ok {
					return nil, errors.New("could not found 'id' parameter")
				}
				list := common.TodoList{
					ID: id,
					Name: name,
				}
				err := common.AddList(list)
				if err != nil {
					return nil, err
				}
				return list, nil
			},
		},

		"createTodoItem": &graphql.Field{
			Type:        TodoItemType,
			Description: "Create new todo item",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"listId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				name, ok := params.Args["name"].(string)
				if !ok {
					return nil, errors.New("could not find 'name' parameter")
				}
				listID, ok := params.Args["listId"].(int)
				if !ok {
					return nil, errors.New("could not find 'listId' parameter")
				}
				newTodo := common.TodoItem{
					Name:   name,
					Done:   false,
					ListID: listID,
				}
				list, err := common.GetListByID(listID)
				if err != nil {
					return nil, err
				}
				err = list.AddItem(newTodo)
				if err != nil {
					return nil, err
				}

				return newTodo, nil
			},
		},

		"updateTodoList": &graphql.Field{
			Type: TodoListType,
			Description: "Update exisint todo list",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
		},

		"updateTodoItem": &graphql.Field{
			Type:        TodoItemType,
			Description: "Update existing todo item, mark it done or not done",
			Args: graphql.FieldConfigArgument{
				"done": &graphql.ArgumentConfig{
					Type: graphql.Boolean,
				},
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				done, _ := params.Args["done"].(bool)
				id, _ := params.Args["id"].(int)
				affectedTodo, err := common.GetItemByID(id)
				if err != nil {
					return nil, err
				}
				affectedTodo.Done = done
				list, err := common.GetListByID(affectedTodo.ListID)
				if err != nil {
					return nil, err
				}
				err = list.UpdateItem(affectedTodo)
				if err != nil {
					return nil, err
				}

				return affectedTodo, nil
			},
		},
	},
})

var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"todoItem": &graphql.Field{
			Type:        TodoItemType,
			Description: "Get single todo item",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				idQuery, isOK := params.Args["id"].(int)
				if isOK {
					// Search for el with id
					item, err := common.GetItemByID(idQuery)
					if err != nil {
						return nil, err
					}
					return item, nil
				}

				return common.TodoItem{}, nil
			},
		},

		"todoList": &graphql.Field{
			Type:        TodoListType,
			Description: "Get a todo list",
			Args: graphql.FieldConfigArgument{
				"listId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				listID, ok := p.Args["listId"].(int)
				if !ok {
					return nil, errors.New("no 'listId' parameter")
				}
				list, err := common.GetListByID(listID)
				if err != nil {
					return nil, err
				}
				return list, nil
			},
		},

		"todoListItems": &graphql.Field{
			Type: graphql.NewList(TodoItemType),
			Description: "Get the items in a todo list",
			Args: graphql.FieldConfigArgument{
				"listId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				listID, ok := p.Args["listId"].(int)
				if !ok {
					return nil, errors.New("no 'listId' parameter")
				}
				list, err := common.GetListByID(listID)
				if err != nil {
					return nil, err
				}
				items, err := list.GetItems()
				if err != nil {
					return nil, err
				}
				return items, nil
			},
		},
	},
})

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})

	if len(result.Errors) > 0 {
		log.Printf("wrong result, unexpected errors: %+v", result.Errors)
	}
	return result
}

func main() {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		log.Fatalf("error initialising schema, error: %v", err)
	}
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, "invalid query", http.StatusBadRequest)
		}
	})

	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
