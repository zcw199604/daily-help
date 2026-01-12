package unraid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type ClientConfig struct {
	Endpoint            string
	APIKey              string
	Origin              string
	ForceUpdateMutation string
}

type Client struct {
	cfg        ClientConfig
	httpClient *http.Client
}

func NewClient(cfg ClientConfig, httpClient *http.Client) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) RestartContainerByName(ctx context.Context, name string) error {
	id, err := c.findContainerIDByName(ctx, name)
	if err != nil {
		return err
	}
	if err := c.stopContainer(ctx, id); err != nil {
		if !errors.Is(err, ErrAlreadyStopped) {
			return err
		}
	}
	if err := c.startContainer(ctx, id); err != nil {
		if !errors.Is(err, ErrAlreadyStarted) {
			return err
		}
	}
	return nil
}

func (c *Client) StopContainerByName(ctx context.Context, name string) error {
	id, err := c.findContainerIDByName(ctx, name)
	if err != nil {
		return err
	}
	if err := c.stopContainer(ctx, id); err != nil {
		if errors.Is(err, ErrAlreadyStopped) {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) ForceUpdateContainerByName(ctx context.Context, name string) error {
	id, err := c.findContainerIDByName(ctx, name)
	if err != nil {
		return err
	}

	meta, supported, err := c.detectDockerForceUpdateMutation(ctx)
	if err != nil {
		return err
	}
	if !supported {
		return fmt.Errorf("当前 Unraid GraphQL API 未发现可用的“强制更新”mutation（可升级 Unraid Connect 插件或更换实现路径）")
	}

	return c.callDockerForceUpdateMutation(ctx, meta, id)
}

var (
	ErrAlreadyStopped = errors.New("container already stopped")
	ErrAlreadyStarted = errors.New("container already started")
)

func (c *Client) findContainerIDByName(ctx context.Context, name string) (string, error) {
	const q = `query { docker { containers { id names state status } } }`
	var resp struct {
		Docker struct {
			Containers []struct {
				ID    string      `json:"id"`
				Names interface{} `json:"names"`
				State string      `json:"state"`
			} `json:"containers"`
		} `json:"docker"`
	}
	if err := c.do(ctx, q, nil, &resp); err != nil {
		return "", err
	}

	seen := make(map[string]struct{}, 64)
	var candidates []string
	want := normalizeName(name)
	for _, ct := range resp.Docker.Containers {
		for _, n := range normalizeContainerNames(ct.Names) {
			nn := normalizeName(n)
			if nn == want {
				return normalizePrefixedID(ct.ID), nil
			}
			if nn == "" {
				continue
			}
			if _, ok := seen[nn]; ok {
				continue
			}
			seen[nn] = struct{}{}
			candidates = append(candidates, nn)
		}
	}

	sort.Strings(candidates)
	if len(candidates) > 0 {
		const max = 10
		if len(candidates) > max {
			candidates = candidates[:max]
		}
		return "", fmt.Errorf("未找到容器：%s（可选容器示例：%s）", name, strings.Join(candidates, ", "))
	}
	return "", fmt.Errorf("未找到容器：%s", name)
}

func (c *Client) stopContainer(ctx context.Context, id string) error {
	const q = `mutation Stop($dockerId: PrefixedID!) { docker { stop(id: $dockerId) { id state status } } }`
	var resp struct {
		Docker struct {
			Stop struct {
				State string `json:"state"`
			} `json:"stop"`
		} `json:"docker"`
	}
	if err := c.do(ctx, q, map[string]interface{}{"dockerId": id}, &resp); err != nil {
		if strings.Contains(err.Error(), "already") && strings.Contains(err.Error(), "stopped") {
			return ErrAlreadyStopped
		}
		return err
	}
	if strings.EqualFold(resp.Docker.Stop.State, "exited") || strings.EqualFold(resp.Docker.Stop.State, "stopped") {
		return nil
	}
	return nil
}

func (c *Client) startContainer(ctx context.Context, id string) error {
	const q = `mutation Start($dockerId: PrefixedID!) { docker { start(id: $dockerId) { id state status } } }`
	var resp struct {
		Docker struct {
			Start struct {
				State string `json:"state"`
			} `json:"start"`
		} `json:"docker"`
	}
	if err := c.do(ctx, q, map[string]interface{}{"dockerId": id}, &resp); err != nil {
		if strings.Contains(err.Error(), "already") && strings.Contains(err.Error(), "started") {
			return ErrAlreadyStarted
		}
		return err
	}
	return nil
}

func normalizePrefixedID(id string) string {
	if parts := strings.SplitN(id, ":", 2); len(parts) == 2 {
		return parts[1]
	}
	return id
}

func normalizeName(name string) string {
	return strings.TrimPrefix(strings.TrimSpace(name), "/")
}

func normalizeContainerNames(v interface{}) []string {
	switch vv := v.(type) {
	case string:
		return []string{vv}
	case []interface{}:
		var ret []string
		for _, item := range vv {
			if s, ok := item.(string); ok {
				ret = append(ret, s)
			}
		}
		return ret
	default:
		return nil
	}
}

func (c *Client) do(ctx context.Context, query string, variables map[string]interface{}, out interface{}) error {
	body, err := json.Marshal(graphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cfg.APIKey)
	if c.cfg.Origin != "" {
		req.Header.Set("Origin", c.cfg.Origin)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
		return fmt.Errorf("unraid graphql http status %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}

	var raw graphQLResponse
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return err
	}
	if len(raw.Errors) > 0 {
		var msgs []string
		for _, e := range raw.Errors {
			if e.Message != "" {
				msgs = append(msgs, e.Message)
			}
		}
		if len(msgs) == 0 {
			return errors.New("graphql error")
		}
		return fmt.Errorf("graphql error: %s", strings.Join(msgs, "; "))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw.Data, out)
}

type dockerMutationMeta struct {
	FieldName            string
	ArgName              string
	ArgType              string
	ReturnNeedsSelection bool
}

func (c *Client) detectDockerForceUpdateMutation(ctx context.Context) (dockerMutationMeta, bool, error) {
	dockerTypeName, err := c.lookupDockerMutationTypeName(ctx)
	if err != nil {
		return dockerMutationMeta{}, false, err
	}
	if dockerTypeName == "" {
		return dockerMutationMeta{}, false, nil
	}

	fields, err := c.lookupTypeFieldsMeta(ctx, dockerTypeName)
	if err != nil {
		return dockerMutationMeta{}, false, err
	}

	if c.cfg.ForceUpdateMutation != "" {
		f, ok := fields[c.cfg.ForceUpdateMutation]
		if !ok {
			return dockerMutationMeta{}, false, fmt.Errorf("未找到配置的 unraid.force_update_mutation: %s", c.cfg.ForceUpdateMutation)
		}
		argName, argType, ok := pickIDArg(f.Args)
		if !ok {
			return dockerMutationMeta{}, false, fmt.Errorf("unraid.force_update_mutation 参数不支持 id/dockerId: %s", c.cfg.ForceUpdateMutation)
		}
		return dockerMutationMeta{
			FieldName:            c.cfg.ForceUpdateMutation,
			ArgName:              argName,
			ArgType:              argType,
			ReturnNeedsSelection: f.Type.RequiresSelectionSet(),
		}, true, nil
	}

	candidates := []string{
		"forceUpdate",
		"forceUpdateDocker",
		"force_update",
		"update",
		"updateContainer",
		"updateDocker",
		"update_container",
		"recreate",
		"pull",
	}
	for _, name := range candidates {
		f, ok := fields[name]
		if !ok {
			continue
		}

		argName, argType, ok := pickIDArg(f.Args)
		if !ok {
			continue
		}

		meta := dockerMutationMeta{
			FieldName:            name,
			ArgName:              argName,
			ArgType:              argType,
			ReturnNeedsSelection: f.Type.RequiresSelectionSet(),
		}
		return meta, true, nil
	}

	var names []string
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		lower := strings.ToLower(name)
		if !strings.Contains(lower, "update") && !strings.Contains(lower, "pull") && !strings.Contains(lower, "recreate") {
			continue
		}

		f, ok := fields[name]
		if !ok {
			continue
		}
		argName, argType, ok := pickIDArg(f.Args)
		if !ok {
			continue
		}
		meta := dockerMutationMeta{
			FieldName:            name,
			ArgName:              argName,
			ArgType:              argType,
			ReturnNeedsSelection: f.Type.RequiresSelectionSet(),
		}
		return meta, true, nil
	}

	return dockerMutationMeta{}, false, nil
}

func (c *Client) lookupDockerMutationTypeName(ctx context.Context) (string, error) {
	const q = `query { __schema { mutationType { fields { name type { name kind ofType { name kind ofType { name kind } } } } } } }`
	var resp struct {
		Schema struct {
			MutationType struct {
				Fields []struct {
					Name string `json:"name"`
					Type struct {
						Name   string `json:"name"`
						Kind   string `json:"kind"`
						OfType *struct {
							Name   string `json:"name"`
							Kind   string `json:"kind"`
							OfType *struct {
								Name string `json:"name"`
								Kind string `json:"kind"`
							} `json:"ofType"`
						} `json:"ofType"`
					} `json:"type"`
				} `json:"fields"`
			} `json:"mutationType"`
		} `json:"__schema"`
	}
	if err := c.do(ctx, q, nil, &resp); err != nil {
		return "", err
	}

	for _, f := range resp.Schema.MutationType.Fields {
		if f.Name != "docker" {
			continue
		}
		if f.Type.Name != "" {
			return f.Type.Name, nil
		}
		if f.Type.OfType != nil && f.Type.OfType.Name != "" {
			return f.Type.OfType.Name, nil
		}
	}
	return "", nil
}

type gqlTypeRef struct {
	Kind   string      `json:"kind"`
	Name   string      `json:"name"`
	OfType *gqlTypeRef `json:"ofType"`
}

func (t gqlTypeRef) String() string {
	switch t.Kind {
	case "NON_NULL":
		if t.OfType == nil {
			return "String!"
		}
		return t.OfType.String() + "!"
	case "LIST":
		if t.OfType == nil {
			return "[String]"
		}
		return "[" + t.OfType.String() + "]"
	default:
		if t.Name != "" {
			return t.Name
		}
		if t.OfType != nil {
			return t.OfType.String()
		}
		return "String"
	}
}

func (t gqlTypeRef) RequiresSelectionSet() bool {
	base := t.baseKind()
	return base == "OBJECT" || base == "INTERFACE" || base == "UNION"
}

func (t gqlTypeRef) baseKind() string {
	switch t.Kind {
	case "NON_NULL", "LIST":
		if t.OfType == nil {
			return t.Kind
		}
		return t.OfType.baseKind()
	default:
		return t.Kind
	}
}

type gqlArgMeta struct {
	Name string     `json:"name"`
	Type gqlTypeRef `json:"type"`
}

type gqlFieldMeta struct {
	Name string       `json:"name"`
	Args []gqlArgMeta `json:"args"`
	Type gqlTypeRef   `json:"type"`
}

func (c *Client) lookupTypeFieldsMeta(ctx context.Context, typeName string) (map[string]gqlFieldMeta, error) {
	const q = `query($name: String!) { __type(name: $name) { fields { name args { name type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } } }`
	var resp struct {
		Type struct {
			Fields []gqlFieldMeta `json:"fields"`
		} `json:"__type"`
	}
	if err := c.do(ctx, q, map[string]interface{}{"name": typeName}, &resp); err != nil {
		return nil, err
	}
	ret := make(map[string]gqlFieldMeta, len(resp.Type.Fields))
	for _, f := range resp.Type.Fields {
		if f.Name == "" {
			continue
		}
		ret[f.Name] = f
	}
	return ret, nil
}

func pickIDArg(args []gqlArgMeta) (string, string, bool) {
	for _, a := range args {
		if a.Name == "id" {
			return "id", a.Type.String(), true
		}
	}
	for _, a := range args {
		if a.Name == "dockerId" {
			return "dockerId", a.Type.String(), true
		}
	}
	if len(args) == 1 {
		return args[0].Name, args[0].Type.String(), true
	}
	return "", "", false
}

func (c *Client) callDockerForceUpdateMutation(ctx context.Context, meta dockerMutationMeta, id string) error {
	selection := ""
	if meta.ReturnNeedsSelection {
		selection = ` { __typename }`
	}
	q := fmt.Sprintf(`mutation ForceUpdate($v: %s) { docker { %s(%s: $v)%s } }`, meta.ArgType, meta.FieldName, meta.ArgName, selection)
	var raw map[string]interface{}
	if err := c.do(ctx, q, map[string]interface{}{"v": id}, &raw); err != nil {
		return err
	}
	return nil
}
