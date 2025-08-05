package registry

import "github.com/PROJECT_NAME/internal/clients/testclient"

func (r *Registry) TestClient() testclient.Client {
	if r.testClient == nil {
		r.testClient = testclient.NewClient(r)
	}

	return r.testClient
}
