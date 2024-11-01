package configserver

import (
	"github.com/meidoworks/nekoq-component/configure/configapi"
)

type writeServer struct {
	DataWriter configapi.DataWriter
}

func (w *writeServer) Startup() error {
	if err := w.DataWriter.Startup(); err != nil {
		return err
	}
	return nil
}

func (w *writeServer) Stop() error {
	if err := w.DataWriter.Stop(); err != nil {
		return err
	}
	return nil
}

func (w *writeServer) SaveConfiguration(cfg *configapi.Configuration) error {
	return w.DataWriter.SaveConfiguration(*cfg)
}

func (w *writeServer) DeleteConfiguration(group, key, sel, optSel string) (bool, error) {
	return w.DataWriter.DeleteConfiguration(group, key, sel, optSel)
}
