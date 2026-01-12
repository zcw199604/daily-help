package unraid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_RestartStopForceUpdate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		q := req.Query

		switch {
		case strings.Contains(q, "docker { containers"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"docker": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"id":     "docker:abc",
								"names":  []string{"app"},
								"state":  "running",
								"status": "Up",
							},
						},
					},
				},
			})
			return

		case strings.Contains(q, "mutation Stop"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"docker": map[string]interface{}{
						"stop": map[string]interface{}{
							"id":     "docker:abc",
							"state":  "exited",
							"status": "Exited",
						},
					},
				},
			})
			return

		case strings.Contains(q, "mutation Start"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"docker": map[string]interface{}{
						"start": map[string]interface{}{
							"id":     "docker:abc",
							"state":  "running",
							"status": "Up",
						},
					},
				},
			})
			return

		case strings.Contains(q, "__schema"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"mutationType": map[string]interface{}{
							"fields": []map[string]interface{}{
								{
									"name": "docker",
									"type": map[string]interface{}{
										"name": "DockerMutations",
										"kind": "OBJECT",
									},
								},
							},
						},
					},
				},
			})
			return

		case strings.Contains(q, "__type"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"__type": map[string]interface{}{
						"fields": []map[string]interface{}{
							{
								"name": "update",
								"args": []map[string]interface{}{
									{
										"name": "id",
										"type": map[string]interface{}{
											"kind": "NON_NULL",
											"ofType": map[string]interface{}{
												"kind": "SCALAR",
												"name": "PrefixedID",
											},
										},
									},
								},
								"type": map[string]interface{}{
									"kind": "OBJECT",
									"name": "DockerContainer",
								},
							},
						},
					},
				},
			})
			return

		case strings.Contains(q, "mutation ForceUpdate"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"docker": map[string]interface{}{
						"update": map[string]interface{}{
							"__typename": "DockerContainer",
						},
					},
				},
			})
			return

		default:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": "unexpected query"},
				},
			})
			return
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(ClientConfig{
		Endpoint: srv.URL,
		APIKey:   "k",
		Origin:   "o",
	}, srv.Client())

	ctx := context.Background()

	if err := c.RestartContainerByName(ctx, "app"); err != nil {
		t.Fatalf("RestartContainerByName() error: %v", err)
	}
	if err := c.StopContainerByName(ctx, "app"); err != nil {
		t.Fatalf("StopContainerByName() error: %v", err)
	}
	if err := c.ForceUpdateContainerByName(ctx, "app"); err != nil {
		t.Fatalf("ForceUpdateContainerByName() error: %v", err)
	}
}
