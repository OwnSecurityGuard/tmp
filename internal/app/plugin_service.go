package app

import (
	"fmt"
	"proxy-system-backend/internal/modules/plugin"
	pluginstore "proxy-system-backend/internal/storage/plugin"
	"time"
)

type PluginService struct {
	repo *pluginstore.Repo
	mgr  *plugin.Manager
}

func NewPluginService(
	repo *pluginstore.Repo,
	mgr *plugin.Manager,
) *PluginService {
	return &PluginService{
		repo: repo,
		mgr:  mgr,
	}
}
func (s *PluginService) Register(name, path string) error {
	fmt.Println("Register", name, path)
	_, err := s.repo.Get(name)
	if err == nil {
		return fmt.Errorf("plugin %s already exists", name)
	}

	now := time.Now().Unix()

	return s.repo.Create(&pluginstore.PluginModel{
		Name:      name,
		Path:      path,
		Status:    "registered",
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *PluginService) Load(name string) error {
	p, err := s.repo.Get(name)
	if err != nil {
		return err
	}

	if err = s.mgr.Load(p.Name); err != nil {
		now := time.Now().Unix()
		_ = s.repo.Update(&pluginstore.PluginModel{
			Name:      p.Name,
			Status:    "error",
			Error:     err.Error(),
			UpdatedAt: now,
		})
		return err
	}

	now := time.Now().Unix()
	return s.repo.Update(&pluginstore.PluginModel{
		Name:      p.Name,
		Status:    "running",
		Error:     "",
		LoadedAt:  now,
		UpdatedAt: now,
	})
}
func (s *PluginService) Unload(name string) error {
	if err := s.mgr.Unload(name); err != nil {
		return err
	}

	now := time.Now().Unix()
	return s.repo.Update(&pluginstore.PluginModel{
		Name:      name,
		Status:    "stopped",
		UpdatedAt: now,
	})
}

func (s *PluginService) Bootstrap() error {
	list, err := s.repo.List()
	if err != nil {
		return err
	}
	fmt.Println("Bootstrap", len(list))
	for _, p := range list {
		s.mgr.Register(p.Name, p.Path)
		//_ = s.mgr.Load(p.Name)

	}
	return nil
}
func (s *PluginService) List() ([]*plugin.PluginInfo, error) {
	list, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	var out []*plugin.PluginInfo
	for _, p := range list {
		out = append(out, &plugin.PluginInfo{
			Name:      p.Name,
			Path:      p.Path,
			Status:    plugin.PluginStatus(p.Status),
			Error:     p.Error,
			LoadedAt:  p.LoadedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}
	return out, nil
}

func (s *PluginService) Get(name string) (*plugin.PluginInfo, error) {
	p, err := s.repo.Get(name)
	if err != nil {
		return nil, err
	}

	return &plugin.PluginInfo{
		Name:      p.Name,
		Path:      p.Path,
		Status:    plugin.PluginStatus(p.Status),
		Error:     p.Error,
		LoadedAt:  p.LoadedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

func (s *PluginService) Decode(name string, isClient bool, payload []byte) (*plugin.DecodeResult, error) {

	return s.mgr.Decode(name, isClient, payload)

}

// GetPluginManager 获取内部的 PluginManager
func (s *PluginService) GetPluginManager() *plugin.Manager {
	return s.mgr
}
