package models

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/fennet82/helga/internal/logger"
	"github.com/fennet82/helga/internal/utils"
	"github.com/fennet82/helga/internal/vars"
	helga_errors "github.com/fennet82/helga/pkg/errors"
	helmclient "github.com/mittwald/go-helm-client"
	"gopkg.in/yaml.v2"
)

type Cluster struct {
	Name                  string       `yaml:"name"`
	Server                string       `yaml:"server"`
	Username              string       `yaml:"username"`
	Password              string       `yaml:"password,omitempty"` // Optional, used for basic authentication
	Token                 string       `yaml:"token,omitempty"`    // Optional, used for token-based authentication
	InsecureSkipTLSVerify bool         `yaml:"insecure_skip_tls_verify"`
	CACertFilePath        string       `yaml:"ca_cert_file_path"`
	Namespaces            []*Namespace `yaml:"namespaces"`
	helmClient            helmclient.Client
}

func (c *Cluster) String() string {
	return c.Name
}

func (c *Cluster) Validate() []error {
	var (
		validationErrs []error
		structName     = "Cluster"
		dReg           = regexp.MustCompile(vars.K8S_API_URL_REGEX)
	)

	if c.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("name field cannot be empty")})
	}

	if !dReg.MatchString(c.Server) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"server: %s, did not pass regex validation please refer to this regex for fixing: %s", c.Server, vars.K8S_API_URL_REGEX,
		)})
	}

	if c.Username == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")})
	}

	if c.Password == "" && c.Token == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("both password and token cannot be empty, please choose one")})
	}

	if !c.InsecureSkipTLSVerify {
		if c.CACertFilePath == "" {
			validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("ca_cert_file_path field cannot be empty when insecure_skip_tls_verify is true")})
		}
	}

	errs, filteredNameSpace := utils.FilterByValidation(utils.ToValidatableSlice(c.Namespaces), "namespace: %s did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	c.Namespaces = utils.FromValidatableSlice[*Namespace](filteredNameSpace)

	if len(c.Namespaces) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("namespaces list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (dest *Cluster) Sync(src *Cluster) error {
	if dest == nil || src == nil {
		return fmt.Errorf("cannot sync nil Cluster objects")
	}

	if src.Name != "" && dest.Name == "" {
		dest.Name = src.Name
	}

	if src.Server != "" && dest.Server == "" {
		dest.Server = src.Server
	}

	if src.Username != "" && dest.Username == "" {
		dest.Username = src.Username
	}

	if src.Password != "" && dest.Password == "" {
		dest.Password = src.Password
	}

	if src.Token != "" && dest.Token == "" {
		dest.Token = src.Token
	}

	if src.InsecureSkipTLSVerify && !dest.InsecureSkipTLSVerify {
		dest.InsecureSkipTLSVerify = true
	}

	if src.CACertFilePath != "" && dest.CACertFilePath == "" {
		dest.CACertFilePath = src.CACertFilePath
	}

	return nil
}

func (c *Cluster) generateKubeConfig() ([]byte, error) {
	kubeConfig := KubeConfig{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters:   []KubeConfigCluster{},
		Users:      []KubeConfigUser{},
		Contexts:   []KubeConfigContext{},
	}

	// Add cluster
	kubeCluster := KubeConfigCluster{
		Name: c.Name,
		Cluster: KubeConfigClusterDetails{
			Server:                c.Server,
			InsecureSkipTLSVerify: c.InsecureSkipTLSVerify,
		},
	}

	if !c.InsecureSkipTLSVerify {
		kubeCluster.Cluster.CertificateAuthorityFile = c.CACertFilePath
	}

	kubeConfig.Clusters = append(kubeConfig.Clusters, kubeCluster)

	// Add user
	kubeUser := KubeConfigUser{
		Name: c.Username,
		User: KubeConfigUserDetails{
			Username: c.Username,
			Password: c.Password,
			Token:    c.Token,
		},
	}

	kubeConfig.Users = append(kubeConfig.Users, kubeUser)

	// Generate contexts for each namespace
	for _, namespace := range c.Namespaces {
		contextName := fmt.Sprintf("%s-%s", c.Name, namespace.Name)
		kubeContext := KubeConfigContext{
			Name: contextName,
			Context: KubeConfigContextDetails{
				Cluster:   c.Name,
				User:      c.Username,
				Namespace: namespace.Name,
			},
		}

		kubeConfig.Contexts = append(kubeConfig.Contexts, kubeContext)
	}

	// Set current context to first context if available
	if len(kubeConfig.Contexts) > 0 {
		kubeConfig.CurrentContext = kubeConfig.Contexts[0].Name
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kubeconfig: %w", err)
	}

	fmt.Println(string(yamlData))

	return yamlData, nil
}

func (c *Cluster) InitiateHelmClientByKubeCtx(KubeCtxName string) error {
	kconf, err := c.generateKubeConfig()
	if err != nil {
		return helga_errors.ErrHelmClient{ErrMsg: fmt.Sprintf("error generating kubeconf for cluster: %s\n derived from err: %s", c.Name, err.Error())}
	}

	options := &helmclient.KubeConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        "default",
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			DebugLog: func(format string, v ...any) {
				logger.GetLoggerInstance().Debug(fmt.Sprintf(format, v...))
			},
		},
		KubeConfig:  kconf,
		KubeContext: KubeCtxName,
	}

	c.helmClient, err = helmclient.NewClientFromKubeConf(options)
	if err != nil {
		return helga_errors.ErrHelmClient{ErrMsg: fmt.Sprintf("error getting helmClient for cluster: %s\n derived from err: %s", c.Name, err.Error())}
	}

	return nil
}

func (c *Cluster) SyncHelmCharts() error {
	if err := c.helmClient.UpdateChartRepos(); err != nil {
		return helga_errors.ErrInSyncProcess{ErrMsg: fmt.Sprintf("error updating helm chart repos for cluster: %s\n derived from err: %s", c.Name, err.Error())}
	}

	

	return nil
}
