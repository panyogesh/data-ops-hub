type Config struct {
	KubeContext string
	Namespace   string

	InstallCrossplane   bool
	CrossplaneRelease   string
	CrossplaneNamespace string

	GCPProjectID       string
	GCPSecretName      string
	GCPSecretNamespace string
	GCPSecretKey       string
	GCPSecretFile      string
	ProviderConfigName string

	NetworkName    string
	SubnetworkName string
	FirewallName   string
	InstanceName   string

	Region      string
	Zone        string
	MachineType string
	Image       string

	CleanupMode string // none|gcp|provider|all
}