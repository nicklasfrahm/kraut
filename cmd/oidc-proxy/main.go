package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nicklasfrahm/kraut/pkg/log"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// TODO: Create a helm chart for this.

var (
	version    = "dev"
	kubeconfig string
	help       bool
)

// HTTPError is a custom error type that
// is used to return a JSON response.
type HTTPError struct {
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func main() {

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.BoolVar(&help, "help", false, "display help")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		return
	}

	logger := log.NewLogger()

	clientSet, err := createKubernetesClientset()
	if err != nil {
		logger.Fatal("failed to create Kubernetes clientset", zap.Error(err))
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			timeStart := time.Now()

			if err := next(c); err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			logger.Info(fmt.Sprintf("%d %s %s %s %s", res.Status, req.Method, req.URL.Path, byteCountDecimal(c.Response().Size), time.Since(timeStart).String()))

			return nil
		}
	})

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok {
			c.JSON(he.Code, HTTPError{
				Code:    he.Code,
				Title:   http.StatusText(he.Code),
				Message: fmt.Sprintf("%v", he.Message),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, HTTPError{
			Code:    http.StatusInternalServerError,
			Title:   http.StatusText(http.StatusInternalServerError),
			Message: fmt.Sprintf("%v", err),
		})
	}

	openIDConfigPath := "/.well-known/openid-configuration"
	e.GET(openIDConfigPath, makePrerenderedDocumentHander(openIDConfigPath, clientSet))

	jwksPath := "/openid/v1/jwks"
	e.GET(jwksPath, makePrerenderedDocumentHander(jwksPath, clientSet))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info(fmt.Sprintf("Version: %s", version))
	logger.Info(fmt.Sprintf("Starting OIDC proxy: http://0.0.0.0:%s", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

// createKubernetesClientset creates a Kubernetes clientset from the given config.
// It automatically detects whether it is running inside a cluster or not based on
// the existence of the `KUBERNETES_SERVICE_HOST` environment variable.
func createKubernetesClientset() (*kubernetes.Clientset, error) {
	isRunningOutsideCluster := os.Getenv("KUBERNETES_SERVICE_HOST") == ""
	if isRunningOutsideCluster {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load local kubeconfig: %w", err)
		}

		return kubernetes.NewForConfig(config)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster kubeconfig: %w", err)
	}

	return kubernetes.NewForConfig(config)
}

// makePrerenderedDocumentHander creates a handler that returns a prerendered document.
func makePrerenderedDocumentHander(path string, clientSet *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Seems weird to me but currently the document is prerendered and therefore
		// does not have a data structure that can be used to generate the JSON response.
		// See: https://github.com/kubernetes/kubernetes/blob/master/pkg/serviceaccount/openidmetadata.go
		info, err := clientSet.RESTClient().Get().AbsPath(path).DoRaw(c.Request().Context())
		if err != nil {
			return err
		}

		// This ensures that the JSON response is properly formatted.
		var document interface{}
		if err := json.Unmarshal(info, &document); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, document)
	}
}

// byteCountDecimal returns a human-readable byte
// string of the form 10M, 12.5K, and so forth.
func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	exponent := int(math.Floor(math.Log(float64(b)) / math.Log(unit)))
	value := float64(b) / math.Pow(float64(unit), float64(exponent))
	return fmt.Sprintf("%.f%cB", value, "kMGTPE"[exponent])
}
