## Plugin

Plugin is an important feature of CBCScan Essentials.
It is very convenient to save **Extrinsic** and **Event** to the database and parse it into various customized content, save it in a new data table, and display the customization data to the frontend through HTTP API.

### Usage

1. Refer [plugin](https://github.com/itering/cbcscan-plugin) to write the plugin you need

1. Just import your plugin in ``plugins/registry.go`` like

```
func init() {
	registerNative(YourPlugin.New()) // Register plugin to cbcscan
}
```