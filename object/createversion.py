import boto3
import boto.s3.connection

access_key = 'USB1A7W8W8RXZ7AWODB1'
secret_key = 'K8HpKFRy2KGLxr0ekH06EiWdnvzNBMqUHS3XHJv8'
#conn = boto3.connect_s3(
#s3client = boto3.client(service_name='s3',
#        aws_access_key_id = access_key,
#        aws_secret_access_key = secret_key,
#        #host = '172.18.136.213/bucketvv', port = 80,
#        endpoint_url = "http://172.18.136.213",
#        )

s3 = boto3.resource('s3',
         endpoint_url="http://172.18.136.213",
         aws_access_key_id=access_key,
         aws_secret_access_key=secret_key)

#print 'into data1'
# list all bucket

# create a new bucket
bucket_name="newbucket"

# set up versioning
bkt_versioning = s3.BucketVersioning(bucket_name)
bkt_versioning.enable()
print"versioning result %s"%(bkt_versioning.status)

bucket = s3.Bucket(bucket_name)
bucket.upload_file(Filename="./anaconda-post.log", Key="mynewdata.log")
bucket.upload_file(Filename="./curlcreatebuckt.sh", Key="mynewdata.log")
print "show version id %s"%s3.Object(bucket_name, "mynewdata.log").version_id


resp=s3.list_object_versions(Bucket="bucket_name")
print "show all versions %s"%resp
)
