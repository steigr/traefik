package provider

import (
	"encoding/json"
	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/types"
	"reflect"
	"testing"
)

func TestLoadIngresses(t *testing.T) {
	ingresses := []k8s.Ingress{{
		Spec: k8s.IngressSpec{
			Rules: []k8s.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: k8s.IngressRuleValue{
						HTTP: &k8s.HTTPIngressRuleValue{
							Paths: []k8s.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: k8s.IngressBackend{
										ServiceName: "service1",
										ServicePort: k8s.FromString("http"),
									},
								},
							},
						},
					},
				},
				{
					Host: "bar",
					IngressRuleValue: k8s.IngressRuleValue{
						HTTP: &k8s.HTTPIngressRuleValue{
							Paths: []k8s.HTTPIngressPath{
								{
									Backend: k8s.IngressBackend{
										ServiceName: "service3",
										ServicePort: k8s.FromInt(443),
									},
								},
								{
									Backend: k8s.IngressBackend{
										ServiceName: "service2",
										ServicePort: k8s.FromInt(802),
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service1",
				UID:  "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service2",
				UID:  "2",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []k8s.ServicePort{
					{
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service3",
				UID:  "3",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 443,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 1,
					},
					"3": {
						URL:    "https://10.0.0.3:443",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"bar": {
						Rule: "Host:bar",
					},
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestRuleType(t *testing.T) {
	ingresses := []k8s.Ingress{
		{
			ObjectMeta: k8s.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathPrefixStrip"}, //camel case
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo1",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar1",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "path"}, //lower case
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo1",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar2",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathPrefix"}, //path prefix
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo2",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar1",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathStrip"}, //path strip
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo2",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar2",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathXXStrip"}, //wrong rule
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo1",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar3",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service1",
				UID:  "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{disablePassHostHeaders: true}
	actualConfig, err := provider.loadIngresses(client)
	actual := actualConfig.Frontends
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := map[string]*types.Frontend{
		"foo1/bar1": {
			Backend: "foo1/bar1",
			Routes: map[string]types.Route{
				"/bar1": {
					Rule: "PathPrefixStrip:/bar1",
				},
				"foo1": {
					Rule: "Host:foo1",
				},
			},
		},
		"foo1/bar2": {
			Backend: "foo1/bar2",
			Routes: map[string]types.Route{
				"/bar2": {
					Rule: "Path:/bar2",
				},
				"foo1": {
					Rule: "Host:foo1",
				},
			},
		},
		"foo2/bar1": {
			Backend: "foo2/bar1",
			Routes: map[string]types.Route{
				"/bar1": {
					Rule: "PathPrefix:/bar1",
				},
				"foo2": {
					Rule: "Host:foo2",
				},
			},
		},
		"foo2/bar2": {
			Backend: "foo2/bar2",
			Routes: map[string]types.Route{
				"/bar2": {
					Rule: "PathStrip:/bar2",
				},
				"foo2": {
					Rule: "Host:foo2",
				},
			},
		},
		"foo1/bar3": {
			Backend: "foo1/bar3",
			Routes: map[string]types.Route{
				"/bar3": {
					Rule: "PathPrefix:/bar3",
				},
				"foo1": {
					Rule: "Host:foo1",
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestGetPassHostHeader(t *testing.T) {
	ingresses := []k8s.Ingress{{
		ObjectMeta: k8s.ObjectMeta{
			Namespace: "awesome",
		},
		Spec: k8s.IngressSpec{
			Rules: []k8s.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: k8s.IngressRuleValue{
						HTTP: &k8s.HTTPIngressRuleValue{
							Paths: []k8s.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: k8s.IngressBackend{
										ServiceName: "service1",
										ServicePort: k8s.FromInt(801),
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service1",
				Namespace: "awesome",
				UID:       "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{disablePassHostHeaders: true}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend: "foo/bar",
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestOnlyReferencesServicesFromOwnNamespace(t *testing.T) {
	ingresses := []k8s.Ingress{
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Backend: k8s.IngressBackend{
											ServiceName: "service",
											ServicePort: k8s.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service",
				UID:       "1",
				Namespace: "awesome",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service",
				UID:       "2",
				Namespace: "not-awesome",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:80",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo": {
				Backend:        "foo",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestLoadNamespacedIngresses(t *testing.T) {
	ingresses := []k8s.Ingress{
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
					{
						Host: "bar",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Backend: k8s.IngressBackend{
											ServiceName: "service3",
											ServicePort: k8s.FromInt(443),
										},
									},
									{
										Backend: k8s.IngressBackend{
											ServiceName: "service2",
											ServicePort: k8s.FromInt(802),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "not-awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "baz",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/baz",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
				Name:      "service1",
				UID:       "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service1",
				Namespace: "not-awesome",
				UID:       "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service2",
				Namespace: "awesome",
				UID:       "2",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []k8s.ServicePort{
					{
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service3",
				Namespace: "awesome",
				UID:       "3",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 443,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{
		Namespaces: []string{"awesome"},
	}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 1,
					},
					"3": {
						URL:    "https://10.0.0.3:443",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"bar": {
						Rule: "Host:bar",
					},
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestLoadMultipleNamespacedIngresses(t *testing.T) {
	ingresses := []k8s.Ingress{
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "foo",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/bar",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
					{
						Host: "bar",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Backend: k8s.IngressBackend{
											ServiceName: "service3",
											ServicePort: k8s.FromInt(443),
										},
									},
									{
										Backend: k8s.IngressBackend{
											ServiceName: "service2",
											ServicePort: k8s.FromInt(802),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "somewhat-awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "awesome",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/quix",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "not-awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "baz",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/baz",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Name:      "service1",
				Namespace: "awesome",
				UID:       "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "somewhat-awesome",
				Name:      "service1",
				UID:       "17",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.4",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
				Name:      "service2",
				UID:       "2",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []k8s.ServicePort{
					{
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
				Name:      "service3",
				UID:       "3",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 443,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{
		Namespaces: []string{"awesome", "somewhat-awesome"},
	}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 1,
					},
					"3": {
						URL:    "https://10.0.0.3:443",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"awesome/quix": {
				Servers: map[string]types.Server{
					"17": {
						URL:    "http://10.0.0.4:801",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"bar": {
						Rule: "Host:bar",
					},
				},
			},
			"awesome/quix": {
				Backend:        "awesome/quix",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/quix": {
						Rule: "PathPrefix:/quix",
					},
					"awesome": {
						Rule: "Host:awesome",
					},
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

type clientMock struct {
	ingresses []k8s.Ingress
	services  []k8s.Service
	watchChan chan interface{}
}

func (c clientMock) GetIngresses(predicate func(k8s.Ingress) bool) ([]k8s.Ingress, error) {
	var ingresses []k8s.Ingress
	for _, ingress := range c.ingresses {
		if predicate(ingress) {
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses, nil
}
func (c clientMock) WatchIngresses(predicate func(k8s.Ingress) bool, stopCh <-chan bool) (chan interface{}, chan error, error) {
	return c.watchChan, make(chan error), nil
}
func (c clientMock) GetServices(predicate func(k8s.Service) bool) ([]k8s.Service, error) {
	var services []k8s.Service
	for _, service := range c.services {
		if predicate(service) {
			services = append(services, service)
		}
	}
	return services, nil
}
func (c clientMock) WatchAll(stopCh <-chan bool) (chan interface{}, chan error, error) {
	return c.watchChan, make(chan error), nil
}
