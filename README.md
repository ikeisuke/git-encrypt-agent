# git-encrypt-agent

git-encrypt-agent stores encryption keys for [git-encrypt-kms](https://github.com/ikeisuke/git-encrypt-kms).

## Installation

1. Download from https://github.com/ikeisuke/git-encrypt-agent/releases .
1. Unzip.
1. Move to PATH environment directory.
1. Set TMPDIR enviromnent on Linux.

## Features

### socket server

#### start

```bash
$ git-encrypt-agent start
```

#### stop

```bash
$ git-encrypt-agent stop
```

#### add key

```bash
$ echo -n "32bytelengthkeyforencryptiondata" | git-encrypt-agent add --name "keyname"
OK
```

Failed to encrypt/decrypt data, when you set not 32byte key.

#### get key hash Value

```bash
$ git-encryprt-agent get --name 'keyname'
[output encrtption key that hashes md5]

#### encryption data

Required after add key command

```bash
$ cat plaintext.txt | git-encrypt-agent encrypt --name "savedkeyname"
[output encrypted hdata encoded by base64 ]
```

#### decryption data
```bash
$ cat encrypted_base64.txt | git-encrypt-agent decrypt --name "savedkeyname"
[output plaintext data]
```
