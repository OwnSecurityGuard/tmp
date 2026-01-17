package pluginstore

import (
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewPluginRepo(db *gorm.DB) *Repo {

	return &Repo{db: db}
}
func (r *Repo) Create(p *PluginModel) error {
	return r.db.Create(p).Error
}

func (r *Repo) Update(p *PluginModel) error {
	return r.db.Model(&PluginModel{}).
		Where("name = ?", p.Name).
		Updates(p).Error
}
func (r *Repo) Get(name string) (*PluginModel, error) {
	var p PluginModel
	if err := r.db.First(&p, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) List() ([]*PluginModel, error) {
	var list []*PluginModel
	r.db.Find(&list)

	return list, nil
}
