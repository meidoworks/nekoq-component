# Configuration

### 0. Get Started

TBD

### 1. Features

##### API features

* [x] Get configuration via [group, key]
* [x] Poll configurations via [group, key, version] for dynamic reloading

##### Advanced features

* [x] Common: Isolations for environments, areas, purposes
    * via Selectors
* [X] Common: General and simple protocols for multiple programming languages
    * http protocol for communication
    * cbor protocol for data marshalling
* [x] Server: http support
* [x] Server: sample cfgserver built on postgresql
* [x] Client: Support *struct as dynamic configure container by ClientAdv
    * Thread-safe while reading and writing the configure container
* [x] Performance: Low resource cost and high throughput
* [ ] Configuration management for history restoring, beta application
* [ ] Configuration authorization
* [ ] Local fallback storage
* [ ] Configuration alternatives - env, parameter, file
* [ ] Configuration encryption
* [ ] Extension APIs for customization: local file storage provider, customization storage provider
* [ ] Separate APIs for retrieving and writing operations
* [ ] Statistics of clients including configure using, client info, client address
* [ ] Server: https support
* [ ] Server: auth support
* [ ] Configuration reference for between different selector combinations
    * In order to support flexible configuration access and sharing
    * Support in management portal rather than client and server, meaning that no special changes to the protocol.
* [ ] Crypto alg for auth and encryption: rsa2048, ecdsa256, rsa4096, ecdsa384, ecdsa521
* [ ] Nested configure server architecture for scalable capacity
* [ ] Server: Data lazy loading to reduce memory usage

##### Advanced client features

* [ ] Go: on change event callback
* [ ] Go: retrieve full dump configurations periodically
* [x] Allow retrieving configurations from multiple selectors via different client instance options
    * Best practise: reduce the number of clients in this scenario to reduce the workload of the server.

##### Corner Case Tolerant

* [ ] Inconsistent configure versions(especially while a new update spreads) in the single cluster. Client configure
  fetching will keep flip-flop configure versions connecting to different state servers in a short period. It will cause
  unstable configuration.
    * History version based configure fetching can avoid the issue.

### 2. Concepts

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
* Timestamp - (unix timestamp in seconds)the effective time(create/update) of this configuration

Valid characters of 'group', 'key' fields:

* Alphabets - [A-Za-z]
* Numbers - [0-9]
* Underscore - [_]
* Dot - [.]

##### Selector

Selector is the core functionality for filtering out desired configurations.

Selector consists of `Key` and `Value` fields the same to an entry of a map. Both fields MUST NOT be empty.

Valid characters of the selector fields are:

* Alphabets - [A-Za-z]
* Numbers - [0-9]
* Underscore - [_]
* Dot - [.]

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
2. **Design for failing fast**. Since configurations are extremely important for correctness, any situation that the
   client could not get desired configuration will lead failure on server and client. This includes configure not
   existing or selector matching rule.

### 3. Design

#### 3.1 Server API

##### 3.1.1 Post /retrieving => Retrieve and listen configurations

* Request Headers:

```text
Request Id header(Optional):
X-Request-Id

General http proxy headers(ordered):
True-Client-IP
X-Real-IP
X-Forwarded-For

MIME header:
Accept = application/cbor
```

* Request body: cbor encoded request

* Response status

```text
200 = success
304 = no update configuration within timeout
400 = bad information in header or/and body
404 = one or more configuration keys are not found
406 = accept header invalid
500 = internal error while processing request
```

* Response headers:

```text
MIME header:
Content-Type = application/cbor
```

* Response body:
    * 200 = cbor encoded response
    * 304 = (empty)
    * 400 = (optional)cbor encoded error info
    * 404 = (optional)existence of each configuration requested. information only rather than structural data. should
      not be used.
    * 406 = (empty)
    * 500 = (optional)cbor encoded error info
    * undefined responses beyond the above scenarios even with status codes = 400,404,500

##### 3.1.2 Get /configure/{group}/{key} => Get specific configuration

* Request Headers:

```text
Request Id header(Optional):
X-Request-Id

General http proxy headers(ordered):
True-Client-IP
X-Real-IP
X-Forwarded-For

MIME header:
Accept = application/cbor

Configure server required headers:
X-Configuration-Sel = (selectors data)
X-Configuration-Opt-Sel = (optional selectors data)
```

* Response status:

```text
200 = success
400 = bad information in header or/and body
404 = configuration key are not found
406 = accept header invalid
500 = internal error while processing request
```

* Response headers:

```text
MIME header:
Content-Type = application/cbor
```

* Response body:
    * 406 = (empty)
    * otherwise: cbor encoded response
    * undefined responses beyond the known scenarios even with status codes = 400,404,500

### A. References

##### A.1 Ways of beta / blue-green / canary / etc.

1. Select part of the existing instances as candidates - pros. real time effective
2. Create new instances as candidates - pros. fresh new instance

##### A.2 Reason to choose cbor as first supported encoding

1. Schemaless: flexible and avoid dependency issue
2. Speed: depending on protocol and implementation
3. Compatibility: support various languages and platforms
4. Security: safe marshalling
5. Support: community support
6. Standard: protocol standard

### B. Dependencies

##### Basic

* [github.com/fxamacker/cbor/v2](github.com/fxamacker/cbor/v2)
* [github.com/go-chi/chi/v5](github.com/go-chi/chi/v5)

