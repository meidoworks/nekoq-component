# Secret

## Note List

* Disable swap to avoid sensitive memory to be swapped into hard disk, causing key leaking issue.
* For issuing certificate purpose, Secret won't provide all possible suites as supported in other tools such as openssl,
  while it focuses on providing common used and secure suites to simply the use as CA.

## Supported features

* [x] Unseal with root key
* [x] First time initialization and unseal initialization
* [x] Multiple levels: root key, level1 key, level2 key
    * Root Key: Cannot be accessed by secret and is provided by external provider, used for unseal secret and encrypting
      level1 keys
        * Custom Implementation1: shamir keys
        * Custom Implementation2: local file (not secure)
        * Custom Implementation3: HSM or external KMS
    * Level1 Key: Cannot be transferred outside secret, used for encrypting level2 keys
        * Note: Level1 keys will not be stored in plaintext in memory. It will be decrypted by root key everytime.
    * Level2 Key: Used for encryption or transferring to external to use
        * Note: Level2 keys can be stored in plaintext in memory after first time loaded
* Key rotate of all levels of keys
* Key re-encrypted: update current key with new upper level key
* Offline cleanup unused keys
* command line tool to initialize local root key
* Support integrated encryption and raw encryption
    * Integrated encryption: encrypted data contains key information maintained by secret
    * Raw encryption: encrypted data is raw output of ciphers
* Key permission management
* Support cipher operations
    * Encryption / Decryption
    * Sign / Verify
* [x] Addon => Jwt token
    * Algorithms: HS256/HS384/HS512, RS256/RS384/RS512, ES256/ES384/ES512, PS256/PS384/RS512
    * 64B key - HS256
    * 128B key - HS384/HS512
    * RSA1024/2048/3072/4096 - RS256/RS384/RS512
    * RSA1024/2048/3072/4096 - PS256/RS384
    * RSA2048/3072/4096 - PS512
    * ECDSA256 - ES256
    * ECDSA384 - ES384
    * ECDSA521 - ES512
* [x] Permission Validation
    * With Jwt
* Support importing external keys and using for external integration
* [x] Cert management
    * Cert tools to create new certs
    * Cert storage to securely store certs
* OpenPGP support
* Cert format support
    * SSL/TLS with client side cert supporting and 3rd party software including go http server, nginx, apache, etc.
    * SSH
    * Email
* Custom cert signature algorithm
* Support encrypted unseal local key file for security
* Support expire time in related use cases
* Support post-quantum cryptography
* Crypto libraries
    * https://github.com/golang/crypto
    * https://github.com/cloudflare/circl
