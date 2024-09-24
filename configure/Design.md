# Configuration

### Concepts

##### Configuration

Identification fields:

* Group
* Key
* Version
* Selectors
* OptionalSelectors

Key Data fields:

* Value
* Signature

##### Selector

Selector is the core functionality for filtering out desired configurations.

Selector consists of `Key` and `Value` fields the same to an entry of a map. Both fields MUST NOT be empty.

Valid characters of the selector fields are:

* Alphabets - [A-Za-z]
* Numbers - [0-9]
* Underscore - [_]

`Bahavior of Selectors:`

1. Selector matching will ALWAYS filter out configurations with the EXACT same count and data of the selectors to the
   ones provided by client. The specific count and data of the selectors is regarded as `selector combination`.
2. In order to accomplish the selector rule, the correct count and data of selectors should be provided while publish a
   new configuration.

`Bahavior of Optional Selectors:`

Optional Selectors is used to support additional and dynamic filtering.

The matching rules are identical to standard Selector with extra requirements:

1. Optional Selectors should be used together with Selectors.
2. There MUST be a configuration record WITHOUT optional selectors as default configuration.

##### Client requirement

1. There should be only one configuration instance for the combination of [group, key]. The other fields should not be
   used for configuration identification on client side.

### Features

##### API features

* [ ] Get configuration via [group, key]
* [ ] Poll configurations via [group, key, version] for dynamic reloading

##### Advanced features

* [ ] Configuration management for history restoring, beta application
* [ ] Isolations for environments, areas, purposes
* [ ] Configuration authorization
* [ ] Local fallback storage
* [ ] General and simple protocols for multiple programming languages
* [ ] Configuration encryption
* [ ] Low resource cost and high throughput
* [ ] Extension APIs for customization: local file storage provider, customization storage provider
* [ ] Separate APIs for retrieving and writing operations
* [ ] Statistics of clients including configure using, client info, client address
* [ ] Server: http support
* [ ] Server: https support
* [ ] Server: auth support
* [ ] Configuration reference for between different selector combinations

##### Advanced client features

* [ ] Go: on change event callback

### References

##### A. Ways of beta / blue-green / canary / etc.

1. Select part of the existing instances as candidates - pros. real time effective
2. Create new instances as candidates - pros. fresh new instance

### Dependencies

##### Basic

* [github.com/fxamacker/cbor/v2](github.com/fxamacker/cbor/v2)
* [github.com/go-chi/chi/v5](github.com/go-chi/chi/v5)


