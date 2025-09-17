package agriroutertestcontainer

import (
	"context"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmaxmax/go-sse"
)

// AgrirouterContainer represents a test container for the agrirouter service.
type AgrirouterContainer struct {
	testcontainers.Container
	BaseURL string
	Events  *TestEvents
}

const httpPort = "8080/tcp"

func (c *AgrirouterContainer) getBaseURL(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, httpPort)
	if err != nil {
		return "", err
	}
	port := mappedPort.Port()

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return "http://" + host + ":" + port, nil
}

// OkHTTPCode returns a function that checks if the status code is equal to the provided goodCode.
func OkHTTPCode(goodCode int) func(int) bool {
	return func(statusCode int) bool {
		return statusCode == goodCode
	}
}

// Run starts the agrirouter test container and returns an AgrirouterContainer instance.
func Run(ctx context.Context) (*AgrirouterContainer, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",
			Dockerfile: "internal/tests/test_server/Dockerfile",
			Repo:       "agrirouter-test-server",
			Tag:        "latest",
		},
		ExposedPorts: []string{httpPort},
		WaitingFor: wait.ForHTTP("/").WithStatusCodeMatcher(
			OkHTTPCode(http.StatusNotFound), // root would return 404, means the server is up and running
		).WithPort(httpPort),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	agrirouterContainer := &AgrirouterContainer{Container: container}

	baseURL, err := agrirouterContainer.getBaseURL(ctx)
	if err != nil {
		return agrirouterContainer, err
	}

	events := TestEvents{}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in event listener: %v", r)
				return
			}
		}()
		listenForEvents(ctx, baseURL, &events)
	}()

	return &AgrirouterContainer{
		Container: container,
		BaseURL:   baseURL,
		Events:    &events,
	}, nil
}

func listenForEvents(ctx context.Context, baseURL string, events *TestEvents) {
	eventsURL := baseURL + "/_testEvents"
	log.Printf("Connecting to test events at %s", eventsURL)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, eventsURL, nil)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to connect to test events: %v", err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	for ev, err := range sse.Read(res.Body, nil) {
		if err != nil {
			log.Printf("Error reading events: %v", err)
			break
		}
		events.add(ev.Type, ev.Data)
		log.Printf("Received event: %s - %s", ev.Type, ev.Data)
	}
}

// TerminateOrLog terminates the agrirouter container and logs any errors that occur during termination.
func (c *AgrirouterContainer) TerminateOrLog() {
	if err := c.Terminate(context.Background()); err != nil {
		log.Printf("Failed to terminate agrirouter container: %v", err)
	}
}
