# git-encrypt-agent

[git-encrypt-kms](https://github.com/ikeisuke/git-encrypt-kms)から利用するキーストア用エージェントです。

## Installation

1. [ここ](https://github.com/ikeisuke/git-encrypt-agent/releases)から対応したバイナリをダウンロード
1. 解凍してパスの通ったところに設置

## Features

### start

agentの起動

```bash
$ git-encrypt-agent start
```

#### stop

agentの停止
- 停止すると保存中のキーは全て消えます

```bash
$ git-encrypt-agent stop
```

#### add key

キーの追加
- 32byte以外の場合はエラーになります

```bash
$ echo -n "32bytelengthkeyforencryptiondata" | git-encrypt-agent add --name "keyname"
OK
```

#### get key hash Value

キーのmd5値の取得
- md5値を返却します

```bash
$ git-encryprt-agent get --name "keyname"
[output encrtption key that hashes md5]
```

#### encryption data

キーを使って標準入力からのデータを暗号化します
- 対象のキーをaddコマンドを利用して追加している必要があります

```bash
$ cat plaintext.txt | git-encrypt-agent encrypt --name "savedkeyname"
[output encrypted data encoded by base64]
```

#### decryption data

キーを使って標準入力からのデータを複合化します
- 対象のキーをaddコマンドを利用して追加している必要があります
- encryptコマンドで返却されたbase64済みのデータをわたしてください。

```bash
$ cat encrypted_base64.txt | git-encrypt-agent decrypt --name "savedkeyname"
[output plaintext data]
```
