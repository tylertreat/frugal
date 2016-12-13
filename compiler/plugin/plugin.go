package plugin

import "plugin"

// FrugalPlugin wraps a plugin.Plugin.
type FrugalPlugin struct {
	Name   string
	plugin *plugin.Plugin
}

// Lookup returns the symbol with the given name or nil if it doesn't exist.
func (f *FrugalPlugin) Lookup(name string) interface{} {
	symbol, _ := f.plugin.Lookup(name)
	return symbol
}

// LoadPlugins opens the FrugalPlugins with the given names and returns an
// error if any fail to open.
func LoadPlugins(names []string) ([]*FrugalPlugin, error) {
	plugins := make([]*FrugalPlugin, len(names))
	for i, name := range names {
		p, err := plugin.Open(name)
		if err != nil {
			return nil, err
		}
		plugins[i] = &FrugalPlugin{Name: name, plugin: p}
	}
	return plugins, nil
}
