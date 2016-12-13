package plugin

import "plugin"

type FrugalPlugin struct {
	plugin *plugin.Plugin
}

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
		plugins[i] = &FrugalPlugin{p}
	}
	return plugins, nil
}
