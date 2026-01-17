package pluginstore

type PluginRepo interface {
	Create(p *PluginModel) error
	Update(p *PluginModel) error
	Get(name string) (*PluginModel, error)
	List() ([]*PluginModel, error)
}
