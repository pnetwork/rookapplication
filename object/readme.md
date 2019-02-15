

### rook-ceph deployment

/root/rookio/rook/cluster/examples/kubernetes/ceph
kubectl apply -f object.yaml

## must need to check this result

```
kubectl -n rook-ceph get pod -l app=rook-ceph-rgw
NAME                                      READY     STATUS    RESTARTS   AGE
rook-ceph-rgw-my-store-564d7f77d4-pwkp4   1/1       Running   0          59m
```

kubectl apply -f object-user.yaml

kubectl -n rook-ceph get secret rook-ceph-object-user-my-store-my-user -o yaml | grep AccessKey | awk '{print $2}' | base64 --decode
USB1A7W8W8RXZ7AWODB1

kubectl -n rook-ceph get secret rook-ceph-object-user-my-store-my-user -oey | awk '{print $2}' | base64 --decode
K8HpKFRy2KGLxr0ekH06EiWdnvzNBMqUHS3XHJv8


### fast testing

yum install s3cmd
and testing command

s3cmd configuration

```
s3cmd --configuration
New settings:
  Access Key: USB1A7W8W8RXZ7AWODB1
  Secret Key: K8HpKFRy2KGLxr0ekH06EiWdnvzNBMqUHS3XHJv8
  Default Region: US
  S3 Endpoint: 172.18.136.213
  DNS-style bucket+hostname:port template for accessing a bucket: 172.18.136.213/%(bucket)
  Encryption password:
  Path to GPG program: /usr/bin/gpg
  Use HTTPS protocol: False
  HTTP Proxy server name:
  HTTP Proxy server port: 0
```

```
s3cmd la
s3cmd mb s3://bucket2 # create bucket
echo "ahha1" > /tmp/rookObj2  #create file
s3cmd  put /tmp/rookObj2 s3://bucket2  # put data to bucket
```



### Application

for simply run
you can create version by the following script for check versioning, excuting it multple times for generating versioning
plz use boto3 to test versioning.

createversion.py 

show the versioning
listversion.py


