package plugin

import "plugin"

type FrugalPlugin struct {
	plugin *plugin.Plugin
}

func (f *FrugalPlugin) Lookup(name string) interface{} {
	symbol, _ := f.plugin.Lookup(name)
	return symbol
}

// LoadPlugins opens the FrugalPlugin with the given name and returns an error
// if it fails to open.
func LoadPlugin(name string) (*FrugalPlugin, error) {
	p, err := plugin.Open(name)
	if err != nil {
		return nil, err
	}
	return &FrugalPlugin{p}, nil
}
