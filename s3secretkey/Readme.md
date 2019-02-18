object-secrets.yaml  object.yaml  sec-gen.sh  testtool.yaml  testtool.yaml-st
1. Generate key and endpoint: sec-gen.sh
2. inject secretes key to pod: object-secretes.yaml
3. check testtool.yaml's  secretName: rook-ceph-object-secrets is the same with object-secretes.yaml defined
4. after launch testtool.yaml, check the directory /etc/secret-volume with access_key, secrets_key, and endpoint
