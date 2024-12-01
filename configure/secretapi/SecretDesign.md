# Secret

## Note List

* Disable swap to avoid sensitive memory to be swapped into hard disk, causing key leaking issue.

## Supported features

* [x] Unseal with root key
* [x] First time initialization and unseal initialization
* Multiple levels: root key, level1 key, level2 key
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
