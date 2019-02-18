# access_key, secretes key, and endpoint
# put three values to object-secrets.yaml 
# get access_key: kubectl -n rook-ceph get secret rook-ceph-object-user-my-store-my-user -o yaml | grep AccessKey | awk '{print $2}' | base64 --decode
# get secrets_key: kubectl -n rook-ceph get secret rook-ceph-object-user-my-store-my-user -o yaml | grep SecretKey | awk '{print $2}' | base64 --decode
# where endpoint is fixed now with name rook-ceph-rgw-my-store and in rook-ceph namespace

echo -n 'CXPS27F2AWRVHJGDVGXA' | base64
echo -n 'FCBqA35AGMsG5bPWdD1mA6Nfl8yV82iSeA6K4Ca7' | base64
echo -n 'rook-ceph-rgw-my-store.rook-ceph' | base64

