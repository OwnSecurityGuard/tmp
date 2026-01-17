package pluginstore

type PluginModel struct {
	Name   string `gorm:"primaryKey;size:128"`
	Path   string `gorm:"not null"`
	Status string `gorm:"size:32;not null"`
	Error  string `gorm:"type:text"`

	CreatedAt int64
	UpdatedAt int64
	LoadedAt  int64
}
