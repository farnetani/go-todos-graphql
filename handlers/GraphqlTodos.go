package handlers

import (
	"bytes"
	"encoding/json"
	"go-todos/database"
	"go-todos/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

func GraphqlTodos(db database.TodoInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := make(map[string]interface{})

		c.BindJSON(&body)
		requestString := body["query"].(string)

		rootQuery := graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"searchTodos": searchTodos(db),
				"getTodo":     getTodo(db),
			},
		})

		rootMutation := graphql.NewObject(graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"insertTodo": insertTodo(db),
				"updateTodo": updateTodo(db),
				"deleteTodo": deleteTodo(db),
			},
		})

		schema, _ := graphql.NewSchema(graphql.SchemaConfig{
			Query:    rootQuery,
			Mutation: rootMutation,
		})

		res := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: requestString,
		})

		c.JSON(http.StatusOK, res)
	}
}

var todoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "TodoType",
	Fields: graphql.Fields{
		"id":        &graphql.Field{Type: graphql.String},
		"userId":    &graphql.Field{Type: graphql.Int},
		"title":     &graphql.Field{Type: graphql.String},
		"completed": &graphql.Field{Type: graphql.Boolean},
	},
})

var updateTodoType = graphql.NewObject(graphql.ObjectConfig{
	Name:   "UpdateTodoType",
	Fields: graphql.Fields{"modifiedCount": &graphql.Field{Type: graphql.Int}, "result": &graphql.Field{Type: graphql.Int}},
})

var deleteTodoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteTodoType",
	Fields: graphql.Fields{
		"deletedCount": &graphql.Field{Type: graphql.Int},
	},
})

func insertTodo(db database.TodoInterface) *graphql.Field {
	args := graphql.FieldConfigArgument{
		"userId": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"title": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"completed": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.Boolean),
		},
	}

	return &graphql.Field{
		Name:        "insertTodo",
		Description: "Insert todo item",
		Type:        todoType,
		Args:        args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			todo := models.Todo{}
			args := p.Args
			b := new(bytes.Buffer)
			// err := json.NewDecoder(p.Args).Decode(&todo)
			// Fonte: https://stackoverflow.com/questions/66596555/how-to-convert-mapstringinterface-to-io-reader/66606719#66606719
			json.NewEncoder(b).Encode(args)
			err := json.NewDecoder(b).Decode(&todo)
			if err != nil {
				return "", err
			}
			return db.Insert(todo)
		},
	}
}

func updateTodo(db database.TodoInterface) *graphql.Field {
	args := graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"userId": &graphql.ArgumentConfig{
			Type: graphql.Int,
		},
		"title": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"completed": &graphql.ArgumentConfig{
			Type: graphql.Boolean,
		},
	}
	return &graphql.Field{
		Name:        "updateTodo",
		Description: "update todo by id",
		Args:        args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			args := p.Args
			id := args["id"].(string)
			delete(args, "id") // deleta o id dos argumentos por ser update
			return db.Update(id, args)
		},
	}
}

func deleteTodo(db database.TodoInterface) *graphql.Field {
	args := graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	}
	return &graphql.Field{
		Name:        "deleteTodo",
		Description: "delete todo by id",
		Args:        args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			return db.Delete(id)
		},
	}
}

func searchTodos(db database.TodoInterface) *graphql.Field {
	args := graphql.FieldConfigArgument{
		"title":     &graphql.ArgumentConfig{Type: graphql.String},
		"completed": &graphql.ArgumentConfig{Type: graphql.Boolean},
		"userId":    &graphql.ArgumentConfig{Type: graphql.Int},
		"id":        &graphql.ArgumentConfig{Type: graphql.String},
	}
	return &graphql.Field{
		Name:        "todos",
		Description: "List of todos",
		Type:        graphql.NewList(todoType),
		Args:        args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			filter := p.Args
			return db.Search(filter)
		},
	}
}

func getTodo(db database.TodoInterface) *graphql.Field {
	args := graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{Type: graphql.String},
	}
	return &graphql.Field{
		Name:        "todo",
		Description: "Get todo by id",
		Type:        todoType,
		Args:        args,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			return db.Get(id)
		},
	}
}
