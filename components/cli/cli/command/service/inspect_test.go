package service

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
)

func formatServiceInspect(t *testing.T, format formatter.Format, now time.Time) string {
	b := new(bytes.Buffer)

	endpointSpec := &swarm.EndpointSpec{
		Mode: "vip",
		Ports: []swarm.PortConfig{
			{
				Protocol:   swarm.PortConfigProtocolTCP,
				TargetPort: 5000,
			},
		},
	}

	two := uint64(2)

	s := swarm.Service{
		ID: "de179gar9d0o7ltdybungplod",
		Meta: swarm.Meta{
			Version:   swarm.Version{Index: 315},
			CreatedAt: now,
			UpdatedAt: now,
		},
		Spec: swarm.ServiceSpec{
			Annotations: swarm.Annotations{
				Name:   "my_service",
				Labels: map[string]string{"com.label": "foo"},
			},
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Image: "foo/bar@sha256:this_is_a_test",
				},
				Networks: []swarm.NetworkAttachmentConfig{
					{
						Target:  "5vpyomhb6ievnk0i0o60gcnei",
						Aliases: []string{"web"},
					},
				},
			},
			Mode: swarm.ServiceMode{
				Replicated: &swarm.ReplicatedService{
					Replicas: &two,
				},
			},
			EndpointSpec: endpointSpec,
		},
		Endpoint: swarm.Endpoint{
			Spec: *endpointSpec,
			Ports: []swarm.PortConfig{
				{
					Protocol:      swarm.PortConfigProtocolTCP,
					TargetPort:    5000,
					PublishedPort: 30000,
				},
			},
			VirtualIPs: []swarm.EndpointVirtualIP{
				{
					NetworkID: "6o4107cj2jx9tihgb0jyts6pj",
					Addr:      "10.255.0.4/16",
				},
			},
		},
		UpdateStatus: &swarm.UpdateStatus{
			StartedAt:   &now,
			CompletedAt: &now,
		},
	}

	ctx := formatter.Context{
		Output: b,
		Format: format,
	}

	err := formatter.ServiceInspectWrite(ctx, []string{"de179gar9d0o7ltdybungplod"},
		func(ref string) (interface{}, []byte, error) {
			return s, nil, nil
		},
		func(ref string) (interface{}, []byte, error) {
			return types.NetworkResource{
				ID:   "5vpyomhb6ievnk0i0o60gcnei",
				Name: "mynetwork",
			}, nil, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestPrettyPrintWithNoUpdateConfig(t *testing.T) {
	s := formatServiceInspect(t, formatter.NewServiceFormat("pretty"), time.Now())
	if strings.Contains(s, "UpdateStatus") {
		t.Fatal("Pretty print failed before parsing UpdateStatus")
	}
	if !strings.Contains(s, "mynetwork") {
		t.Fatal("network name not found in inspect output")
	}
}

func TestJSONFormatWithNoUpdateConfig(t *testing.T) {
	now := time.Now()
	// s1: [{"ID":..}]
	// s2: {"ID":..}
	s1 := formatServiceInspect(t, formatter.NewServiceFormat(""), now)
	s2 := formatServiceInspect(t, formatter.NewServiceFormat("{{json .}}"), now)
	var m1Wrap []map[string]interface{}
	if err := json.Unmarshal([]byte(s1), &m1Wrap); err != nil {
		t.Fatal(err)
	}
	if len(m1Wrap) != 1 {
		t.Fatalf("strange s1=%s", s1)
	}
	m1 := m1Wrap[0]
	var m2 map[string]interface{}
	if err := json.Unmarshal([]byte(s2), &m2); err != nil {
		t.Fatal(err)
	}
	assert.Check(t, is.DeepEqual(m1, m2))
}
