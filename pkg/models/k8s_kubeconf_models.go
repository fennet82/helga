package models

type KubeConfig struct {
	APIVersion     string              `yaml:"apiVersion"`
	Kind           string              `yaml:"kind"`
	CurrentContext string              `yaml:"current-context"`
	Clusters       []KubeConfigCluster `yaml:"clusters"`
	Users          []KubeConfigUser    `yaml:"users"`
	Contexts       []KubeConfigContext `yaml:"contexts"`
}

type KubeConfigCluster struct {
	Name    string                   `yaml:"name"`
	Cluster KubeConfigClusterDetails `yaml:"cluster"`
}

type KubeConfigClusterDetails struct {
	Server                   string `yaml:"server"`
	InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify"`
	CertificateAuthorityFile string `yaml:"certificate-authority-file,omitempty"`
}

type KubeConfigUser struct {
	Name string                `yaml:"name"`
	User KubeConfigUserDetails `yaml:"user"`
}

type KubeConfigUserDetails struct {
	Username string `yaml:"username"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

type KubeConfigContext struct {
	Name    string                   `yaml:"name"`
	Context KubeConfigContextDetails `yaml:"context"`
}

type KubeConfigContextDetails struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace"`
}
