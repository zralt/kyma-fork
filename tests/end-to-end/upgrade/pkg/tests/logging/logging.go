package logging

import (
	"net/http"

	dex "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/fetch-dex-token"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/jwt"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/logging/pkg/logstream"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// LoggingTest checks if logging continues to work properly after upgrade
type LoggingTest struct {
	coreInterface kubernetes.Interface
	domainName    string
	idpConfig     dex.IdProviderConfig
	httpClient    *http.Client
}

// NewLoggingTest creates a new instance of logging upgrade test
func NewLoggingTest(coreInterface kubernetes.Interface, domainName string, dexConfig dex.IdProviderConfig) LoggingTest {
	return LoggingTest{
		coreInterface: coreInterface,
		domainName:    domainName,
		idpConfig:     dexConfig,
		httpClient:    getHttpClient(),
	}
}

// CreateResources creates resources for logging upgrade test
func (t LoggingTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("Cleaning up before creating resources")
	err := logstream.Cleanup(namespace, t.coreInterface)
	if err != nil {
		return err
	}
	log.Println("Deploying test-counter-pod")
	err = logstream.DeployDummyPod(namespace, t.coreInterface)
	if err != nil {
		return err
	}
	log.Println("Waiting for test-counter-pod to run...")
	err = logstream.WaitForDummyPodToRun(namespace, t.coreInterface)
	if err != nil {
		return err
	}
	log.Println("Test if logs from test-counter-pod are streamed by Loki before upgrade")
	err = t.testLogStream(namespace)
	if err != nil {
		logstream.Cleanup(namespace, t.coreInterface)
		return err
	}
	return nil
}

// TestResources checks if resources are working properly after upgrade
func (t LoggingTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("Test if new logs from test-counter-pod are streamed by Loki after upgrade")
	err := t.testLogStream(namespace)
	if err != nil {
		logstream.Cleanup(namespace, t.coreInterface)
		return err
	}
	log.Println("Deleting test-counter-pod")
	err = logstream.Cleanup(namespace, t.coreInterface)
	if err != nil {
		return err
	}
	return nil
}

func (t LoggingTest) testLogStream(namespace string) error {
	token, err := jwt.GetToken(t.idpConfig)
	if err != nil {
		return errors.Wrap(err, "cannot fetch dex token")
	}
	authHeader := jwt.SetAuthHeader(token)
	err = logstream.Test("container", "count", authHeader, t.httpClient)
	if err != nil {
		return err
	}
	err = logstream.Test("app", "test-counter-pod", authHeader, t.httpClient)
	if err != nil {
		return err
	}
	err = logstream.Test("namespace", namespace, authHeader, t.httpClient)
	if err != nil {
		return err
	}
	return nil
}

func getHttpClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	return client
}
