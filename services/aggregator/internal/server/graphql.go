// services/aggregator/internal/server/graphql.go
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"

	"github.com/system-highload-architect/qos-pipeline/services/aggregator/internal/adapters/postgres"
)

// GraphQLHandler обрабатывает GraphQL-запросы.
type GraphQLHandler struct {
	schema graphql.Schema
	store  *postgres.AggregateStore
}

// NewGraphQLHandler создаёт новый обработчик GraphQL.
func NewGraphQLHandler(store *postgres.AggregateStore) (*GraphQLHandler, error) {
	h := &GraphQLHandler{store: store}
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: h.buildQuery(),
	})
	if err != nil {
		return nil, fmt.Errorf("graphql schema: %w", err)
	}
	h.schema = schema
	return h, nil
}

// buildQuery определяет корневые поля запросов.
func (h *GraphQLHandler) buildQuery() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"metrics": &graphql.Field{
				Type: graphql.NewList(metricType),
				Args: graphql.FieldConfigArgument{
					"source": &graphql.ArgumentConfig{Type: graphql.String},
					"name":   &graphql.ArgumentConfig{Type: graphql.String},
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: h.resolveMetrics,
			},
			"sloStatus": &graphql.Field{
				Type: sloStatusType,
				Args: graphql.FieldConfigArgument{
					"source": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: h.resolveSLOStatus,
			},
			"sloHistory": &graphql.Field{
				Type: graphql.NewList(sloHistoryType),
				Args: graphql.FieldConfigArgument{
					"source": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"from":   &graphql.ArgumentConfig{Type: graphql.DateTime},
					"to":     &graphql.ArgumentConfig{Type: graphql.DateTime},
				},
				Resolve: h.resolveSLOHistory,
			},
		},
	})
}

// ---------- Типы GraphQL ----------

var metricType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Metric",
	Fields: graphql.Fields{
		"id":        &graphql.Field{Type: graphql.Int},
		"source":    &graphql.Field{Type: graphql.String},
		"name":      &graphql.Field{Type: graphql.String},
		"sum":       &graphql.Field{Type: graphql.Float},
		"count":     &graphql.Field{Type: graphql.Int},
		"min":       &graphql.Field{Type: graphql.Float},
		"max":       &graphql.Field{Type: graphql.Float},
		"p50":       &graphql.Field{Type: graphql.Float},
		"p95":       &graphql.Field{Type: graphql.Float},
		"p99":       &graphql.Field{Type: graphql.Float},
		"createdAt": &graphql.Field{Type: graphql.DateTime},
	},
})

var sloStatusType = graphql.NewObject(graphql.ObjectConfig{
	Name: "SLOStatus",
	Fields: graphql.Fields{
		"source":    &graphql.Field{Type: graphql.String},
		"available": &graphql.Field{Type: graphql.Float},
		"p95":       &graphql.Field{Type: graphql.Float},
		"p99":       &graphql.Field{Type: graphql.Float},
		"status":    &graphql.Field{Type: graphql.String},
	},
})

var sloHistoryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "SLOHistory",
	Fields: graphql.Fields{
		"source":    &graphql.Field{Type: graphql.String},
		"date":      &graphql.Field{Type: graphql.DateTime},
		"available": &graphql.Field{Type: graphql.Float},
		"p95":       &graphql.Field{Type: graphql.Float},
		"p99":       &graphql.Field{Type: graphql.Float},
		"status":    &graphql.Field{Type: graphql.String},
	},
})

// ---------- Резолверы ----------

func (h *GraphQLHandler) resolveMetrics(p graphql.ResolveParams) (interface{}, error) {
	source, _ := p.Args["source"].(string)
	name, _ := p.Args["name"].(string)
	limit, _ := p.Args["limit"].(int)
	if limit <= 0 {
		limit = 100
	}

	// Демо-заглушка
	return []map[string]interface{}{
		{"id": 1, "source": source, "name": name, "sum": 100.0, "count": 10, "p95": 10.5},
	}, nil
}

func (h *GraphQLHandler) resolveSLOStatus(p graphql.ResolveParams) (interface{}, error) {
	source := p.Args["source"].(string)
	return map[string]interface{}{
		"source":    source,
		"available": 99.95,
		"p95":       145.2,
		"p99":       480.0,
		"status":    "ok",
	}, nil
}

func (h *GraphQLHandler) resolveSLOHistory(p graphql.ResolveParams) (interface{}, error) {
	source := p.Args["source"].(string)
	now := time.Now()
	var history []map[string]interface{}
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Truncate(24 * time.Hour)
		history = append(history, map[string]interface{}{
			"source":    source,
			"date":      date,
			"available": 99.9 + float64(i)*0.01,
			"p95":       140.0 + float64(i)*2,
			"p99":       460.0 + float64(i)*5,
			"status":    "ok",
		})
	}
	return history, nil
}

// ServeHTTP обрабатывает GraphQL-запросы по HTTP.
func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result := graphql.Do(graphql.Params{
		Schema:        h.schema,
		RequestString: req.Query,
		Context:       r.Context(),
	})
	json.NewEncoder(w).Encode(result)
}
