package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	gen "github.com/Neratus/geoguide/internal/delivery/gen/graphql"
	"github.com/Neratus/geoguide/internal/delivery/graphql/resolvers"
	"github.com/vektah/gqlparser/v2/ast"
)

func NewHandler(r *resolvers.Resolver) http.Handler {
	srv := handler.New(gen.NewExecutableSchema(gen.Config{Resolvers: r}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{Cache: lru.New[string](100)})

	return srv
}
